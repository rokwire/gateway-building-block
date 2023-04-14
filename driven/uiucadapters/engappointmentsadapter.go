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
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

// EngineeringAppointmentsAdapter is a college of engineering implementation of the driven/appointments adapter
type EngineeringAppointmentsAdapter struct {
}

// NewEngineeringAppontmentsAdapter returns a vendor specific implementation of the Appointments interface
func NewEngineeringAppontmentsAdapter() EngineeringAppointmentsAdapter {
	return EngineeringAppointmentsAdapter{}

}

// GetUnits returns a list of courses for the given tudent
func (lv EngineeringAppointmentsAdapter) GetUnits(uin string, accessToken string, providerid int, conf *model.EnvConfigData) (*[]model.AppointmentUnit, error) {

	//baseURL := conf.EngAppointmentBaseURL
	baseURL := "https://myengr.test.engr.illinois.edu/advisingws/api/"
	finalURL := baseURL + "users/" + uin + "/calendars"
	var headers = make(map[string]string)
	headers["Authorization"] = "Bearer " + accessToken

	vendorData, err := lv.getVendorData(finalURL, "GET", headers)
	if err != nil {
		return nil, err
	}

	var calendars []uiuc.EngineeringCalendar
	err = json.Unmarshal(vendorData, &calendars)
	if err != nil {
		return nil, err
	}

	s := make([]model.AppointmentUnit, 0)

	for i := 0; i < len(calendars); i++ {
		calendar := calendars[i]
		au := model.AppointmentUnit{ID: calendar.ID, ProviderID: providerid, Name: calendar.Name, Location: "", HoursOfOperation: "", Details: ""}
		s = append(s, au)
	}

	return &s, nil
}

// GetPeople returns a list of people with appointment calendars from engineering
func (lv EngineeringAppointmentsAdapter) GetPeople(uin string, unitId int, providerid int, accesstoken string, conf *model.EnvConfigData) (*[]model.AppointmentPerson, error) {
	//baseURL := conf.EngAppointmentBaseURL
	baseURL := "https://myengr.test.engr.illinois.edu/advisingws/api/"
	finalURL := baseURL + "users/" + uin + "/calendars/" + strconv.FormatInt(int64(unitId), 10) + "/advisors"
	var headers = make(map[string]string)
	headers["Authorization"] = "Bearer " + accesstoken

	vendorData, err := lv.getVendorData(finalURL, "GET", headers)
	if err != nil {
		return nil, err
	}

	var advisors []uiuc.EngineeringAdvisor
	err = json.Unmarshal(vendorData, &advisors)
	if err != nil {
		return nil, err
	}

	s := make([]model.AppointmentPerson, 0)

	for i := 0; i < len(advisors); i++ {
		advisor := advisors[i]
		p := model.AppointmentPerson{ID: advisor.ID, ProviderID: providerid, UnitID: unitId, Notes: advisor.Message, Name: advisor.Name, NextAvailable: advisor.NextAvailableDate}
		s = append(s, p)
	}

	return &s, nil
}

func (lv EngineeringAppointmentsAdapter) GetTimeSlots(uin string, unitid int, advisorid int, providerid int, accesstoken string, conf *model.EnvConfigData) (*model.AppointmentOptions, error) {
	//baseURL := conf.EngAppointmentBaseURL
	baseURL := "https://myengr.test.engr.illinois.edu/advisingws/api/"
	finalURL := baseURL + "users/" + uin + "/calendars/" + strconv.FormatInt(int64(unitid), 10) + "/advisors/" + strconv.FormatInt(int64(advisorid), 10) + "/appointments"
	var headers = make(map[string]string)
	headers["Authorization"] = "Bearer " + accesstoken

	vendorData, err := lv.getVendorData(finalURL, "GET", headers)
	if err != nil {
		return nil, err
	}

	var options uiuc.EngineeringAdvisorAppointments
	err = json.Unmarshal(vendorData, &options)
	if err != nil {
		return nil, err
	}

	ts := make([]model.TimeSlot, 0)
	qu := make([]model.Question, 0)

	for i := 0; i < len(options.TimeSlots); i++ {
		timeslot := options.TimeSlots[i]
		t := model.TimeSlot{ID: timeslot.ID, EndTime: timeslot.EndDate, StartTime: timeslot.StartDate, UnitID: unitid, ProviderID: providerid, PersonID: advisorid, Capacity: 1, Filled: false}
		ts = append(ts, t)
	}

	for i := 0; i < len(options.Questions); i++ {
		question := options.Questions[i]
		q := model.Question{ID: question.ID, ProviderID: providerid, Required: true, Type: question.Type, SelectValues: question.SelectionValues, Question: question.Title}
		qu = append(qu, q)
	}

	returnData := model.AppointmentOptions{Questions: qu, TimeSlots: ts}
	return &returnData, nil
}

func (lv EngineeringAppointmentsAdapter) getVendorData(targetURL string, method string, headers map[string]string) ([]byte, error) {
	client := &http.Client{}
	req, err := http.NewRequest(method, targetURL, nil)

	if err != nil {
		return nil, err
	}

	for key, element := range headers {
		req.Header.Add(key, element)
	}

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
		return nil, errors.New(res.Status + " : " + string(body))
	}
	if res.StatusCode == 401 {
		return nil, errors.New(res.Status + " : " + string(body))
	}

	if res.StatusCode == 403 {
		return nil, errors.New(res.Status + ": " + string(body))
	}

	if res.StatusCode == 406 {
		return nil, errors.New("server returned 406: possible uin claim mismatch")
	}

	if res.StatusCode == 409 {
		return nil, errors.New(res.Status + " : " + string(body))
	}

	if res.StatusCode == 200 || res.StatusCode == 201 {

		return body, nil
	}

	return nil, errors.New("error making request: " + fmt.Sprint(res.StatusCode) + ": " + string(body))
}
