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
	"time"
)

// appBBs contains BB implementations
type appBBs struct {
	app            *Application
	EngApptAdapter interfaces.Appointments
}

// GetExample gets an Example by ID
func (a appBBs) GetExample(orgID string, appID string, id string) (*model.Example, error) {
	return a.app.shared.getExample(orgID, appID, id)
}

func (a appBBs) GetAppointmentUnits(providerid int, uin string, accesstoken string) (*[]model.AppointmentUnit, error) {
	conf, _ := a.app.GetEnvConfigs()
	retData, err := a.EngApptAdapter.GetUnits(uin, accesstoken, providerid, conf)
	if err != nil {
		return nil, err
	}
	return retData, nil
}

func (a appBBs) GetPeople(uin string, unitid int, providerid int, accesstoken string) (*[]model.AppointmentPerson, error) {
	conf, _ := a.app.GetEnvConfigs()
	retData, err := a.EngApptAdapter.GetPeople(uin, unitid, providerid, accesstoken, conf)
	if err != nil {
		return nil, err
	}
	return retData, nil

}

func (a appBBs) GetAppointmentOptions(uin string, unitid int, peopleid int, providerid int, startdate time.Time, enddate time.Time, accesstoken string) (*model.AppointmentOptions, error) {
	conf, _ := a.app.GetEnvConfigs()
	retData, err := a.EngApptAdapter.GetTimeSlots(uin, unitid, peopleid, providerid, startdate, enddate, accesstoken, conf)
	if err != nil {
		return nil, err
	}
	return retData, nil
}

func (a appBBs) CreateAppointment(appt *model.AppointmentPost, accessToken string) (string, error) {
	conf, _ := a.app.GetEnvConfigs()
	ret, err := a.EngApptAdapter.CreateAppointment(appt, accessToken, conf)
	if err != nil {
		return "", err
	}
	return ret, nil
}

// newAppBBs creates new appBBs
func newAppBBs(app *Application) appBBs {
	appBB := appBBs{app: app}
	appBB.EngApptAdapter = uiucadapters.NewEngineeringAppontmentsAdapter()
	return appBB
}
