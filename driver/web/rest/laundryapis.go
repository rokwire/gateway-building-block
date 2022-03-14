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
	"apigateway/utils"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

type serviceSubmission struct {
	MachineID   *string `json:"machineid"`
	ProblemType *string `json:"problemtype"`
	Comments    *string `json:"comments"`
	FirstName   *string `json:"firstname"`
	LastName    *string `json:"lastname"`
	Phone       *string `json:"phone"`
}

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
// @Param  body sampleRecord true "body json"
// @Accept  json
// @Success 200
// @Security RokwireAuth UserAuth
// @Router /rooms [get]
func (h LaundryApisHandler) GetLaundryRooms(w http.ResponseWriter, r *http.Request) {

	org, err := h.app.Services.ListLaundryRooms()
	if err != nil {
		log.Printf("Error on listing laundry rooms: %s\n", err)
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

// GetRoomDetails returns a laundry room detail record
// @Tags Client
// @ID Name
// @Param id query
// @Accept  json
// @Success 200
// @Security RokwireAuth UserAuth
// @Router /roomdetail [get]
func (h LaundryApisHandler) GetRoomDetails(w http.ResponseWriter, r *http.Request) {
	reqParams := utils.ConstructFilter(r)
	id := ""
	for _, v := range reqParams.Items {
		if v.Field == "id" {
			id = v.Value[0]
			break
		}
	}

	if id != "" {
		rd, err := h.app.Services.GetLaundryRoom(id)
		if err != nil {
			log.Printf("Error retrieving laundry room details: %s\n", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		resAsJSON, err := json.Marshal(rd)
		if err != nil {
			log.Printf("Error on marshalling laundry room detail: %s\n", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write(resAsJSON)

	} else {
		//no id field was found
		log.Printf("Error on retrieving laundry detail: missing id parameter")
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

}

// InitServiceRequest returns a laundry room detail record
// @Tags Client
// @ID Name
// @Param id query
// @Accept  json
// @Success 200
// @Security RokwireAuth UserAuth
// @Router /roomdetail [get]
func (h LaundryApisHandler) InitServiceRequest(w http.ResponseWriter, r *http.Request) {
	reqParams := utils.ConstructFilter(r)
	id := ""
	for _, v := range reqParams.Items {
		if v.Field == "machineid" {
			//do work here
			id = v.Value[0]
			break
		}
	}

	if id != "" {
		mrd, err := h.app.Services.InitServiceRequest(id)
		if err != nil {
			log.Printf("Error retrieving machine service details: %s\n", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		resAsJSON, err := json.Marshal(mrd)
		if err != nil {
			log.Printf("Error on marshalling laundry room detail: %s\n", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write(resAsJSON)

	} else {
		//no id field was found
		log.Printf("Error on retrieving machine request detail: missing machine id parameter")
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
}

// SubmitServiceRequest returns the results of attempting to submit a service request for a laundyr appliance
// @Tags Client
// @ID Name
// @Param id query
// @Accept  json
// @Success 200
// @Security RokwireAuth UserAuth
// @Router /roomdetail [get]
func (h LaundryApisHandler) SubmitServiceRequest(w http.ResponseWriter, r *http.Request) {
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error on marshal token data - %s\n", err.Error())
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	var record serviceSubmission
	err = json.Unmarshal(data, &record)
	if err != nil {
		if jsonErr, ok := err.(*json.SyntaxError); ok {
			problemPart := data[jsonErr.Offset : jsonErr.Offset+10]
			log.Printf("json error new '%s'", problemPart)
		}
		log.Printf("Error on unmarshal the request submission data - %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if record.MachineID == nil || len(*record.MachineID) == 0 {
		log.Printf("machine id is empty or null")
		http.Error(w, fmt.Sprintf("token is empty or null\n"), http.StatusBadRequest)
		return
	}

	if record.ProblemType == nil || len(*record.ProblemType) == 0 {
		log.Printf("Problem type is empty or null")
		http.Error(w, fmt.Sprintf("token is empty or null\n"), http.StatusBadRequest)
		return
	}

	if record.FirstName == nil || len(*record.FirstName) == 0 {
		log.Printf("First name is empty or null")
		http.Error(w, fmt.Sprintf("token is empty or null\n"), http.StatusBadRequest)
		return
	}

	if record.LastName == nil || len(*record.LastName) == 0 {
		log.Printf("Last name is empty or null")
		http.Error(w, fmt.Sprintf("token is empty or null\n"), http.StatusBadRequest)
		return
	}

	sr, err := h.app.Services.SubmitServiceRequest(*record.MachineID, *record.ProblemType, *record.Comments, *record.FirstName, *record.LastName, *record.Phone)

	if err != nil {
		log.Printf("Error submitting laundry service request: %s\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resAsJSON, err := json.Marshal(sr)
	if err != nil {
		log.Printf("Error on marshalling laundry service request result: %s\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(resAsJSON)
}
