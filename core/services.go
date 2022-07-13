/*
 *   Copyright (c) 2020 Board of Trustees of the University of Illinois.
 *   All rights reserved.

 *   Licensed under the Apache License, Version 2.0 (the "License");
 *   you may not use this file except in compliance with the License.
 *   You may obtain a copy of the License at

 *   http://www.apache.org/licenses/LICENSE-2.0

 *   Unless required by applicable law or agreed to in writing, software
 *   distributed under the License is distributed on an "AS IS" BASIS,
 *   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *   See the License for the specific language governing permissions and
 *   limitations under the License.
 */

package core

import (
	model "apigateway/core/model"
)

func (app *Application) getVersion() string {
	return app.version
}

func (app *Application) storeRecord(name string) error {
	return app.storage.StoreRecord(name)
}

func (app *Application) listLaundryRooms() (model.Organization, error) {
	lr, _ := app.laundry.ListRooms()
	return *lr, nil
}

func (app *Application) listAppliances(id string) (model.RoomDetail, error) {
	ap, _ := app.laundry.GetLaundryRoom(id)
	return *ap, nil
}

func (app *Application) initServiceRequest(machineid string) (model.MachineRequestDetail, error) {
	sr, _ := app.laundry.InitServiceRequest(machineid)
	return *sr, nil
}

func (app *Application) submitServiceRequest(machineID string, problemCode string, comments string, firstname string, lastname string, phone string, email string) (model.ServiceRequestResult, error) {
	srr, _ := app.laundry.SubmitServiceRequest(machineID, problemCode, comments, firstname, lastname, phone, email)
	return *srr, nil
}

func (app *Application) getBuilding(bldgID string, adaOnly bool) (model.Building, error) {
	bldg, err := app.locationAdapter.GetBuilding(bldgID, adaOnly)
	if err != nil {
		return *bldg, err
	}
	return *bldg, nil
}

func (app *Application) getEntrance(bldgID string, adaOnly bool, latitude float64, longitude float64) (*model.Entrance, error) {
	entrance, err := app.locationAdapter.GetEntrance(bldgID, adaOnly, latitude, longitude)
	if err != nil {
		if entrance == nil {
			return nil, nil
		}
		return entrance, err
	}
	return entrance, nil
}

func (app *Application) getBuildings() (*[]model.Building, error) {
	buildings, err := app.locationAdapter.GetBuildings()
	if err != nil {
		return nil, err
	}
	return buildings, nil
}

func (app *Application) getContactInfo(uin string, accessToken string, mode string) (*model.Person, int, error) {
	person, statusCode, err := app.contactInfoAdapter.GetContactInformation(uin, accessToken, mode)
	if err != nil {
		return nil, statusCode, err
	}
	return person, 200, nil
}

func (app *Application) getGiesCourses(uin string, accessToken string) (*[]model.GiesCourse, int, error) {
	courseList, statusCode, err := app.giesCourseAdapter.GetGiesCourses(uin, accessToken)
	if err != nil {
		return nil, statusCode, err
	}
	return courseList, 200, nil
}
