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
	"application/utils"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/rokwire/core-auth-library-go/v3/tokenauth"
	"github.com/rokwire/logging-library-go/v2/logs"
	"github.com/rokwire/logging-library-go/v2/logutils"
)

// TPSAPIsHandler handles the rest third-party service APIs implementation
type TPSAPIsHandler struct {
	app *core.Application
}

func (h TPSAPIsHandler) getExample(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	params := mux.Vars(r)
	id := params["id"]
	if len(id) <= 0 {
		return l.HTTPResponseErrorData(logutils.StatusMissing, logutils.TypePathParam, logutils.StringArgs("id"), nil, http.StatusBadRequest, false)
	}

	example, err := h.app.TPS.GetExample(claims.OrgID, claims.AppID, id)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionGet, model.TypeExample, nil, err, http.StatusInternalServerError, true)
	}

	response, err := json.Marshal(example)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionMarshal, logutils.TypeResponseBody, nil, err, http.StatusInternalServerError, false)
	}
	return l.HTTPResponseSuccessJSON(response)
}

func (h TPSAPIsHandler) createEvents(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		return l.HTTPResponseErrorData(logutils.StatusInvalid, logutils.TypeRequestBody, nil, err, http.StatusBadRequest, false)
	}

	var e []Def.TpsReqCreateEvent
	err = json.Unmarshal(data, &e)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionMarshal, logutils.TypeRequestBody, nil, err, http.StatusBadRequest, false)
	}

	var createdEvents []model.LegacyEventItem

	syncSourse := "events-tps-api"
	syncDate := time.Now()
	now := time.Now()
	createInfo := model.CreateInfo{Time: now, AccountID: claims.Subject}
	for _, w := range e {

		id := uuid.NewString()

		/*	type TpsReqCreateEvent struct {

			Contacts          *[]TpsReqCreateEventContact `json:"contacts,omitempty"`
			Location          *TpsReqCreateEventLocation  `json:"location,omitempty"`

			RecurrenceId      *int                        `json:"recurrence_id,omitempty"`
			RecurringFlag     *bool                       `json:"recurring_flag,omitempty"`
			RegistrationLabel *string                     `json:"registration_label,omitempty"`
			RegistrationUrl   *string                     `json:"registration_url,omitempty"`
			Sponsor           *string                     `json:"sponsor,omitempty"`
			Subcategocy       *string                     `json:"subcategocy,omitempty"`
			Tags              *[]string                   `json:"tags,omitempty"`
			TargetAudience    *[]string                   `json:"target_audience,omitempty"`
			Title             *string                     `json:"title,omitempty"`
			TitleUrl          *string                     `json:"title_url,omitempty"`
		} */

		legacyEvent := model.LegacyEvent{ID: id, AllDay: utils.GetBool(w.AllDay), Category: utils.GetString(w.Category),
			Cost: utils.GetString(w.Cost), CreatedBy: utils.GetString(w.CreatedBy), DataModified: utils.GetString(w.DateModified),
			StartDate: utils.GetString(w.StartDate), EndDate: utils.GetString(w.EndDate), ImageURL: w.ImageUrl,
			IsVirtial: utils.GetBool(w.IsVirtual), LongDescription: utils.GetString(w.LongDescription)}

		createdEvent := model.LegacyEventItem{
			SyncProcessSource: syncSourse, SyncDate: syncDate,
			Item:       legacyEvent,
			CreateInfo: createInfo}

		createdEvents = append(createdEvents, createdEvent)
	}

	/*	_, err = h.app.TPS.CreateEvents(createdEvents)
		if err != nil {
			return l.HTTPResponseErrorAction(logutils.ActionGet, model.TypeExample, nil, err, http.StatusInternalServerError, true)
		} */

	return l.HTTPResponseSuccess()
}

// NewTPSAPIsHandler creates new third-party service API handler instance
func NewTPSAPIsHandler(app *core.Application) TPSAPIsHandler {
	return TPSAPIsHandler{app: app}
}
