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

	"github.com/rokwire/core-auth-library-go/v3/tokenauth"
	"github.com/rokwire/logging-library-go/v2/logs"
	"github.com/rokwire/logging-library-go/v2/logutils"
)

// APIKeyHandler handles api key
type APIKeyHandler struct {
	app *core.Application
}

func (h APIKeyHandler) getLegacyEvents(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	legacyEvents, err := h.app.BBs.GetLegacyEvents()
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionGet, model.TypeAppointments, nil, err, http.StatusInternalServerError, true)
	}
	response, err := json.Marshal(legacyEvents)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionMarshal, logutils.TypeResponseBody, nil, err, http.StatusInternalServerError, false)
	}
	return l.HTTPResponseSuccessJSON(response)
}

// NewAPIKeyHandler creates new api key handler
func NewAPIKeyHandler(app *core.Application) APIKeyHandler {
	return APIKeyHandler{app: app}
}
