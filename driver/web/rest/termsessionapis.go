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

package rest

import (
	"apigateway/core"
	"encoding/json"
	"log"
	"net/http"
)

// TermSessionAPIHandler handles the term session information rest APIs implementation
type TermSessionAPIHandler struct {
	app *core.Application
}

// NewTermSessionAPIHandler creates new rest Handler instance for getting term sessions
func NewTermSessionAPIHandler(app *core.Application) TermSessionAPIHandler {
	return TermSessionAPIHandler{app: app}
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
func (h TermSessionAPIHandler) GetTermSessions(w http.ResponseWriter, r *http.Request) {

	termSessions, err := h.app.Services.GetTermSessions()
	if err != nil {
		log.Printf("Error retrieving term sessions: %s\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resAsJSON, err := json.Marshal(termSessions)
	if err != nil {
		log.Printf("Error on marshalling term session data: %s\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(resAsJSON)
}
