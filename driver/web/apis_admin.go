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
	Def "application/driver/web/docs/gen"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/rokwire/rokwire-building-block-sdk-go/services/core/auth/tokenauth"
	"github.com/rokwire/rokwire-building-block-sdk-go/utils/logging/logs"
	"github.com/rokwire/rokwire-building-block-sdk-go/utils/logging/logutils"
	"github.com/rokwire/rokwire-building-block-sdk-go/utils/rokwireutils"
)

// AdminAPIsHandler handles the rest Admin APIs implementation
type AdminAPIsHandler struct {
	app *core.Application
}

func (h AdminAPIsHandler) getExample(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	params := mux.Vars(r)
	id := params["id"]
	if len(id) <= 0 {
		return l.HTTPResponseErrorData(logutils.StatusMissing, logutils.TypePathParam, logutils.StringArgs("id"), nil, http.StatusBadRequest, false)
	}

	example, err := h.app.Admin.GetExample(claims.OrgID, claims.AppID, id)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionGet, model.TypeExample, nil, err, http.StatusInternalServerError, true)
	}

	response, err := json.Marshal(example)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionMarshal, logutils.TypeResponseBody, nil, err, http.StatusInternalServerError, false)
	}
	return l.HTTPResponseSuccessJSON(response)
}

func (h AdminAPIsHandler) createExample(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	var requestData model.Example
	err := json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionUnmarshal, logutils.TypeRequestBody, nil, err, http.StatusBadRequest, true)
	}

	requestData.OrgID = claims.OrgID
	requestData.AppID = claims.AppID
	example, err := h.app.Admin.CreateExample(requestData)
	if err != nil || example == nil {
		return l.HTTPResponseErrorAction(logutils.ActionCreate, model.TypeExample, nil, err, http.StatusInternalServerError, true)
	}

	response, err := json.Marshal(example)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionMarshal, logutils.TypeResponseBody, nil, err, http.StatusInternalServerError, false)
	}
	return l.HTTPResponseSuccessJSON(response)
}

func (h AdminAPIsHandler) updateExample(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	params := mux.Vars(r)
	id := params["id"]
	if len(id) <= 0 {
		return l.HTTPResponseErrorData(logutils.StatusMissing, logutils.TypePathParam, logutils.StringArgs("id"), nil, http.StatusBadRequest, false)
	}

	var requestData model.Example
	err := json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionUnmarshal, logutils.TypeRequestBody, nil, err, http.StatusBadRequest, true)
	}

	requestData.ID = id
	requestData.OrgID = claims.OrgID
	requestData.AppID = claims.AppID
	err = h.app.Admin.UpdateExample(requestData)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionUpdate, model.TypeExample, nil, err, http.StatusInternalServerError, true)
	}

	return l.HTTPResponseSuccess()
}

func (h AdminAPIsHandler) deleteExample(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	params := mux.Vars(r)
	id := params["id"]
	if len(id) <= 0 {
		return l.HTTPResponseErrorData(logutils.StatusMissing, logutils.TypePathParam, logutils.StringArgs("id"), nil, http.StatusBadRequest, false)
	}

	err := h.app.Admin.DeleteExample(claims.OrgID, claims.AppID, id)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionDelete, model.TypeExample, nil, err, http.StatusInternalServerError, true)
	}

	return l.HTTPResponseSuccess()
}

func (h AdminAPIsHandler) getConfig(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	params := mux.Vars(r)
	id := params["id"]
	if len(id) <= 0 {
		return l.HTTPResponseErrorData(logutils.StatusMissing, logutils.TypePathParam, logutils.StringArgs("id"), nil, http.StatusBadRequest, false)
	}

	config, err := h.app.Admin.GetConfig(id, claims)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionGet, model.TypeConfig, nil, err, http.StatusInternalServerError, true)
	}

	data, err := json.Marshal(config)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionMarshal, model.TypeConfig, nil, err, http.StatusInternalServerError, false)
	}

	return l.HTTPResponseSuccessJSON(data)
}

func (h AdminAPIsHandler) getConfigs(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	var configType *string
	typeParam := r.URL.Query().Get("type")
	if len(typeParam) > 0 {
		configType = &typeParam
	}

	configs, err := h.app.Admin.GetConfigs(configType, claims)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionGet, model.TypeConfig, nil, err, http.StatusInternalServerError, true)
	}

	data, err := json.Marshal(configs)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionMarshal, model.TypeConfig, nil, err, http.StatusInternalServerError, false)
	}

	return l.HTTPResponseSuccessJSON(data)
}

type adminUpdateConfigsRequest struct {
	AllApps *bool       `json:"all_apps,omitempty"`
	AllOrgs *bool       `json:"all_orgs,omitempty"`
	Data    interface{} `json:"data"`
	System  bool        `json:"system"`
	Type    string      `json:"type"`
}

func (h AdminAPIsHandler) createConfig(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	var requestData adminUpdateConfigsRequest
	err := json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionUnmarshal, logutils.TypeRequestBody, nil, err, http.StatusBadRequest, true)
	}

	appID := claims.AppID
	if requestData.AllApps != nil && *requestData.AllApps {
		appID = rokwireutils.AllApps
	}
	orgID := claims.OrgID
	if requestData.AllOrgs != nil && *requestData.AllOrgs {
		orgID = rokwireutils.AllOrgs
	}
	config := model.Config{Type: requestData.Type, AppID: appID, OrgID: orgID, System: requestData.System, Data: requestData.Data}

	newConfig, err := h.app.Admin.CreateConfig(config, claims)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionCreate, model.TypeConfig, nil, err, http.StatusInternalServerError, true)
	}

	data, err := json.Marshal(newConfig)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionMarshal, model.TypeConfig, nil, err, http.StatusInternalServerError, false)
	}

	return l.HTTPResponseSuccessJSON(data)
}

func (h AdminAPIsHandler) updateConfig(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	params := mux.Vars(r)
	id := params["id"]
	if len(id) <= 0 {
		return l.HTTPResponseErrorData(logutils.StatusMissing, logutils.TypePathParam, logutils.StringArgs("id"), nil, http.StatusBadRequest, false)
	}

	var requestData adminUpdateConfigsRequest
	err := json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionUnmarshal, logutils.TypeRequestBody, nil, err, http.StatusBadRequest, true)
	}

	appID := claims.AppID
	if requestData.AllApps != nil && *requestData.AllApps {
		appID = rokwireutils.AllApps
	}
	orgID := claims.OrgID
	if requestData.AllOrgs != nil && *requestData.AllOrgs {
		orgID = rokwireutils.AllOrgs
	}
	config := model.Config{ID: id, Type: requestData.Type, AppID: appID, OrgID: orgID, System: requestData.System, Data: requestData.Data}

	err = h.app.Admin.UpdateConfig(config, claims)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionUpdate, model.TypeConfig, nil, err, http.StatusInternalServerError, true)
	}

	return l.HTTPResponseSuccess()
}

func (h AdminAPIsHandler) deleteConfig(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	params := mux.Vars(r)
	id := params["id"]
	if len(id) <= 0 {
		return l.HTTPResponseErrorData(logutils.StatusMissing, logutils.TypePathParam, logutils.StringArgs("id"), nil, http.StatusBadRequest, false)
	}

	err := h.app.Admin.DeleteConfig(id, claims)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionDelete, model.TypeConfig, nil, err, http.StatusInternalServerError, true)
	}

	return l.HTTPResponseSuccess()
}

func (h AdminAPIsHandler) addwebtoolsblacklist(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	var requestData Def.AdminReqAddWebtoolsBlacklist
	err := json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionUnmarshal, logutils.TypeRequestBody, nil, err, http.StatusBadRequest, true)
	}

	var dataSourceIDs []string
	if requestData.DataSourceIds != nil {
		for _, w := range *requestData.DataSourceIds {
			if w != "" {
				dataSourceIDs = append(dataSourceIDs, w)
			}
		}
	}

	var dataCalendarIDs []string
	if requestData.DataCalendarIds != nil {
		for _, w := range *requestData.DataCalendarIds {
			if w != "" {
				dataCalendarIDs = append(dataCalendarIDs, w)
			}
		}
	}

	var dataOriginatingCalendarIDs []string
	if requestData.DataOriginatingCalendarIds != nil {
		for _, w := range *requestData.DataOriginatingCalendarIds {
			if w != "" {
				dataOriginatingCalendarIDs = append(dataOriginatingCalendarIDs, w)
			}
		}
	}

	err = h.app.Admin.AddWebtoolsBlackList(dataSourceIDs, dataCalendarIDs, dataOriginatingCalendarIDs)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionCreate, model.TypeConfig, nil, err, http.StatusInternalServerError, true)
	}

	return l.HTTPResponseSuccess()
}

func (h AdminAPIsHandler) getwebtoolsblacklist(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {

	blacklist, err := h.app.Admin.GetWebtoolsBlackList()
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionCreate, model.TypeConfig, nil, err, http.StatusInternalServerError, true)
	}

	data, err := json.Marshal(blacklist)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionMarshal, model.TypeConfig, nil, err, http.StatusInternalServerError, false)
	}

	return l.HTTPResponseSuccessJSON(data)
}

func (h AdminAPIsHandler) removewebtoolsblacklist(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	var sourceIdsList []string
	sourceIdsArg := r.URL.Query().Get("source_ids")

	if sourceIdsArg != "" {
		sourceIdsList = strings.Split(sourceIdsArg, ",")
	} else {
		sourceIdsList = nil
	}

	var calendarIdsList []string
	calendarIdsArg := r.URL.Query().Get("calendar_ids")

	if calendarIdsArg != "" {
		calendarIdsList = strings.Split(calendarIdsArg, ",")
	} else {
		calendarIdsList = nil
	}

	var originatingCalendarIdsList []string
	originatingCalendarIdsArg := r.URL.Query().Get("originating_calendar_ids")

	if originatingCalendarIdsArg != "" {
		originatingCalendarIdsList = strings.Split(originatingCalendarIdsArg, ",")
	} else {
		originatingCalendarIdsList = nil
	}

	err := h.app.Admin.RemoveWebtoolsBlackList(sourceIdsList, calendarIdsList, originatingCalendarIdsList)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionCreate, model.TypeConfig, nil, err, http.StatusInternalServerError, true)
	}

	return l.HTTPResponseSuccess()
}

func (h AdminAPIsHandler) getEventsSummary(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	summary, err := h.app.Admin.GetEventsSummary()
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionCreate, model.TypeConfig, nil, err, http.StatusInternalServerError, true)
	}

	data, err := json.Marshal(summary)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionMarshal, model.TypeConfig, nil, err, http.StatusInternalServerError, false)
	}

	return l.HTTPResponseSuccessJSON(data)
}

func (h AdminAPIsHandler) loadEvents(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {

	var source *string
	sourceParam := r.URL.Query().Get("source")
	if len(sourceParam) > 0 {
		source = &sourceParam
	}

	var status *string
	statusParam := r.URL.Query().Get("status")
	if len(statusParam) > 0 {
		status = &statusParam
	}

	var dataSourceEventID *string
	dataSourceEventIDParam := r.URL.Query().Get("data-source-event-id")
	if len(dataSourceEventIDParam) > 0 {
		dataSourceEventID = &dataSourceEventIDParam
	}

	var calendarID *string
	calendarIDParam := r.URL.Query().Get("calendar-id")
	if len(calendarIDParam) > 0 {
		calendarID = &calendarIDParam
	}
	var originatingCalendarID *string
	originatingCalendarIDParam := r.URL.Query().Get("originating-calendar-id")
	if len(originatingCalendarIDParam) > 0 {
		originatingCalendarID = &originatingCalendarIDParam
	}

	events, err := h.app.Admin.GetEventsItems(source, status, dataSourceEventID, calendarID, originatingCalendarID)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionCreate, model.TypeConfig, nil, err, http.StatusInternalServerError, true)
	}

	resEvents := legacyEventsItemsToDef(events)

	data, err := json.Marshal(resEvents)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionMarshal, model.TypeConfig, nil, err, http.StatusInternalServerError, false)
	}

	return l.HTTPResponseSuccessJSON(data)
}

// NewAdminAPIsHandler creates new rest Handler instance
func NewAdminAPIsHandler(app *core.Application) AdminAPIsHandler {
	return AdminAPIsHandler{app: app}
}
