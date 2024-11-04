// Copyright 2022 Board of Trustees of the University of Illinois.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package uiucadapters

import (
	model "application/core/model"
	uiuc "application/core/model/uiuc"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
)

// UIUCWayFinding is a vendor specific structure that implements the BuildingLocation interface
type UIUCWayFinding struct {
	KnownBuildingFeatures *map[string]model.AppBuildingFeature
}

// NewUIUCWayFinding returns a new instance of a UIUCWayFinding struct
func NewUIUCWayFinding(knownfeatures *map[string]model.AppBuildingFeature) *UIUCWayFinding {
	return &UIUCWayFinding{KnownBuildingFeatures: knownfeatures}
}

// GetEntrance returns the active entrance closest to the user's position that meets the ADA Accessibility filter requirement
func (uwf *UIUCWayFinding) GetEntrance(bldgID string, adaAccessibleOnly bool, latitude float64, longitude float64, conf *model.EnvConfigData) (*model.Entrance, error) {
	apiURL := conf.WayFindingURL
	apikey := conf.WayFindingKey

	lat := fmt.Sprintf("%f", latitude)
	long := fmt.Sprintf("%f", longitude)
	url := apiURL + "/ccf"

	parameters := "{\"v\": 2, \"ranged\": true, \"point\": {\"latitude\": " + lat + ", \"longitude\": " + long + "}}"
	bldSelection := "\"number\": \"" + bldgID + "\""
	adaSelection := ""
	if adaAccessibleOnly {
		adaSelection = ",\"entrances\": {\"ada_compliant\": true}"
	}
	query := "{" + bldSelection + adaSelection + "}"

	bldg, err := uwf.getBuildingData(url, apikey, query, parameters, false)
	if err != nil {
		ent := model.Entrance{}
		return &ent, err
	}
	ent := uwf.closestEntrance((*bldg)[0])
	if ent != nil {
		return uiuc.NewEntrance(*ent), nil
	}
	return nil, nil
}

// GetBuildings returns a list of all buildings
func (uwf *UIUCWayFinding) GetBuildings(conf *model.EnvConfigData) (*[]model.Building, error) {
	apiURL := conf.WayFindingURL
	apikey := conf.WayFindingKey

	url := apiURL + "/ccf"
	parameters := "{\"v\": 2}"

	cmpBldgs, err := uwf.getBuildingData(url, apikey, "{}", parameters, true)
	if err != nil {
		return nil, err
	}
	returnList := uiuc.NewBuildingList(cmpBldgs, uwf.KnownBuildingFeatures)
	return returnList, nil
}

// GetBuilding returns the requested building with all of its entrances that meet the ADA accessibility filter
func (uwf *UIUCWayFinding) GetBuilding(bldgID string, adaAccessibleOnly bool, latitude float64, longitude float64, conf *model.EnvConfigData) (*model.Building, error) {
	apiURL := conf.WayFindingURL
	apikey := conf.WayFindingKey

	url := apiURL + "/ccf/"
	lat := fmt.Sprintf("%f", latitude)
	long := fmt.Sprintf("%f", longitude)

	parameters := ""
	if latitude == 0 && longitude == 0 {
		parameters = "{\"v\": 2, \"ranged\": true, \"point\": {\"latitude\": " + lat + ", \"longitude\": " + long + "}}"
	} else {
		parameters = "{\"v\": 2}"
	}

	bldSelection := "\"number\": \"" + bldgID + "\""
	adaSelection := ""
	if adaAccessibleOnly {
		adaSelection = ",\"entrances\": {\"ada_compliant\": true}"
	}
	query := "{" + bldSelection + adaSelection + "}"
	cmpBldg, err := uwf.getBuildingData(url, apikey, query, parameters, false)
	if err != nil {
		bldg := model.Building{}
		return &bldg, err
	}
	return uiuc.NewBuilding((*cmpBldg)[0], uwf.KnownBuildingFeatures), nil
}

// GetFloorPlan returns the requested floor plan
func (uwf *UIUCWayFinding) GetFloorPlan(bldgNum string, floornumber string, markers string, highlites string, conf *model.EnvConfigData) (*model.FloorPlan, error) {
	apiURL := conf.WayFindingURL
	apikey := conf.WayFindingKey
	reqParams := "?"
	url := apiURL + "/floorplans/number/" + bldgNum + "/floor/" + floornumber
	if markers != "" {
		reqParams += "render_markers=" + markers
		if highlites != "" {
			reqParams += "&"
		}
	}

	if highlites != "" {
		reqParams += "render_highlites=" + highlites
	}
	if reqParams != "?" {
		url += reqParams
	}

	uiucfp, err := uwf.getFloorPlanData(url, apikey)
	if err != nil {
		return nil, err
	}
	return uiuc.NewFloorPlan(*uiucfp), nil
}

// the entrance list coming back from a ranged query to the API is sorted closest to farthest from
// the user's coordinates. The first entrance in the list that is active and matches the ADA filter
// will be the one to return
func (uwf *UIUCWayFinding) closestEntrance(bldg uiuc.CampusBuilding) *uiuc.CampusEntrance {
	for _, n := range bldg.Entrances {
		if n.Available {
			return &n
		}
	}
	return nil
}

func (uwf *UIUCWayFinding) getBuildingData(targetURL string, apikey string, queryString string, parameters string, allBuildings bool) (*[]uiuc.CampusBuilding, error) {
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

	req.Header.Add("Authorization", "Bearer "+apikey)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode == 400 {
		return nil, errors.New("bad request to api end point")
	}

	//used to indicate no building found
	if res.StatusCode == 202 {
		return nil, errors.New("building not found")
	}

	data := uiuc.ServerLocationData{}
	err = json.Unmarshal(body, &data)

	if err != nil {
		return nil, err
	}

	campusBldgs := data.Buildings
	return &campusBldgs, nil
}

func (uwf *UIUCWayFinding) getFloorPlanData(targetURL string, apikey string) (*uiuc.CampusFloorPlan, error) {
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, targetURL, nil)

	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", "Bearer "+apikey)
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode == 400 {
		return nil, errors.New("bad request to api end point")
	}

	data := uiuc.CampusFloorPlanResult{}
	err = json.Unmarshal(body, &data)

	if err != nil {
		return nil, err
	}
	if data.Response.Status == "failed" {
		return nil, errors.New("building not found")
	}
	floorplan := data.Result
	return &floorplan, nil
}
