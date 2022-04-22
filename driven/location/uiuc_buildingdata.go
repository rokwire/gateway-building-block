package buildinglocation

import (
	wayfinding "apigateway/core/model/Wayfinding"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
)

type campusEntrance struct {
	UUID         string  `json:"uuid"`
	Name         string  `json:"descriptive_name"`
	ADACompliant bool    `json:"is_ada_compliant"`
	Available    bool    `json:"is_available_for_use"`
	ImageURL     string  `json:"image"`
	Latitude     float32 `json:"latitude"`
	Longitude    float32 `json:"longitude"`
}

type campusBuilding struct {
	UUID        string           `json:"uuid"`
	Name        string           `json:"name"`
	Number      string           `json:"number"`
	FullAddress string           `json:"location"`
	Address1    string           `json:"address_1"`
	Address2    string           `json:"address_2"`
	City        string           `json:"city"`
	State       string           `json:"state"`
	ZipCode     string           `json:"zipcode"`
	ImageURL    string           `json:"image"`
	MailCode    string           `json:"mailcode"`
	Entrances   []campusEntrance `json:"entrances"`
}

type serverResponse struct {
	Status         string `json:"status"`
	HttpStatusCode int    `json:"http_return"`
	CollectionType string `json:"collection"`
	Count          int    `json:"count"`
	ErrorList      string `json:"errors"`
	ErrorMessage   string `json:"error_text"`
}

type serverLocationData struct {
	Response  serverResponse   `json:"response"`
	Buildings []campusBuilding `json:"results"`
}

type UIUCWayFinding struct {
	APIKey string
	APIUrl string
}

func NewUIUCWayFinding(apikey string, apiurl string) *UIUCWayFinding {
	return &UIUCWayFinding{APIKey: apikey, APIUrl: apiurl}
}

func NewBuilding(bldg campusBuilding) *wayfinding.Building {
	newBldg := wayfinding.Building{ID: bldg.UUID, Name: bldg.Name, ImageURL: bldg.ImageURL, Address1: bldg.Address1, Address2: bldg.Address2, FullAddress: bldg.FullAddress, City: bldg.City, ZipCode: bldg.ZipCode, State: bldg.State}
	newBldg.Entrances = make([]wayfinding.Entrance, 0)
	for _, n := range bldg.Entrances {
		if n.Available {
			newBldg.Entrances = append(newBldg.Entrances, *NewEntrance(n))
		}
	}
	return &newBldg
}

func NewEntrance(ent campusEntrance) *wayfinding.Entrance {
	newEnt := wayfinding.Entrance{ID: ent.UUID, Name: ent.Name, ADACompliant: ent.ADACompliant, Available: ent.Available, ImageURL: ent.ImageURL, Latitude: ent.Latitude, Longitude: ent.Longitude}
	return &newEnt
}

//GetEntrance returns the active entrance closest to the user's position that meets the ADA Accessibility filter requirement
func (uwf *UIUCWayFinding) GetEntrance(bldgID string, adaAccessibleOnly bool, latitude float64, longitude float64) (*wayfinding.Entrance, error) {
	url := uwf.APIUrl + "/buildings/number/" + bldgID + "?v=2&ranged=true&point={latitude: " + fmt.Sprintf("%f", latitude) + ", longitude: " + fmt.Sprintf("%f", longitude) + "}"
	bldg, err := uwf.getBuildingData(url)
	if err != nil {
		return nil, err
	}
	ent := uwf.closestEntrance(*bldg, adaAccessibleOnly)
	return NewEntrance(*ent), nil
}

//GetBuilding returns the requested building with all of its entrances that meet the ADA accessibility filter
func (uwf *UIUCWayFinding) GetBuilding(bldgID string, adaAccessibleOnly bool) (*wayfinding.Building, error) {
	url := uwf.APIUrl + "/buildings/number/" + bldgID + "?v=2"
	cmpBldg, err := uwf.getBuildingData(url)
	if err != nil {
		return nil, err
	}
	return NewBuilding(*cmpBldg), nil
}

//the entrance list coming back from a ranged query to the API is sorted closest to farthest from
//the user's coordinates. The first entrance in the list that is active and matches the ADA filter
//will be the one to return
func (uwf *UIUCWayFinding) closestEntrance(bldg campusBuilding, adaOnly bool) *campusEntrance {
	for _, n := range bldg.Entrances {
		if n.Available {
			if adaOnly && n.ADACompliant {
				return &n
			}

			if !adaOnly {
				return &n
			}
		}
	}
	return nil
}

func (uwf *UIUCWayFinding) getBuildingData(url string) (*campusBuilding, error) {
	method := "GET"

	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	err := writer.Close()
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", "Bearer "+uwf.APIKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	data := serverLocationData{}
	err = json.Unmarshal(body, &data)

	if err != nil {
		return nil, err
	}
	return &data.Buildings[0], nil

}
