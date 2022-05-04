package buildinglocation

import (
	wayfinding "apigateway/core/model/wayfinding"
	"bytes"
	"encoding/json"
	"errors"
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
	HTTPStatusCode int    `json:"http_return"`
	CollectionType string `json:"collection"`
	Count          int    `json:"count"`
	ErrorList      string `json:"errors"`
	ErrorMessage   string `json:"error_text"`
}

type serverLocationData struct {
	Response  serverResponse   `json:"response"`
	Buildings []campusBuilding `json:"results"`
}

//UIUCWayFinding is a vendor specific structure that implements the BuildingLocation interface
type UIUCWayFinding struct {
	APIKey string
	APIUrl string
}

//NewUIUCWayFinding returns a new instance of a UIUCWayFinding struct
func NewUIUCWayFinding(apikey string, apiurl string) *UIUCWayFinding {
	return &UIUCWayFinding{APIKey: apikey, APIUrl: apiurl}
}

//NewBuilding creates a wayfinding.Building instance from a campusBuilding,
//including all active entrances for the building
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

//NewEntrance creates a wayfinding.Entrance instance from a campusEntrance object
func NewEntrance(ent campusEntrance) *wayfinding.Entrance {
	newEnt := wayfinding.Entrance{ID: ent.UUID, Name: ent.Name, ADACompliant: ent.ADACompliant, Available: ent.Available, ImageURL: ent.ImageURL, Latitude: ent.Latitude, Longitude: ent.Longitude}
	return &newEnt
}

//GetEntrance returns the active entrance closest to the user's position that meets the ADA Accessibility filter requirement
func (uwf *UIUCWayFinding) GetEntrance(bldgID string, adaAccessibleOnly bool, latitude float64, longitude float64) (*wayfinding.Entrance, error) {
	lat := fmt.Sprintf("%f", latitude)
	long := fmt.Sprintf("%f", longitude)
	url := uwf.APIUrl + "/ccf"

	parameters := "{\"v\": 1, \"ranged\": true, \"point\": {\"latitude\": " + lat + ", \"longitude\": " + long + "}}"
	bldSelection := "\"banner_code\": \"" + bldgID + "\""
	adaSelection := ""
	if adaAccessibleOnly {
		adaSelection = ",\"entrances\": {\"ada_compliant\": true}"
	}
	query := "{" + bldSelection + adaSelection + "}"

	bldg, err := uwf.getBuildingData(url, query, parameters)
	if err != nil {
		ent := wayfinding.Entrance{}
		return &ent, err
	}
	ent := uwf.closestEntrance(*bldg)
	if ent != nil {
		return NewEntrance(*ent), nil
	}
	return nil, nil
}

//GetBuilding returns the requested building with all of its entrances that meet the ADA accessibility filter
func (uwf *UIUCWayFinding) GetBuilding(bldgID string, adaAccessibleOnly bool) (*wayfinding.Building, error) {
	url := uwf.APIUrl + "/ccf"
	parameters := "{\"v\": 1}"
	bldSelection := "\"banner_code\": \"" + bldgID + "\""
	adaSelection := ""
	if adaAccessibleOnly {
		adaSelection = ",\"entrances\": {\"ada_compliant\": true}"
	}
	query := "{" + bldSelection + adaSelection + "}"
	cmpBldg, err := uwf.getBuildingData(url, query, parameters)
	if err != nil {
		bldg := wayfinding.Building{}
		return &bldg, err
	}
	return NewBuilding(*cmpBldg), nil
}

//the entrance list coming back from a ranged query to the API is sorted closest to farthest from
//the user's coordinates. The first entrance in the list that is active and matches the ADA filter
//will be the one to return
func (uwf *UIUCWayFinding) closestEntrance(bldg campusBuilding) *campusEntrance {
	for _, n := range bldg.Entrances {
		if n.Available {
			return &n
		}
	}
	return nil
}

func (uwf *UIUCWayFinding) getBuildingData(targetURL string, queryString string, parameters string) (*campusBuilding, error) {
	method := "POST"

	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	_ = writer.WriteField("collection", "buildings")
	_ = writer.WriteField("action", "fetch")
	_ = writer.WriteField("query", queryString)
	_ = writer.WriteField("parameters", parameters)
	err := writer.Close()
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	fmt.Println(targetURL)
	req, err := http.NewRequest(method, targetURL, payload)

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

	if res.StatusCode == 400 {
		fmt.Println(string(body))
		return nil, errors.New("Bad request to api end point")
	}

	data := serverLocationData{}
	err = json.Unmarshal(body, &data)

	if err != nil {
		return nil, err
	}
	return &data.Buildings[0], nil

}
