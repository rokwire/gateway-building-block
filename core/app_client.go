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

package core

import (
	"application/core/interfaces"
	"application/core/model"
	"application/driven/uiucadapters"
	"encoding/json"
	"os"
)

// appClient contains client implementations
type appClient struct {
	app                *Application
	Courseadapter      interfaces.Courses
	LocationAdapter    interfaces.WayFinding
	LaundryAdapter     interfaces.LaundryService
	ContactAdapter     interfaces.Contact
	SuccessTeamAdapter interfaces.SuccessTeam
}

// GetExample gets an Example by ID
func (a appClient) GetExample(orgID string, appID string, id string) (*model.Example, error) {
	return a.app.shared.getExample(orgID, appID, id)
}

func (a appClient) GetUnitCalendars(id string) (*[]model.UnitCalendar, error) {
	return a.app.storage.FindCalendars(id)
}

func (a appClient) ListLaundryRooms() (*model.Organization, error) {
	conf, _ := a.app.GetEnvConfigs()
	retData, err := a.LaundryAdapter.ListRooms(conf)
	if err != nil {
		return nil, err
	}
	return retData, nil
}

func (a appClient) GetLaundryRoom(roomid string) (*model.RoomDetail, error) {
	conf, _ := a.app.GetEnvConfigs()
	retData, err := a.LaundryAdapter.GetLaundryRoom(roomid, conf)
	if err != nil {
		return nil, err
	}
	return retData, nil
}

func (a appClient) InitServiceRequest(machineid string) (*model.MachineRequestDetail, error) {
	conf, _ := a.app.GetEnvConfigs()
	retData, err := a.LaundryAdapter.InitServiceRequest(machineid, conf)
	if err != nil {
		return nil, err
	}
	return retData, nil

}

func (a appClient) SubmitServiceRequest(machineID string, problemCode string, comments string, firstname string, lastname string, phone string, email string) (*model.ServiceRequestResult, error) {
	conf, _ := a.app.GetEnvConfigs()
	retData, err := a.LaundryAdapter.SubmitServiceRequest(machineID, problemCode, comments, firstname, lastname, phone, email, conf)
	if err != nil {
		return nil, err
	}
	return retData, nil
}

func (a appClient) GetBuilding(bldgID string, adaOnly bool, latitude float64, longitude float64) (*model.Building, error) {
	conf, _ := a.app.GetEnvConfigs()
	retData, err := a.LocationAdapter.GetBuilding(bldgID, adaOnly, latitude, longitude, conf)
	if err != nil {
		return nil, err
	}
	return retData, nil

}

func (a appClient) GetEntrance(bldgID string, adaOnly bool, latitude float64, longitude float64) (*model.Entrance, error) {
	conf, _ := a.app.GetEnvConfigs()
	retData, err := a.LocationAdapter.GetEntrance(bldgID, adaOnly, latitude, longitude, conf)
	if err != nil {
		return nil, err
	}
	return retData, nil

}

func (a appClient) GetBuildings() (*[]model.Building, error) {
	conf, _ := a.app.GetEnvConfigs()
	retData, err := a.LocationAdapter.GetBuildings(conf)
	if err != nil {
		return nil, err
	}
	return retData, nil

}

func (a appClient) GetContactInfo(uin string, accessToken string, mode string) (*model.Person, int, error) {
	conf, _ := a.app.GetEnvConfigs()
	retData, statuscode, err := a.ContactAdapter.GetContactInformation(uin, accessToken, mode, conf)
	if err != nil {
		return nil, statuscode, err
	}
	return retData, statuscode, nil
}

func (a appClient) GetGiesCourses(uin string, accessToken string) (*[]model.GiesCourse, int, error) {
	conf, _ := a.app.GetEnvConfigs()
	retData, statuscode, err := a.Courseadapter.GetGiesCourses(uin, accessToken, conf)
	if err != nil {
		return nil, statuscode, err
	}
	return retData, statuscode, nil

}

func (a appClient) GetStudentCourses(uin string, termid string, accessToken string) (*[]model.Course, int, error) {
	conf, _ := a.app.GetEnvConfigs()
	retData, statuscode, err := a.Courseadapter.GetStudentCourses(uin, termid, accessToken, conf)
	if err != nil {
		return nil, statuscode, err
	}
	return retData, statuscode, nil
}

func (a appClient) GetTermSessions() (*[4]model.TermSession, error) {

	retData, err := a.Courseadapter.GetTermSessions()
	if err != nil {
		return nil, err
	}
	return retData, nil
}

func (a appClient) GetSuccessTeam(uin string, unitid string, accesstoken string) (*model.SuccessTeam, int, error) {
	conf, _ := a.app.GetEnvConfigs()
	retData, status, err := a.SuccessTeamAdapter.GetSuccessTeam(uin, unitid, accesstoken, conf)
	if err != nil {
		return nil, status, err
	}
	return retData, status, nil

}

func (a appClient) GetPrimaryCareProvider(uin string, accesstoken string) (*[]model.SuccessTeamMember, int, error) {
	conf, _ := a.app.GetEnvConfigs()
	retData, status, err := a.SuccessTeamAdapter.GetPrimaryCareProvider(uin, accesstoken, conf)
	if err != nil {
		return nil, status, err
	}
	return retData, status, nil
}

func (a appClient) GetAcademicAdvisors(uin string, unitid string, accesstoken string) (*[]model.SuccessTeamMember, int, error) {
	conf, _ := a.app.GetEnvConfigs()
	retData, status, err := a.SuccessTeamAdapter.GetAcademicAdvisors(uin, unitid, accesstoken, conf)
	if err != nil {
		return nil, status, err
	}
	return retData, status, nil

}

// newAppClient creates new appClient
func newAppClient(app *Application) appClient {

	client := appClient{app: app}
	//read assets
	file, _ := os.ReadFile("./assets/assets.json")
	assets := model.Asset{}
	_ = json.Unmarshal([]byte(file), &assets)
	laundryAssets := make(map[string]model.LaundryDetails)

	for i := 0; i < len(assets.Laundry.Assets); i++ {
		laundryAsset := assets.Laundry.Assets[i]
		laundryAssets[laundryAsset.LocationID] = laundryAsset.Details
	}

	client.ContactAdapter = uiucadapters.NewUIUCContactAdapter()
	client.LaundryAdapter = uiucadapters.NewCSCLaundryAdapter(laundryAssets)
	client.Courseadapter = uiucadapters.NewCourseAdapter()
	client.LocationAdapter = uiucadapters.NewUIUCWayFinding()
	client.SuccessTeamAdapter = uiucadapters.NewSuccessTeamAdapter(app.storage)
	return client
}
