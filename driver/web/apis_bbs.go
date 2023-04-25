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
	"errors"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/rokwire/core-auth-library-go/v3/tokenauth"
	"github.com/rokwire/logging-library-go/v2/logs"
	"github.com/rokwire/logging-library-go/v2/logutils"
)

// BBsAPIsHandler handles the rest BBs APIs implementation
type BBsAPIsHandler struct {
	app *core.Application
}

func (h BBsAPIsHandler) getExample(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	params := mux.Vars(r)
	id := params["id"]
	if len(id) <= 0 {
		return l.HTTPResponseErrorData(logutils.StatusMissing, logutils.TypePathParam, logutils.StringArgs("id"), nil, http.StatusBadRequest, false)
	}

	example, err := h.app.BBs.GetExample(claims.OrgID, claims.AppID, id)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionGet, model.TypeExample, nil, err, http.StatusInternalServerError, true)
	}

	response, err := json.Marshal(example)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionMarshal, logutils.TypeResponseBody, nil, err, http.StatusInternalServerError, false)
	}
	return l.HTTPResponseSuccessJSON(response)
}

// appointment apis
func (h BBsAPIsHandler) getAppointmentUnits(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {

	reqParams := utils.ConstructFilter(r)

	reqValues, resp, err := h.checkAppointmentParams(reqParams, r, l)
	if err != nil {
		return resp
	}

	if len(reqValues.UIN) != 9 {
		return l.HTTPResponseErrorData(logutils.StatusMissing, logutils.TypePathParam, logutils.StringArgs("external_id"), nil, http.StatusBadRequest, false)
	}

	example, err := h.app.BBs.GetAppointmentUnits(reqValues.ProviderID, reqValues.UIN, reqValues.ExternalToken)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionGet, model.TypeAppointments, nil, err, http.StatusInternalServerError, true)
	}

	response, err := json.Marshal(example)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionMarshal, logutils.TypeResponseBody, nil, err, http.StatusInternalServerError, false)
	}
	return l.HTTPResponseSuccessJSON(response)
}

func (h BBsAPIsHandler) getAppointmentPeople(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	reqParams := utils.ConstructFilter(r)
	reqValues, resp, err := h.checkAppointmentParams(reqParams, r, l)
	if err != nil {
		return resp
	}

	if reqValues.ProviderID == 0 {
		return l.HTTPResponseErrorData(logutils.StatusInvalid, logutils.TypeQueryParam, logutils.StringArgs("provider_id"), nil, http.StatusBadRequest, false)
	}

	if reqValues.UnitID == 0 {
		return l.HTTPResponseErrorData(logutils.StatusInvalid, logutils.TypeQueryParam, logutils.StringArgs("unit_id"), nil, http.StatusBadRequest, false)
	}

	if reqValues.UIN == "" {
		return l.HTTPResponseErrorData(logutils.StatusMissing, logutils.TypePathParam, logutils.StringArgs("external_id"), nil, http.StatusBadRequest, false)
	}

	people, err := h.app.BBs.GetPeople(reqValues.UIN, reqValues.UnitID, reqValues.ProviderID, reqValues.ExternalToken)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionGet, model.TypeAppointments, nil, err, http.StatusInternalServerError, true)
	}

	response, err := json.Marshal(people)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionMarshal, logutils.TypeResponseBody, nil, err, http.StatusInternalServerError, false)
	}
	return l.HTTPResponseSuccessJSON(response)
}

func (h BBsAPIsHandler) getAppointmentOptions(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {

	reqParams := utils.ConstructFilter(r)

	reqValues, resp, err := h.checkAppointmentParams(reqParams, r, l)
	if err != nil {
		return resp
	}

	//check request only validates a parameter value if it is in the parameters. Need to make sure we have everythign we need for this call
	if reqValues.UIN == "" || reqValues.UnitID == 0 || reqValues.ProviderID == 0 || reqValues.PersonID == 0 || (reqValues.StartTime.IsZero() && !reqValues.EndTime.IsZero()) || (!reqValues.StartTime.IsZero() && reqValues.EndTime.IsZero()) {
		return l.HTTPResponseErrorData(logutils.StatusInvalid, logutils.TypeQueryParam, logutils.StringArgs("missing parameter"), nil, http.StatusBadRequest, false)
	}

	apptOptions, err := h.app.BBs.GetAppointmentOptions(reqValues.UIN, reqValues.UnitID, reqValues.PersonID, reqValues.ProviderID, reqValues.StartTime, reqValues.EndTime, reqValues.ExternalToken)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionGet, model.TypeAppointments, nil, err, http.StatusInternalServerError, true)
	}

	response, err := json.Marshal(apptOptions)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionMarshal, logutils.TypeResponseBody, nil, err, http.StatusInternalServerError, false)
	}
	return l.HTTPResponseSuccessJSON(response)
}

func (h BBsAPIsHandler) createAppointment(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {

	data, err := io.ReadAll(r.Body)
	if err != nil {
		return l.HTTPResponseErrorData(logutils.StatusInvalid, logutils.TypeRequestBody, nil, err, http.StatusBadRequest, false)
	}

	var record model.AppointmentPost
	err = json.Unmarshal(data, &record)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionMarshal, logutils.TypeRequestBody, nil, err, http.StatusBadRequest, false)
	}

	externalToken := r.Header.Get("External-Authorization")
	if externalToken == "" {
		return l.HTTPResponseErrorData(logutils.StatusMissing, logutils.TypeHeader, logutils.StringArgs("external auth token"), nil, http.StatusBadRequest, false)
	}

	newAppt, err := h.app.BBs.CreateAppointment(&record, externalToken)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionGet, model.TypeAppointments, nil, err, http.StatusInternalServerError, true)
	}

	response, err := json.Marshal(newAppt)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionMarshal, logutils.TypeResponseBody, nil, err, http.StatusInternalServerError, false)
	}
	return l.HTTPResponseSuccessJSON(response)
}

func (h BBsAPIsHandler) deleteAppointment(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {

	params := mux.Vars(r)
	id := params["id"]
	if len(id) <= 0 {
		return l.HTTPResponseErrorData(logutils.StatusMissing, logutils.TypePathParam, logutils.StringArgs("id"), nil, http.StatusBadRequest, false)
	}

	reqParams := utils.ConstructFilter(r)
	reqValues, resp, err := h.checkAppointmentParams(reqParams, r, l)
	if err != nil {
		return resp
	}

	if reqValues.UIN == "" {
		return l.HTTPResponseErrorData(logutils.StatusInvalid, logutils.TypeQueryParam, logutils.StringArgs("external_id"), nil, http.StatusBadRequest, false)
	}

	if reqValues.ProviderID == 0 {
		return l.HTTPResponseErrorData(logutils.StatusInvalid, logutils.TypeQueryParam, logutils.StringArgs("provider_id"), nil, http.StatusBadRequest, false)
	}

	_, err = h.app.BBs.DeleteAppointment(reqValues.UIN, reqValues.ProviderID, id, reqValues.ExternalToken)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionGet, model.TypeAppointments, nil, err, http.StatusInternalServerError, true)
	}

	return l.HTTPResponseSuccess()
}

func (h BBsAPIsHandler) getAppointmentTimeSlots(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {

	reqParams := utils.ConstructFilter(r)

	reqValues, resp, err := h.checkAppointmentParams(reqParams, r, l)
	if err != nil {
		return resp
	}

	//check request only validates a parameter value if it is in the parameters. Need to make sure we have everythign we need for this call
	if reqValues.UIN == "" || reqValues.UnitID == 0 || reqValues.ProviderID == 0 || reqValues.PersonID == 0 || (reqValues.StartTime.IsZero() && !reqValues.EndTime.IsZero()) || (!reqValues.StartTime.IsZero() && reqValues.EndTime.IsZero()) {
		return l.HTTPResponseErrorData(logutils.StatusInvalid, logutils.TypeQueryParam, logutils.StringArgs("missing parameter"), nil, http.StatusBadRequest, false)
	}

	apptOptions, err := h.app.BBs.GetAppointmentOptions(reqValues.UIN, reqValues.UnitID, reqValues.PersonID, reqValues.ProviderID, reqValues.StartTime, reqValues.EndTime, reqValues.ExternalToken)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionGet, model.TypeAppointments, nil, err, http.StatusInternalServerError, true)
	}

	response, err := json.Marshal(apptOptions.TimeSlots)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionMarshal, logutils.TypeResponseBody, nil, err, http.StatusInternalServerError, false)
	}
	return l.HTTPResponseSuccessJSON(response)
}

func (h BBsAPIsHandler) getAppointmentQuestions(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {

	reqParams := utils.ConstructFilter(r)

	reqValues, resp, err := h.checkAppointmentParams(reqParams, r, l)
	if err != nil {
		return resp
	}

	//check request only validates a parameter value if it is in the parameters. Need to make sure we have everythign we need for this call
	if reqValues.UIN == "" || reqValues.UnitID == 0 || reqValues.ProviderID == 0 || reqValues.PersonID == 0 {
		return l.HTTPResponseErrorData(logutils.StatusInvalid, logutils.TypeQueryParam, logutils.StringArgs("missing parameter"), nil, http.StatusBadRequest, false)
	}

	apptOptions, err := h.app.BBs.GetAppointmentOptions(reqValues.UIN, reqValues.UnitID, reqValues.PersonID, reqValues.ProviderID, reqValues.StartTime, reqValues.EndTime, reqValues.ExternalToken)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionGet, model.TypeAppointments, nil, err, http.StatusInternalServerError, true)
	}

	response, err := json.Marshal(apptOptions.Questions)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionMarshal, logutils.TypeResponseBody, nil, err, http.StatusInternalServerError, false)
	}
	return l.HTTPResponseSuccessJSON(response)
}

func (h BBsAPIsHandler) checkAppointmentParams(reqParms *utils.Filter, req *http.Request, l *logs.Log) (timeSlotRequest, logs.HTTPResponse, error) {

	reqValues := timeSlotRequest{UnitID: 0, ProviderID: 0, UIN: "", PersonID: 0, ExternalToken: ""}
	externalToken := req.Header.Get("External-Authorization")
	if externalToken == "" {
		return reqValues, l.HTTPResponseErrorData(logutils.StatusMissing, logutils.TypeHeader, logutils.StringArgs("external auth token"), nil, http.StatusBadRequest, false), errors.New("missing auth token")
	}
	reqValues.ExternalToken = externalToken

	for _, v := range reqParms.Items {
		switch v.Field {
		case "provider_id":
			provideridstr := v.Value[0]
			intvar, err := strconv.Atoi(provideridstr)
			if err != nil {
				return reqValues, l.HTTPResponseErrorData(logutils.StatusInvalid, logutils.TypeQueryParam, logutils.StringArgs("provider_id"), nil, http.StatusBadRequest, false), err
			}
			if intvar == 0 {
				return reqValues, l.HTTPResponseErrorData(logutils.StatusInvalid, logutils.TypeQueryParam, logutils.StringArgs("provider_id"), nil, http.StatusBadRequest, false), errors.New("invalid providerid")
			}
			_, ok := h.app.AppointmentAdapters[provideridstr]
			if !ok {
				return reqValues, l.HTTPResponseErrorData(logutils.StatusInvalid, logutils.TypeQueryParam, logutils.StringArgs("provider_id"), nil, http.StatusBadRequest, false), errors.New("invalid providerid")
			}
			reqValues.ProviderID = intvar
		case "unit_id":
			unitidstr := v.Value[0]
			intvar, err := strconv.Atoi(unitidstr)
			if err != nil {
				return reqValues, l.HTTPResponseErrorData(logutils.StatusInvalid, logutils.TypeQueryParam, logutils.StringArgs("unit_id"), nil, http.StatusBadRequest, false), err
			}
			if intvar == 0 {
				return reqValues, l.HTTPResponseErrorData(logutils.StatusInvalid, logutils.TypeQueryParam, logutils.StringArgs("unit_id"), nil, http.StatusBadRequest, false), errors.New("invalid unitid")
			}
			reqValues.UnitID = intvar
		case "person_id":
			peopleidstr := v.Value[0]
			intvar, err := strconv.Atoi(peopleidstr)
			if err != nil {
				return reqValues, l.HTTPResponseErrorData(logutils.StatusInvalid, logutils.TypeQueryParam, logutils.StringArgs("person_id"), nil, http.StatusBadRequest, false), err
			}
			if intvar == 0 {
				return reqValues, l.HTTPResponseErrorData(logutils.StatusMissing, logutils.TypePathParam, logutils.StringArgs("external_id"), nil, http.StatusBadRequest, false), errors.New("invalid uin")
			}
			reqValues.PersonID = intvar
		case "external_id":
			reqValues.UIN = v.Value[0]
		case "start_time":
			st, err := time.Parse(time.DateOnly, v.Value[0])
			if err != nil {
				return reqValues, l.HTTPResponseErrorData(logutils.StatusInvalid, logutils.TypeQueryParam, logutils.StringArgs("start_time"), nil, http.StatusBadRequest, false), err
			}
			reqValues.StartTime = st
		case "end_time":
			et, err := time.Parse(time.DateOnly, v.Value[0])
			if err != nil {
				return reqValues, l.HTTPResponseErrorData(logutils.StatusInvalid, logutils.TypeQueryParam, logutils.StringArgs("end_time"), nil, http.StatusBadRequest, false), err
			}
			reqValues.EndTime = et
		}
	}

	if reqValues.ProviderID == 0 {
		return reqValues, l.HTTPResponseErrorData(logutils.StatusMissing, logutils.TypeQueryParam, logutils.StringArgs("provider_id"), nil, http.StatusBadRequest, false), errors.New("missing provider_id")
	}
	return reqValues, l.HTTPResponseSuccess(), nil

}

// NewBBsAPIsHandler creates new Building Block API handler instance
func NewBBsAPIsHandler(app *core.Application) BBsAPIsHandler {
	return BBsAPIsHandler{app: app}
}

type timeSlotRequest struct {
	UnitID        int
	ProviderID    int
	UIN           string
	PersonID      int
	StartTime     time.Time
	EndTime       time.Time
	ExternalToken string
}
