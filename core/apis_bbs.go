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
	"application/core/model"
	"strconv"
	"time"
)

// appBBs contains BB implementations
type appBBs struct {
	app *Application
}

// GetExample gets an Example by ID
func (a appBBs) GetExample(orgID string, appID string, id string) (*model.Example, error) {
	return a.app.shared.getExample(orgID, appID, id)
}

func (a appBBs) GetAppointmentUnits(providerid int, uin string, accesstoken string) (*[]model.AppointmentUnit, error) {
	conf, _ := a.app.GetEnvConfigs()
	apptAdapter := a.app.AppointmentAdapters[strconv.Itoa(providerid)]
	retData, err := apptAdapter.GetUnits(uin, accesstoken, providerid, conf)
	if err != nil {
		return nil, err
	}
	return retData, nil
}

func (a appBBs) GetPeople(uin string, unitid int, providerid int, accesstoken string) (*[]model.AppointmentPerson, error) {
	conf, _ := a.app.GetEnvConfigs()
	apptAdapter := a.app.AppointmentAdapters[strconv.Itoa(providerid)]
	retData, err := apptAdapter.GetPeople(uin, unitid, providerid, accesstoken, conf)
	if err != nil {
		return nil, err
	}
	return retData, nil

}

func (a appBBs) GetAppointmentOptions(uin string, unitid int, peopleid int, providerid int, startdate time.Time, enddate time.Time, accesstoken string) (*model.AppointmentOptions, error) {
	conf, _ := a.app.GetEnvConfigs()
	apptAdapter := a.app.AppointmentAdapters[strconv.Itoa(providerid)]
	retData, err := apptAdapter.GetTimeSlots(uin, unitid, peopleid, providerid, startdate, enddate, accesstoken, conf)
	if err != nil {
		return nil, err
	}
	return retData, nil
}

func (a appBBs) CreateAppointment(appt *model.AppointmentPost, accessToken string) (*model.BuildingBlockAppointment, error) {
	conf, _ := a.app.GetEnvConfigs()
	apptAdapter := a.app.AppointmentAdapters[appt.ProviderID]
	ret, err := apptAdapter.CreateAppointment(appt, accessToken, conf)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func (a appBBs) UpdateAppointment(appt *model.AppointmentPost, accessToken string) (*model.BuildingBlockAppointment, error) {
	conf, _ := a.app.GetEnvConfigs()
	apptAdapter := a.app.AppointmentAdapters[appt.ProviderID]
	ret, err := apptAdapter.UpdateAppointment(appt, accessToken, conf)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func (a appBBs) DeleteAppointment(uin string, providerid int, sourceid string, accesstoken string) (string, error) {
	conf, _ := a.app.GetEnvConfigs()
	apptAdapter := a.app.AppointmentAdapters[strconv.Itoa(providerid)]
	ret, err := apptAdapter.DeleteAppointment(uin, sourceid, accesstoken, conf)
	if err != nil {
		return "", err
	}
	return ret, nil
}

func (a appBBs) GetLegacyEvents() ([]model.LegacyEvent, error) {

	leEvents, err := a.app.storage.FindAllLegacyEvents()
	if err != nil {
		return nil, err
	}

	blacklist, err := a.app.storage.FindWebtoolsBlacklistData(nil)
	if err != nil {
		return nil, err
	}

	var newLegacyEvents []model.LegacyEvent
	for _, le := range leEvents {

		isBlacklisted := a.isBlacklisted(blacklist, le)
		if !isBlacklisted {
			newLegacyEvents = append(newLegacyEvents, le)
		}
	}

	return newLegacyEvents, nil

}

func (a appBBs) isBlacklisted(blacklists []model.WebToolsItem, event model.LegacyEvent) bool {
	for _, blacklist := range blacklists {
		switch blacklist.Name {
		case "webtools_events_ids":
			for _, id := range blacklist.Data {
				if event.DataSourceEventID == id {
					return true
				}
			}
		case "webtools_calendar_ids":
			for _, id := range blacklist.Data {
				if event.CalendarID == id {
					return true
				}
			}
		case "webtools_originating_calendar_ids":
			for _, id := range blacklist.Data {
				if event.OriginatingCalendarID == id {
					return true
				}
			}
		}
	}
	return false
}

// newAppBBs creates new appBBs
func newAppBBs(app *Application) appBBs {
	appBB := appBBs{app: app}
	return appBB
}
