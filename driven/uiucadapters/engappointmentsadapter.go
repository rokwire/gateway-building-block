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
	"strconv"
	"strings"
	"time"
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

	baseURL := conf.EngAppointmentBaseURL
	finalURL := baseURL + "users/" + uin + "/calendars"
	var headers = make(map[string]string)
	headers["Authorization"] = "Bearer " + accessToken

	vendorData, err := lv.getVendorData(finalURL, "GET", headers, nil)
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
		au := model.AppointmentUnit{ID: calendar.ID, ProviderID: providerid, Name: calendar.Name, Location: "", HoursOfOperation: "", Details: "", NextAvailable: "", ImageURL: ""}
		s = append(s, au)
	}

	return &s, nil
}

// GetPeople returns a list of people with appointment calendars from engineering
func (lv EngineeringAppointmentsAdapter) GetPeople(uin string, unitID int, providerid int, accesstoken string, conf *model.EnvConfigData) (*[]model.AppointmentPerson, error) {
	baseURL := conf.EngAppointmentBaseURL
	finalURL := baseURL + "users/" + uin + "/calendars/" + strconv.FormatInt(int64(unitID), 10) + "/advisors"
	var headers = make(map[string]string)
	headers["Authorization"] = "Bearer " + accesstoken

	vendorData, err := lv.getVendorData(finalURL, "GET", headers, nil)
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
		p := model.AppointmentPerson{ID: advisor.ID, ProviderID: providerid, UnitID: unitID, Notes: advisor.Message, Name: advisor.Name, NextAvailable: advisor.NextAvailableDate, ImageURL: ""}
		s = append(s, p)
	}

	return &s, nil
}

// GetTimeSlots returns an object consisting of the time slots and questions for a given personid between startdate and enddate
func (lv EngineeringAppointmentsAdapter) GetTimeSlots(uin string, unitID int, advisorid int, providerid int, startdate time.Time, enddate time.Time, accesstoken string, conf *model.EnvConfigData) (*model.AppointmentOptions, error) {
	baseURL := conf.EngAppointmentBaseURL
	//baseURL := "https://myengr.test.engr.illinois.edu/advisingws/api/"
	finalURL := baseURL + "users/" + uin + "/calendars/" + strconv.FormatInt(int64(unitID), 10) + "/advisors/" + strconv.FormatInt(int64(advisorid), 10) + "/appointments"
	var headers = make(map[string]string)
	headers["Authorization"] = "Bearer " + accesstoken

	vendorData, err := lv.getVendorData(finalURL, "GET", headers, nil)
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
	if !startdate.IsZero() && !enddate.IsZero() {

		const timeLayout = "2006-01-02T15:04:00"
		for i := 0; i < len(options.TimeSlots); i++ {
			timeslot := options.TimeSlots[i]
			slotStartDate, _ := time.Parse(timeLayout, timeslot.StartDate)
			slotStartDateOnly, _, _ := strings.Cut(timeslot.StartDate, "T")
			slotEndDateOnly, _, _ := strings.Cut(timeslot.EndDate, "T")

			slotStartDatepart, _ := time.Parse(time.DateOnly, slotStartDateOnly)
			slotEndDatePart, _ := time.Parse(time.DateOnly, slotEndDateOnly)

			if (slotStartDatepart.Equal(startdate) || slotEndDatePart.Equal(enddate)) || (slotStartDate.After(startdate) && slotStartDate.Before(enddate)) {
				t := model.TimeSlot{ID: timeslot.ID, EndTime: timeslot.EndDate, StartTime: timeslot.StartDate, UnitID: unitID, ProviderID: providerid, PersonID: advisorid, Capacity: 1, Filled: 0}
				ts = append(ts, t)
			}
		}
	}

	for i := 0; i < len(options.Questions); i++ {
		question := options.Questions[i]
		q := model.Question{ID: question.ID, ProviderID: providerid, Required: true, Type: question.Type, SelectValues: question.SelectionValues, Question: question.Title}
		qu = append(qu, q)
	}

	returnData := model.AppointmentOptions{Questions: qu, TimeSlots: ts}
	return &returnData, nil
}

// CreateAppointment creates an appointment in the engieering system.
func (lv EngineeringAppointmentsAdapter) CreateAppointment(appt *model.AppointmentPost, accesstoken string, conf *model.EnvConfigData) (*model.BuildingBlockAppointment, error) {
	baseURL := conf.EngAppointmentBaseURL
	finalURL := baseURL + "Appointment"
	var headers = make(map[string]string)
	headers["Authorization"] = "Bearer " + accesstoken
	headers["Content-Type"] = "application/json"

	slotid, err := strconv.Atoi(appt.SlotID)
	if err != nil {
		return nil, err
	}
	eap := engAppointmentPost{UIN: appt.UserExternalIDs.UIN, SlotID: slotid}
	eap.Answers = make([]engAppointmentAnswerPost, 0)

	for i := 0; i < len(appt.Answers); i++ {
		apptAnswer := appt.Answers[i]
		for j := 0; j < len(apptAnswer.Values); j++ {
			finalAnswer := apptAnswer.Values[j]
			switch finalAnswer {
			case "true":
				finalAnswer = "X"
			case "false":
				finalAnswer = ""
			}
			engAns := engAppointmentAnswerPost{QuestionID: apptAnswer.QuestionID, Value: finalAnswer, UploadID: 0}
			eap.Answers = append(eap.Answers, engAns)
		}

	}
	postData, err := json.Marshal(eap)
	if err != nil {
		return nil, err
	}
	payload := strings.NewReader(string(postData))
	_, err = lv.getVendorData(finalURL, "POST", headers, payload)
	if err != nil {
		return nil, err
	}

	retData := model.BuildingBlockAppointment{ProviderID: appt.ProviderID, UnitID: appt.UnitID, PersonID: appt.PersonID, UserExternalIDs: appt.UserExternalIDs, Type: appt.Type, StartTime: appt.StartTime, EndTime: appt.EndTime, SourceID: appt.SlotID}

	return &retData, nil
}

// UpdateAppointment updates an appointment in the engieering system.
func (lv EngineeringAppointmentsAdapter) UpdateAppointment(appt *model.AppointmentPost, accesstoken string, conf *model.EnvConfigData) (*model.BuildingBlockAppointment, error) {

	slotid, err := strconv.Atoi(appt.SlotID)
	if err != nil {
		return nil, err
	}

	sourceid, err := strconv.Atoi(appt.SourceID)
	if err != nil {
		return nil, err
	}

	if sourceid != slotid {
		_, err = lv.DeleteAppointment(appt.UserExternalIDs.UIN, appt.SourceID, accesstoken, conf)
		if err != nil {
			return nil, err
		}
	}

	ret, err := lv.CreateAppointment(appt, accesstoken, conf)
	if err != nil {
		return nil, err
	}

	return ret, nil
}

// DeleteAppointment cancels an appointment in the engineering appointment system
func (lv EngineeringAppointmentsAdapter) DeleteAppointment(uin string, sourceid string, accesstoken string, conf *model.EnvConfigData) (string, error) {
	baseURL := conf.EngAppointmentBaseURL
	finalURL := baseURL + "Appointment/" + sourceid

	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	_ = writer.WriteField("uin", uin)
	err := writer.Close()
	if err != nil {
		return "", err
	}
	var headers = make(map[string]string)
	headers["Authorization"] = "Bearer " + accesstoken
	headers["Content-Type"] = writer.FormDataContentType()

	vendorData, err := lv.getVendorData(finalURL, "DELETE", headers, strings.NewReader(payload.String()))
	if err != nil {
		return "", err
	}
	return string(vendorData), nil
}

func (lv EngineeringAppointmentsAdapter) getVendorData(targetURL string, method string, headers map[string]string, postdata *strings.Reader) ([]byte, error) {

	client := &http.Client{}

	var postbody = io.Reader(nil)
	if postdata != nil {
		postbody = postdata
	}

	req, err := http.NewRequest(method, targetURL, postbody)

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

	if res.StatusCode == 200 || res.StatusCode == 201 || res.StatusCode == 204 {

		return body, nil
	}

	return nil, errors.New("error making request: " + fmt.Sprint(res.StatusCode) + ": " + string(body))
}

type engAppointmentPost struct {
	UIN     string                     `json:"uin" bson:"uin"`
	SlotID  int                        `json:"slotId" bson:"slotId"`
	Answers []engAppointmentAnswerPost `json:"answers" bson:"answers"`
}

type engAppointmentAnswerPost struct {
	QuestionID string `json:"questionId" bson:"questionId"`
	Value      string `json:"value" bson:"value"`
	UploadID   int    `json:"uploadId" bson:"uploadId"`
}
