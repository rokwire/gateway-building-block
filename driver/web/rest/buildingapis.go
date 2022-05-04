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
	"log"
	"net/http"
	"strconv"
)

type errorMessage struct {
	Message string
}

// BuildingAPIHandler handles the building rest APIs implementation
type BuildingAPIHandler struct {
	app *core.Application
}

// NewBuildingAPIHandler creates new rest Handler instance for building location functions
func NewBuildingAPIHandler(app *core.Application) BuildingAPIHandler {
	return BuildingAPIHandler{app: app}
}

// GetBuilding returns an the building matching the provided building id
// @Summary Get the requested building with all of its available entrances filterd by the ADA only flag
// @Tags Client
// @ID Building
// @Accept  json
// @Produce json
// @Param id query string true "Building identifier"
// @Param adaOnly query bool false "ADA entrances filter"
// @Success 200 {object} model.Building
// @Security RokwireAuth
// @Router /wayfinding/building [get]
func (h BuildingAPIHandler) GetBuilding(w http.ResponseWriter, r *http.Request) {

	bldgid := ""
	adaOnly := false
	reqParams := utils.ConstructFilter(r)
	for _, v := range reqParams.Items {
		if v.Field == "id" {
			bldgid = v.Value[0]
		}
		if v.Field == "adaOnly" {
			ada, err := strconv.ParseBool(v.Value[0])
			if err != nil {
				log.Printf("Invalid parameter value (adaOnly): %s\n", err)
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			adaOnly = ada
		}
	}

	if bldgid == "" {
		log.Printf("Error on retrieving building informaiton: missing id parameter")
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	bldg, err := h.app.Services.GetBuilding(bldgid, adaOnly)
	if err != nil {
		log.Printf("Error retrieving building details: %s\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resAsJSON, err := json.Marshal(bldg)
	if err != nil {
		log.Printf("Error on marshalling building data: %s\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(resAsJSON)
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
func (h BuildingAPIHandler) GetEntrance(w http.ResponseWriter, r *http.Request) {
	reqParams := utils.ConstructFilter(r)
	bldgID := ""
	adaOnly := false
	var latitude, longitude float64

	if len(reqParams.Items) < 3 || len(reqParams.Items) > 4 {
		log.Printf("Invalid number of parameters passed")
		http.Error(w, "Invalid number of parameters", http.StatusBadRequest)
		return
	}

	for _, v := range reqParams.Items {
		if v.Field == "id" {
			bldgID = v.Value[0]
		}

		if v.Field == "adaOnly" {
			ada, err := strconv.ParseBool(v.Value[0])
			if err != nil {
				log.Printf("Invalid parameter value (adaOnly): %s\n", err)
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			adaOnly = ada
		}

		if v.Field == "lat" {
			lat, err := strconv.ParseFloat(v.Value[0], 64)
			if err != nil {
				log.Printf("Invalid parameter value (lat): %s\n", err)
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			latitude = lat
		}

		if v.Field == "long" {
			long, err := strconv.ParseFloat(v.Value[0], 64)
			if err != nil {
				log.Printf("Invalid parameter value (long): %s\n", err)
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			longitude = long
		}
	}
	if latitude == 0 && longitude == 0 {
		log.Printf("Missing latitude or longitude parameter")
		http.Error(w, "Missing latitude or longitude parameter", http.StatusBadRequest)
		return
	}

	if bldgID == "" {
		log.Printf("Error on retrieving entrance: missing id parameter")
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	entrance, err := h.app.Services.GetEntrance(bldgID, adaOnly, latitude, longitude)
	if err != nil {
		log.Printf("Error retrieving entrance: %s\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if entrance == nil {
		w.WriteHeader(http.StatusNotFound)
		resp := errorMessage{Message: "Resource Not Found"}
		jsonResp, err := json.Marshal(resp)
		if err != nil {
			log.Printf("Error on marshaling json response. Err: %s", err)
			return
		}
		w.Write(jsonResp)
		return

	}
	resAsJSON, err := json.Marshal(entrance)
	if err != nil {
		log.Printf("Error on marshalling entrance: %s\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(resAsJSON)

}
