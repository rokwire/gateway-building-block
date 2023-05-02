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
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rokwire/core-auth-library-go/v3/authutils"
	"github.com/rokwire/core-auth-library-go/v3/tokenauth"
	"github.com/rokwire/logging-library-go/v2/logs"
	"github.com/rokwire/logging-library-go/v2/logutils"
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
		appID = authutils.AllApps
	}
	orgID := claims.OrgID
	if requestData.AllOrgs != nil && *requestData.AllOrgs {
		orgID = authutils.AllOrgs
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
		appID = authutils.AllApps
	}
	orgID := claims.OrgID
	if requestData.AllOrgs != nil && *requestData.AllOrgs {
		orgID = authutils.AllOrgs
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

// NewAdminAPIsHandler creates new rest Handler instance
func NewAdminAPIsHandler(app *core.Application) AdminAPIsHandler {
	return AdminAPIsHandler{app: app}
}
