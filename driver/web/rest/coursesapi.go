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
	model "apigateway/core/model"
	"apigateway/utils"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
)

// CourseApisHandler handles the course information rest APIs implementation
type CourseApisHandler struct {
	app *core.Application
}

// NewCourseApisHandler creates new rest Handler instance for course info functions
func NewCourseApisHandler(app *core.Application) CourseApisHandler {
	return CourseApisHandler{app: app}
}

// GetGiesCourses returns a list of registered courses for GIES students
// @Summary Returns a list of registered courses
// @Tags Client
// @ID GiesCourses
// @Param id query string true "User ID"
// @Accept  json
// @Produce json
// @Success 200 {object} []model.GiesCourse
// @Security RokwireAuth ExternalAuth
// @Router /courses/giescourses [get]
func (h CourseApisHandler) GetGiesCourses(w http.ResponseWriter, r *http.Request) {

	externalToken := r.Header.Get("External-Authorization")

	id := ""

	reqParams := utils.ConstructFilter(r)
	if reqParams != nil {
		for _, v := range reqParams.Items {
			switch v.Field {
			case "id":
				id = v.Value[0]
			}
		}
	}

	if id == "" || id == "null" {
		log.Printf("Error: missing id parameter")
		http.Error(w, "Missing id parameter", http.StatusBadRequest)
		return
	}

	if externalToken == "" {
		log.Printf("Error: External access token not includeed for %s", id)
		http.Error(w, "Missing external access token", http.StatusBadRequest)
		return
	}

	giesCourseList, statusCode, err := h.app.Services.GetGiesCourses(id, externalToken)
	if err != nil {
		log.Printf("Error getting gies courses for %s: Server returned %d %s \n", id, statusCode, err.Error())
		switch statusCode {
		case 401:
			http.Error(w, err.Error(), http.StatusUnauthorized)
		case 403:
			http.Error(w, err.Error(), http.StatusForbidden)
		case 404:
			http.Error(w, err.Error(), http.StatusNotFound)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	resAsJSON, err := json.Marshal(giesCourseList)
	if err != nil {
		log.Printf("Error on marshalling gies course information for %s: %s\n", id, err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(resAsJSON)
}

// GetStudentcourses returns a list of registered courses for a student
// @Summary Returns a list of registered courses
// @Tags Client
// @ID Studentcourses
// @Param id query string true "User ID"
// @Param termid query string true "term id"
// @Accept  json
// @Produce json
// @Success 200 {object} []model.Course
// @Security RokwireAuth ExternalAuth
// @Router /courses/studentcourses [get]
func (h CourseApisHandler) GetStudentcourses(w http.ResponseWriter, r *http.Request) {

	externalToken := r.Header.Get("External-Authorization")

	id := ""
	termid := ""
	adaOnly := false
	var latitude, longitude float64
	latitude = 0
	longitude = 0

	reqParams := utils.ConstructFilter(r)
	if reqParams != nil {
		for _, v := range reqParams.Items {
			switch v.Field {
			case "id":
				id = v.Value[0]
			case "termid":
				termid = v.Value[0]
			case "long":
				long, err := strconv.ParseFloat(v.Value[0], 64)
				if err != nil {
					log.Printf("Invalid parameter value (long): %s\n", err)
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				longitude = long
			case "lat":
				lat, err := strconv.ParseFloat(v.Value[0], 64)
				if err != nil {
					log.Printf("Invalid parameter value (lat): %s\n", err)
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				latitude = lat
			case "adaOnly":
				ada, err := strconv.ParseBool(v.Value[0])
				if err != nil {
					log.Printf("Invalid parameter value (adaOnly): %s\n", err)
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				adaOnly = ada
			}
		}
	}

	if id == "" || id == "null" {
		log.Printf("Error: missing id parameter")
		http.Error(w, "Missing id parameter", http.StatusBadRequest)
		return
	}

	if termid == "" || termid == "null" {
		log.Printf("Error: missing termid parameter")
		http.Error(w, "Missing termid parameter", http.StatusBadRequest)
		return
	}

	if externalToken == "" {
		log.Printf("Error: External access token not includeed for %s", id)
		http.Error(w, "Missing external access token", http.StatusBadRequest)
		return
	}

	courseList, statusCode, err := h.app.Services.GetStudentCourses(id, termid, externalToken)
	if err != nil {
		log.Printf("Error getting courses for %s: Server returned %d %s \n", id, statusCode, err.Error())
		switch statusCode {
		case 401:
			http.Error(w, err.Error(), http.StatusUnauthorized)
		case 403:
			http.Error(w, err.Error(), http.StatusForbidden)
		case 404:
			http.Error(w, err.Error(), http.StatusNotFound)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	//create a map of buildings we need so we don't retrieve the same building multiple times
	neededBuildings := make(map[string]model.Building)
	for index, crntCourse := range *courseList {
		//check if course is not online here then proceed
		if crntCourse.Section.BuildingID != "" {
			bldg, bldgexists := neededBuildings[crntCourse.Section.BuildingID]
			if bldgexists {
				(*courseList)[index].Section.Location = bldg
			} else {
				bldg, err := (h.app.Services.GetBuilding(crntCourse.Section.BuildingID, adaOnly, latitude, longitude))
				if err != nil {
					log.Printf("Error retrieving building details: %s\n", err)
				} else {
					(*courseList)[index].Section.Location = *bldg
					neededBuildings[bldg.ID] = *bldg
				}
			}
		}

	}

	resAsJSON, err := json.Marshal(courseList)
	if err != nil {
		log.Printf("Error on marshalling  course information for %s: %s\n", id, err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(resAsJSON)
}
