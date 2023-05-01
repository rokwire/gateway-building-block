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

package web

import (
	"application/core"
	"application/core/model"
	utils "application/utils"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/rokwire/core-auth-library-go/v3/tokenauth"
	"github.com/rokwire/logging-library-go/v2/logs"
	"github.com/rokwire/logging-library-go/v2/logutils"
)

// ClientAPIsHandler handles the client rest APIs implementation
type ClientAPIsHandler struct {
	app *core.Application
}

func (h ClientAPIsHandler) getExample(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	params := mux.Vars(r)
	id := params["id"]
	if len(id) <= 0 {
		return l.HTTPResponseErrorData(logutils.StatusMissing, logutils.TypePathParam, logutils.StringArgs("id"), nil, http.StatusBadRequest, false)
	}

	example, err := h.app.Client.GetExample(claims.OrgID, claims.AppID, id)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionGet, model.TypeExample, nil, err, http.StatusInternalServerError, true)
	}

	response, err := json.Marshal(example)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionMarshal, logutils.TypeResponseBody, nil, err, http.StatusInternalServerError, false)
	}
	return l.HTTPResponseSuccessJSON(response)
}

// GetBuilding returns an the building matching the provided building id
// @Summary Get the requested building with all of its available entrances filterd by the ADA only flag
// @Tags Client
// @ID Building
// @Accept  json
// @Produce json
// @Param id query string true "Building identifier"
// @Param adaOnly query bool false "ADA entrances filter"
// @Param lat query number false "latitude coordinate of the user"
// @Param long query number false "longitude coordinate of the user"
// @Success 200 {object} model.Building
// @Security RokwireAuth
// @Router /wayfinding/building [get]
func (h ClientAPIsHandler) getBuilding(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {

	bldgid := ""
	adaOnly := false
	reqParams := utils.ConstructFilter(r)
	var latitude, longitude float64
	latitude = 0
	longitude = 0
	for _, v := range reqParams.Items {
		if v.Field == "id" {
			bldgid = v.Value[0]
		}
		if v.Field == "adaOnly" {
			ada, err := strconv.ParseBool(v.Value[0])
			if err != nil {
				return l.HTTPResponseErrorData(logutils.StatusInvalid, logutils.TypeQueryParam, logutils.StringArgs("adaOnly"), nil, http.StatusBadRequest, false)
			}
			adaOnly = ada
		}
		if v.Field == "lat" {
			lat, err := strconv.ParseFloat(v.Value[0], 64)
			if err != nil {
				return l.HTTPResponseErrorData(logutils.StatusInvalid, logutils.TypeQueryParam, logutils.StringArgs("lat"), nil, http.StatusBadRequest, false)
			}
			latitude = lat
		}

		if v.Field == "long" {
			long, err := strconv.ParseFloat(v.Value[0], 64)
			if err != nil {
				return l.HTTPResponseErrorData(logutils.StatusInvalid, logutils.TypeQueryParam, logutils.StringArgs("long"), nil, http.StatusBadRequest, false)
			}
			longitude = long
		}
	}
	if bldgid == "" || bldgid == "nil" {
		return l.HTTPResponseErrorData(logutils.StatusInvalid, logutils.TypeQueryParam, logutils.StringArgs("id"), nil, http.StatusBadRequest, false)
	}

	bldg, err := h.app.Client.GetBuilding(bldgid, adaOnly, latitude, longitude)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionGet, model.TypeBuilding, nil, err, http.StatusInternalServerError, true)
	}

	resAsJSON, err := json.Marshal(bldg)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionMarshal, logutils.TypeResult, nil, err, http.StatusInternalServerError, false)
	}

	return l.HTTPResponseSuccessJSON(resAsJSON)
}

// GetEntrance returns a building entrance record
// @Summary Returns the entrance of the specified building that is closest to the user
// @Tags Client
// @ID Entrance
// @Param id query string true "Building identifier"
// @Param adaOnly query bool false "ADA entrances filter"
// @Param lat query number true "latitude coordinate of the user"
// @Param long query number true "longitude coordinate of the user"
// @Accept  json
// @Success 200 {object} model.Entrance
// @Failure 404 {object} rest.errorMessage
// @Security RokwireAuth
// @Router /wayfinding/entrance [get]
func (h ClientAPIsHandler) getEntrance(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	bldgid := ""
	adaOnly := false
	reqParams := utils.ConstructFilter(r)
	var latitude, longitude float64
	latitude = 0
	longitude = 0
	for _, v := range reqParams.Items {
		if v.Field == "id" {
			bldgid = v.Value[0]
		}
		if v.Field == "adaOnly" {
			ada, err := strconv.ParseBool(v.Value[0])
			if err != nil {
				return l.HTTPResponseErrorData(logutils.StatusInvalid, logutils.TypeQueryParam, logutils.StringArgs("adaOnly"), nil, http.StatusBadRequest, false)
			}
			adaOnly = ada
		}
		if v.Field == "lat" {
			lat, err := strconv.ParseFloat(v.Value[0], 64)
			if err != nil {
				return l.HTTPResponseErrorData(logutils.StatusInvalid, logutils.TypeQueryParam, logutils.StringArgs("lat"), nil, http.StatusBadRequest, false)
			}
			latitude = lat
		}

		if v.Field == "long" {
			long, err := strconv.ParseFloat(v.Value[0], 64)
			if err != nil {
				return l.HTTPResponseErrorData(logutils.StatusInvalid, logutils.TypeQueryParam, logutils.StringArgs("long"), nil, http.StatusBadRequest, false)
			}
			longitude = long
		}
	}
	if bldgid == "" || bldgid == "nil" {
		return l.HTTPResponseErrorData(logutils.StatusInvalid, logutils.TypeQueryParam, logutils.StringArgs("id"), nil, http.StatusBadRequest, false)
	}

	entrance, err := h.app.Client.GetEntrance(bldgid, adaOnly, latitude, longitude)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionGet, model.TypeBuilding, nil, err, http.StatusInternalServerError, true)
	}

	if entrance == nil {
		return l.HTTPResponseErrorAction(logutils.ActionFind, model.TypeBuilding, nil, err, http.StatusNotFound, true)

	}
	resAsJSON, err := json.Marshal(entrance)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionMarshal, logutils.TypeResult, nil, err, http.StatusInternalServerError, false)
	}

	return l.HTTPResponseSuccessJSON(resAsJSON)

}

// GetBuildings returns a list of all buildings
// @Summary Get a list of all buildings with a list of active entrances
// @Tags Client
// @ID BuildingList
// @Accept  json
// @Produce json
// @Success 200 {object} []model.Building
// @Security RokwireAuth
// @Router /wayfinding/buildings [get]
func (h ClientAPIsHandler) getBuildings(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	bldgs, err := h.app.Client.GetBuildings()

	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionGet, model.TypeBuilding, nil, err, http.StatusInternalServerError, true)
	}

	if bldgs == nil {
		return l.HTTPResponseErrorAction(logutils.ActionGet, model.TypeBuilding, nil, err, http.StatusNotFound, true)

	}
	resAsJSON, err := json.Marshal(bldgs)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionMarshal, logutils.TypeResult, nil, err, http.StatusInternalServerError, false)
	}

	return l.HTTPResponseSuccessJSON(resAsJSON)
}

// GetTermSessions returns a list of recent, current and upcoming term sessions
// @Summary Get a list of term sessions centered on the calculated current session
// @Tags Client
// @ID TermSession
// @Accept  json
// @Produce json
// @Success 200 {object} []model.TermSession
// @Security RokwireAuth
// @Router /termsessions/listcurrent [get]
func (h ClientAPIsHandler) getTermSessions(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {

	termSessions, err := h.app.Client.GetTermSessions()
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionGet, model.TypeTermSession, nil, err, http.StatusInternalServerError, true)
	}

	if termSessions == nil {
		return l.HTTPResponseErrorAction(logutils.ActionGet, model.TypeBuilding, nil, err, http.StatusNotFound, true)

	}
	resAsJSON, err := json.Marshal(termSessions)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionMarshal, logutils.TypeResult, nil, err, http.StatusInternalServerError, false)
	}

	return l.HTTPResponseSuccessJSON(resAsJSON)
}

// GetContactInfo returns the contact information of a person
// @Summary Returns the name, permanent and mailing addresses, phone number and emergency contact information for a person
// @Tags Client
// @ID ConatctInfo
// @Param id query string true "User ID"
// @Accept  json
// @Produce json
// @Success 200 {object} model.Person
// @Security RokwireAuth ExternalAuth
// @Router /person/contactinfo [get]
func (h ClientAPIsHandler) getContactInfo(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {

	externalToken := r.Header.Get("External-Authorization")
	if externalToken == "" {
		return l.HTTPResponseErrorData(logutils.StatusMissing, logutils.TypeHeader, logutils.StringArgs("external auth token"), nil, http.StatusBadRequest, false)
	}

	mode := "0"
	uin := ""
	reqParams := utils.ConstructFilter(r)
	if reqParams != nil {
		for _, v := range reqParams.Items {
			switch v.Field {
			case "id":
				uin = v.Value[0]
			case "mode":
				mode = v.Value[0]
			}
		}
	}

	if uin == "" || uin == "null" {
		return l.HTTPResponseErrorData(logutils.StatusMissing, logutils.TypePathParam, logutils.StringArgs("id"), nil, http.StatusBadRequest, false)
	}

	if uin == "123456789" {
		mode = "1"
	}

	person, statusCode, err := h.app.Client.GetContactInfo(uin, externalToken, mode)
	if err != nil {
		return h.setReturnDataOnHTTPError(l, statusCode)
	}

	resAsJSON, err := json.Marshal(person)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionMarshal, logutils.TypeResult, nil, err, http.StatusInternalServerError, false)
	}
	return l.HTTPResponseSuccessJSON(resAsJSON)
}

// GetGiesCourses returns a list of registered courses for GIES students
// @Summary Returns a list of registered courses
// @Tags Client
// @ID GiesCourses
// @Param id query string true "User ID"
// @Accept  json
// @Produce json
// @Success 200 {object} []model.GiesCourse
// @Security RokwireAuth ExternalAuth
// @Router /courses/giescourses [get]
func (h ClientAPIsHandler) getGiesCourses(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {

	externalToken := r.Header.Get("External-Authorization")
	if externalToken == "" {
		return l.HTTPResponseErrorData(logutils.StatusMissing, logutils.TypeHeader, logutils.StringArgs("external auth token"), nil, http.StatusBadRequest, false)
	}

	id := ""
	reqParams := utils.ConstructFilter(r)
	if reqParams != nil {
		for _, v := range reqParams.Items {
			switch v.Field {
			case "id":
				id = v.Value[0]
			}
		}
	}

	if id == "" || id == "null" {
		return l.HTTPResponseErrorData(logutils.StatusInvalid, logutils.TypeQueryParam, logutils.StringArgs("id"), nil, http.StatusBadRequest, false)
	}

	giesCourseList, statusCode, err := h.app.Client.GetGiesCourses(id, externalToken)
	if err != nil {
		return h.setReturnDataOnHTTPError(l, statusCode)
	}

	resAsJSON, err := json.Marshal(giesCourseList)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionMarshal, logutils.TypeResult, nil, err, http.StatusInternalServerError, false)
	}
	return l.HTTPResponseSuccessJSON(resAsJSON)
}

// GetStudentcourses returns a list of registered courses for a student
// @Summary Returns a list of registered courses
// @Tags Client
// @ID Studentcourses
// @Param id query string true "User ID"
// @Param termid query string true "term id"
// @Accept  json
// @Produce json
// @Success 200 {object} []model.Course
// @Security RokwireAuth ExternalAuth
// @Router /courses/studentcourses [get]
func (h ClientAPIsHandler) getStudentCourses(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {

	externalToken := r.Header.Get("External-Authorization")
	if externalToken == "" {
		return l.HTTPResponseErrorData(logutils.StatusMissing, logutils.TypeHeader, logutils.StringArgs("external auth token"), nil, http.StatusBadRequest, false)
	}

	id := ""
	termid := ""
	adaOnly := false
	var latitude, longitude float64
	latitude = 0
	longitude = 0

	reqParams := utils.ConstructFilter(r)
	if reqParams != nil {
		for _, v := range reqParams.Items {
			switch v.Field {
			case "id":
				id = v.Value[0]
			case "termid":
				termid = v.Value[0]
			case "long":
				long, err := strconv.ParseFloat(v.Value[0], 64)
				if err != nil {
					return l.HTTPResponseErrorData(logutils.StatusInvalid, logutils.TypeQueryParam, logutils.StringArgs("long"), nil, http.StatusBadRequest, false)
				}
				longitude = long
			case "lat":
				lat, err := strconv.ParseFloat(v.Value[0], 64)
				if err != nil {
					return l.HTTPResponseErrorData(logutils.StatusInvalid, logutils.TypeQueryParam, logutils.StringArgs("lat"), nil, http.StatusBadRequest, false)
				}
				latitude = lat
			case "adaOnly":
				ada, err := strconv.ParseBool(v.Value[0])
				if err != nil {
					return l.HTTPResponseErrorData(logutils.StatusInvalid, logutils.TypeQueryParam, logutils.StringArgs("adaOnly"), nil, http.StatusBadRequest, false)
				}
				adaOnly = ada
			}
		}
	}

	if id == "" || id == "null" {
		return l.HTTPResponseErrorData(logutils.StatusMissing, logutils.TypeQueryParam, logutils.StringArgs("id"), nil, http.StatusBadRequest, false)
	}

	if termid == "" || termid == "null" {
		return l.HTTPResponseErrorData(logutils.StatusMissing, logutils.TypeQueryParam, logutils.StringArgs("termid"), nil, http.StatusBadRequest, false)
	}

	courseList, statusCode, err := h.app.Client.GetStudentCourses(id, termid, externalToken)
	if err != nil {
		return h.setReturnDataOnHTTPError(l, statusCode)
	}

	//create a map of buildings we need so we don't retrieve the same building multiple times
	neededBuildings := make(map[string]model.Building)
	for index, crntCourse := range *courseList {
		//check if course is not online here then proceed
		if crntCourse.Section.BuildingID != "" {
			bldg, bldgexists := neededBuildings[crntCourse.Section.BuildingID]
			if bldgexists {
				(*courseList)[index].Section.Location = bldg
			} else {
				bldg, err := (h.app.Client.GetBuilding(crntCourse.Section.BuildingID, adaOnly, latitude, longitude))
				if err != nil {
					log.Printf("Error retrieving building details: %s\n", err)
				} else {
					(*courseList)[index].Section.Location = *bldg
					neededBuildings[bldg.ID] = *bldg
				}
			}
		}

	}

	resAsJSON, err := json.Marshal(courseList)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionMarshal, logutils.TypeResult, nil, err, http.StatusInternalServerError, false)
	}
	return l.HTTPResponseSuccessJSON(resAsJSON)
}

// GetLaundryRooms returns an organization record
// @Summary Get list of all campus laundry rooms
// @Tags Client
// @ID Rooms
// @Accept  json
// @Produce json
// @Success 200 {object} model.Organization
// @Security RokwireAuth
// @Router /laundry/rooms [get]
func (h ClientAPIsHandler) getLaundryRooms(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	org, err := h.app.Client.ListLaundryRooms()
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionGet, model.TypeLaundryRooms, nil, err, http.StatusInternalServerError, true)
	}

	if org == nil {
		return l.HTTPResponseErrorAction(logutils.ActionGet, model.TypeLaundryRooms, nil, err, http.StatusNotFound, true)

	}
	resAsJSON, err := json.Marshal(org)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionMarshal, logutils.TypeResult, nil, err, http.StatusInternalServerError, false)
	}

	return l.HTTPResponseSuccessJSON(resAsJSON)
}

// GetRoomDetails returns a laundry room detail record
// @Summary Returns the list of machines and the number of washers and dryers available in a laundry room
// @Tags Client
// @ID Room
// @Param id query int true "Room id"
// @Accept  json
// @Success 200 {object} model.RoomDetail
// @Security RokwireAuth
// @Router /laundry/roomdetail [get]
func (h ClientAPIsHandler) getRoomDetails(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	reqParams := utils.ConstructFilter(r)
	id := ""
	for _, v := range reqParams.Items {
		if v.Field == "id" {
			id = v.Value[0]
			break
		}
	}

	if id == "" || id == "nil" {
		return l.HTTPResponseErrorData(logutils.StatusMissing, logutils.TypeQueryParam, logutils.StringArgs("id"), nil, http.StatusBadRequest, false)
	}

	rd, err := h.app.Client.GetLaundryRoom(id)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionGet, model.TypeLaundryRooms, nil, err, http.StatusInternalServerError, true)
	}

	if rd == nil {
		return l.HTTPResponseErrorAction(logutils.ActionGet, model.TypeLaundryRooms, nil, err, http.StatusNotFound, true)

	}

	resAsJSON, err := json.Marshal(rd)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionMarshal, logutils.TypeResult, nil, err, http.StatusInternalServerError, false)
	}

	return l.HTTPResponseSuccessJSON(resAsJSON)
}

// InitServiceRequest returns a laundry room detail record
// @Summary Returns the problem codes and pending service reqeust status for a laundry machine.
// @Tags Client
// @ID InitRequest
// @Param machineid query string true "machine service tag id"
// @Accept  json
// @Success 200 {object} model.MachineRequestDetail
// @Security RokwireAuth
// @Router /laundry/initrequest [get]
func (h ClientAPIsHandler) initServiceRequest(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	reqParams := utils.ConstructFilter(r)
	id := ""
	for _, v := range reqParams.Items {
		if v.Field == "machineid" {
			//do work here
			id = v.Value[0]
			break
		}
	}
	if id == "" || id == "nil" {
		return l.HTTPResponseErrorData(logutils.StatusMissing, logutils.TypeQueryParam, logutils.StringArgs("machineid"), nil, http.StatusBadRequest, false)
	}

	mrd, err := h.app.Client.InitServiceRequest(id)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionGet, model.TypeLaundryServiceSubmission, nil, err, http.StatusInternalServerError, true)
	}

	if mrd == nil {
		return l.HTTPResponseErrorAction(logutils.ActionGet, model.TypeLaundryRooms, nil, err, http.StatusNotFound, true)

	}

	resAsJSON, err := json.Marshal(mrd)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionMarshal, logutils.TypeResult, nil, err, http.StatusInternalServerError, false)
	}

	return l.HTTPResponseSuccessJSON(resAsJSON)
}

// SubmitServiceRequest returns the results of attempting to submit a service request for a laundyr appliance
// @Tags Client
// @ID RequestService
// @Param data body model.ServiceSubmission true "body json"
// @Accept  json
// @Success 200 {object} model.ServiceRequestResult
// @Security RokwireAuth
// @Router /laundry/requestservice [post]
func (h ClientAPIsHandler) submitServiceRequest(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		return l.HTTPResponseErrorData(logutils.StatusInvalid, logutils.TypeRequestBody, nil, err, http.StatusBadRequest, false)
	}

	var record model.ServiceSubmission
	err = json.Unmarshal(data, &record)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionMarshal, logutils.TypeRequestBody, nil, err, http.StatusBadRequest, false)
	}

	if record.MachineID == nil || len(*record.MachineID) == 0 {
		return l.HTTPResponseErrorData(logutils.StatusMissing, logutils.TypeRequestBody, nil, err, http.StatusBadRequest, false)
	}

	if record.ProblemCode == nil || len(*record.ProblemCode) == 0 {
		return l.HTTPResponseErrorData(logutils.StatusMissing, logutils.TypeRequestBody, nil, err, http.StatusBadRequest, false)
	}

	if record.FirstName == nil || len(*record.FirstName) == 0 {
		return l.HTTPResponseErrorData(logutils.StatusMissing, logutils.TypeRequestBody, nil, err, http.StatusBadRequest, false)
	}

	if record.LastName == nil || len(*record.LastName) == 0 {
		return l.HTTPResponseErrorData(logutils.StatusMissing, logutils.TypeRequestBody, nil, err, http.StatusBadRequest, false)
	}

	if record.Email == nil || len(*record.Email) == 0 {
		return l.HTTPResponseErrorData(logutils.StatusMissing, logutils.TypeRequestBody, nil, err, http.StatusBadRequest, false)
	}

	if record.Phone == nil {
		newPhone := ""
		record.Phone = &newPhone
	}

	sr, err := h.app.Client.SubmitServiceRequest(*record.MachineID, *record.ProblemCode, *record.Comments, *record.FirstName, *record.LastName, *record.Phone, *record.Email)

	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionGet, model.TypeLaundryServiceSubmission, nil, err, http.StatusInternalServerError, true)
	}

	resAsJSON, err := json.Marshal(sr)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionMarshal, logutils.TypeResult, nil, err, http.StatusInternalServerError, false)
	}
	return l.HTTPResponseSuccessJSON(resAsJSON)
}

// NewClientAPIsHandler creates new client API handler instance
func NewClientAPIsHandler(app *core.Application) ClientAPIsHandler {
	return ClientAPIsHandler{app: app}
}

func (h ClientAPIsHandler) setReturnDataOnHTTPError(l *logs.Log, statuscode int) logs.HTTPResponse {
	switch statuscode {
	case 401:
		return l.HTTPResponseErrorData(logutils.MessageDataStatus(logutils.StatusError), logutils.TypeClaim, logutils.StringArgs("id"), nil, http.StatusForbidden, false)
	case 403:
		return l.HTTPResponseErrorData(logutils.MessageDataStatus(logutils.StatusError), logutils.TypeClaim, logutils.StringArgs("id"), nil, http.StatusForbidden, false)
	case 404:
		return l.HTTPResponseErrorAction(logutils.ActionFind, logutils.TypeResult, logutils.StringArgs("id"), nil, statuscode, false)
	default:
		return l.HTTPResponseErrorData(logutils.MessageDataStatus(logutils.StatusError), logutils.TypeError, logutils.StringArgs("id"), nil, http.StatusInternalServerError, false)
	}
}
