/*
 *   Copyright (c) 2020 Board of Trustees of the University of Illinois.
 *   All rights reserved.

 *   Licensed under the Apache License, Version 2.0 (the "License");
 *   you may not use this file except in compliance with the License.
 *   You may obtain a copy of the License at

 *   http://www.apache.org/licenses/LICENSE-2.0

 *   Unless required by applicable law or agreed to in writing, software
 *   distributed under the License is distributed on an "AS IS" BASIS,
 *   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *   See the License for the specific language governing permissions and
 *   limitations under the License.
 */

package rest

import (
	"apigateway/core"
	"encoding/json"
	"log"
	"net/http"
)

// LaundryApisHandler handles the laudnry rest APIs implementation
type LaundryApisHandler struct {
	app *core.Application
}

// NewLaundryApisHandler creates new rest Handler instance for Laundry functions
func NewLaundryApisHandler(app *core.Application) LaundryApisHandler {
	return LaundryApisHandler{app: app}
}

// GetLaundryRooms returns an organization record
// @Tags Client
// @ID Name
// @Param data body sampleRecord true "body json"
// @Accept  json
// @Success 200
// @Security RokwireAuth UserAuth
// @Router /token [post]
func (h LaundryApisHandler) GetLaundryRooms(w http.ResponseWriter, r *http.Request) {

	org, err := h.app.Services.ListLaundryRooms()
	log.Printf("Org is " + org.SchoolName)
	if err != nil {
		log.Printf("Error on creating student guide: %s\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resAsJSON, err := json.Marshal(org)
	if err != nil {
		log.Printf("Error on marshalling laundry room list: %s\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(resAsJSON)
}
