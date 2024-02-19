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

func (h TPSAPIsHandler) createEvent(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		return l.HTTPResponseErrorData(logutils.StatusInvalid, logutils.TypeRequestBody, nil, err, http.StatusBadRequest, false)
	}

	var e model.LegacyEvent
	err = json.Unmarshal(data, &e)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionMarshal, logutils.TypeRequestBody, nil, err, http.StatusBadRequest, false)
	}

	syncSourse := "events-tps-api"
	syncDate := time.Now()
	ID := uuid.NewString()
	if e.StartDate != "" {
		startDate, err := time.Parse("2006/01/02T15:04:05", e.StartDate)
		if err != nil {
			return l.HTTPResponseErrorAction(logutils.ActionMarshal, logutils.TypeRequestBody, nil, err, http.StatusBadRequest, false)

		}
		e.StartDate = startDate.Format(time.RFC3339)
	}
	if e.EndDate != "" {
		endDate, err := time.Parse("2006/01/02T15:04:05", e.EndDate)
		if err != nil {
			return l.HTTPResponseErrorAction(logutils.ActionMarshal, logutils.TypeRequestBody, nil, err, http.StatusBadRequest, false)

		}
		e.EndDate = endDate.Format(time.RFC3339)
	}

	var location *model.LocationLegacy
	if e.Location != nil || e.Location.Longitude == 0 || e.Location.Latitude == 0 {
		location = &model.LocationLegacy{
			Description: e.Location.Description,
			Longitude:   e.Location.Longitude,
			Latitude:    e.Location.Latitude,
		}
	}

	createdEvent := model.LegacyEventItem{SyncProcessSource: syncSourse, SyncDate: syncDate,
		Item: model.LegacyEvent{AllDay: e.AllDay, CalendarID: e.CalendarID, Category: e.Category, Subcategory: e.Subcategory,
			CreatedBy: e.CreatedBy, LongDescription: e.LongDescription, DataModified: e.DataModified, DataSourceEventID: e.DataSourceEventID,
			DateCreated: e.DateCreated, EndDate: e.EndDate, EventID: ID, IcalURL: e.IcalURL, ID: ID, ImageURL: e.ImageURL,
			IsEventFree: e.IsEventFree, IsVirtial: e.IsVirtial, Location: location, OriginatingCalendarID: e.OriginatingCalendarID,
			OutlookURL: e.OutlookURL, RecurrenceID: e.RecurrenceID, IsSuperEvent: e.IsSuperEvent, RecurringFlag: e.RecurringFlag,
			SourceID: e.SourceID, Sponsor: e.Sponsor, StartDate: e.StartDate, Title: e.Title, TitleURL: e.TitleURL,
			RegistrationURL: e.RegistrationURL, Contacts: e.Contacts, SubEvents: e.SubEvents}}

	_, err = h.app.TPS.CreateEvent(&createdEvent)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionGet, model.TypeExample, nil, err, http.StatusInternalServerError, true)
	}

	response, err := json.Marshal(createdEvent)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionMarshal, logutils.TypeResponseBody, nil, err, http.StatusInternalServerError, false)
	}
	return l.HTTPResponseSuccessJSON(response)
}

// NewTPSAPIsHandler creates new third-party service API handler instance
func NewTPSAPIsHandler(app *core.Application) TPSAPIsHandler {
	return TPSAPIsHandler{app: app}
}
