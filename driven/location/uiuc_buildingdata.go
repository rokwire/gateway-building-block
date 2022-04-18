package buildinglocation

import (
	wayfinding "apigateway/core/model/Wayfinding"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

type entrance struct {
	UUID         string  `json:"uuid"`
	Name         string  `json:"descriptive_name"`
	ADACompliant bool    `json:"is_ada_compliant"`
	Available    bool    `json:"is_available_for_use"`
	ImageURL     string  `json:"image"`
	Latitude     float32 `json:"latitude"`
	Longitude    float32 `json:"longitude"`
}

type building struct {
	UUID        string     `json:"uuid"`
	Name        string     `json:"name"`
	Number      string     `json:"number"`
	FullAddress string     `json:"location"`
	Address1    string     `json:"address_1"`
	Address2    string     `json:"address_2"`
	City        string     `json:"city"`
	State       string     `json:"state"`
	ZipCode     string     `json:"zipcode"`
	ImageURL    string     `json:"image"`
	MailCode    string     `json:"mailcode"`
	Entrances   []entrance `json:"entrances"`
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
	Response  serverResponse `json:"response"`
	Buildings []building     `json:"results"`
}

type UIUCWayFinding struct {
	APIKey string
	APIUrl string
}

func NewUIUCWayFinding(apikey string, apiurl string) *UIUCWayFinding {
	return &UIUCWayFinding{APIKey: apikey, APIUrl: apiurl}
}

func (uwf *UIUCWayFinding) GetEntrances(bldgID string, activeonly bool, adaonly bool) ([]wayfinding.Entrance, error) {
	ents := make([]wayfinding.Entrance, 0)
	return ents, nil
}

func (uwf *UIUCWayFinding) ClosestEntrance(bldgID string, activeonly bool, adaonly bool, lat float64, long float64) (*wayfinding.Entrance, error) {
	ent := wayfinding.Entrance{}
	return &ent, nil
}

func (uwf *UIUCWayFinding) GetBuilding(bldgID string) (*wayfinding.Building, error) {
	bldg := wayfinding.Building{}
	return &bldg, nil
}

func (uwf *UIUCWayFinding) getBuildingData(bldgID string, latitutde float64, longitude float64) (*building, error) {
	//anged=true&point={"latitude":{{REF_RANGE_LATITUDE}},"longitude":{{REF_RANGE_LONGITUDE}}}
	url := uwf.APIUrl + "/buildings/number/" + bldgID + "?v=2&ranged=true&point={"
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		log.Print(err.Error())
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+uwf.APIKey)

	resp, err := client.Do(req)
	if err != nil {
		log.Print(err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	if resp.Status == "200 OK" {

	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	data := serverLocationData{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, err
	}

	bldg := data.Buildings[0]
	return &bldg, nil

}
