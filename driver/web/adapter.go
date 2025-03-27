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
	"bytes"
	"fmt"
	"net/http"
	"os"
	"time"

	"gopkg.in/yaml.v2"

	"github.com/gorilla/mux"
	"github.com/rokwire/core-auth-library-go/v3/authservice"
	"github.com/rokwire/core-auth-library-go/v3/tokenauth"

	"github.com/rokwire/logging-library-go/v2/logs"
	"github.com/rokwire/logging-library-go/v2/logutils"

	httpSwagger "github.com/swaggo/http-swagger"
)

// Adapter entity
type Adapter struct {
	baseURL   string
	port      string
	serviceID string

	auth *Auth

	cachedYamlDoc []byte

	defaultAPIsHandler DefaultAPIsHandler
	clientAPIsHandler  ClientAPIsHandler
	adminAPIsHandler   AdminAPIsHandler
	bbsAPIsHandler     BBsAPIsHandler
	tpsAPIsHandler     TPSAPIsHandler
	systemAPIsHandler  SystemAPIsHandler
	apiKeyHandler      APIKeyHandler

	app *core.Application

	logger *logs.Logger
}

type handlerFunc = func(*logs.Log, *http.Request, *tokenauth.Claims) logs.HTTPResponse

// Start starts the module
func (a Adapter) Start() {

	router := mux.NewRouter().StrictSlash(true)

	// handle apis
	baseRouter := router.PathPrefix("/" + a.serviceID).Subrouter()
	baseRouter.PathPrefix("/doc/ui").Handler(a.serveDocUI())
	baseRouter.HandleFunc("/doc", a.serveDoc)
	baseRouter.HandleFunc("/version", a.wrapFunc(a.defaultAPIsHandler.version, nil)).Methods("GET")

	mainRouter := baseRouter.PathPrefix("/api").Subrouter()

	// Client APIs
	mainRouter.HandleFunc("/examples/{id}", a.wrapFunc(a.clientAPIsHandler.getExample, a.auth.client.Permissions)).Methods("GET")
	mainRouter.HandleFunc("/calendars/{id}", a.wrapFunc(a.clientAPIsHandler.getUnitCalendar, a.auth.client.User)).Methods("GET")
	mainRouter.HandleFunc("/laundry/rooms", a.wrapFunc(a.clientAPIsHandler.getLaundryRooms, a.auth.client.User)).Methods("GET")
	mainRouter.HandleFunc("/laundry/room", a.wrapFunc(a.clientAPIsHandler.getRoomDetails, a.auth.client.User)).Methods("GET")
	mainRouter.HandleFunc("/laundry/initrequest", a.wrapFunc(a.clientAPIsHandler.initServiceRequest, a.auth.client.User)).Methods("GET")
	mainRouter.HandleFunc("/laundry/requestservice", a.wrapFunc(a.clientAPIsHandler.submitServiceRequest, nil)).Methods("POST")

	mainRouter.HandleFunc("/wayfinding/building", a.wrapFunc(a.clientAPIsHandler.getBuilding, a.auth.client.Standard)).Methods("GET")
	mainRouter.HandleFunc("/wayfinding/entrance", a.wrapFunc(a.clientAPIsHandler.getEntrance, a.auth.client.Standard)).Methods("GET")
	mainRouter.HandleFunc("/wayfinding/buildings", a.wrapFunc(a.clientAPIsHandler.getBuildings, a.auth.client.Standard)).Methods("GET")
	mainRouter.HandleFunc("/wayfinding/floorplan", a.wrapFunc(a.clientAPIsHandler.getFloorPlan, a.auth.client.Standard)).Methods("GET")
	mainRouter.HandleFunc("/wayfinding/searchbuildings", a.wrapFunc(a.clientAPIsHandler.searchBuildings, a.auth.client.Standard)).Methods("GET")

	mainRouter.HandleFunc("/person/contactinfo", a.wrapFunc(a.clientAPIsHandler.getContactInfo, a.auth.client.User)).Methods("GET")
	mainRouter.HandleFunc("/courses/giescourses", a.wrapFunc(a.clientAPIsHandler.getGiesCourses, a.auth.client.User)).Methods("GET")
	mainRouter.HandleFunc("/courses/studentcourses", a.wrapFunc(a.clientAPIsHandler.getStudentCourses, a.auth.client.User)).Methods("GET")
	mainRouter.HandleFunc("/termsessions/listcurrent", a.wrapFunc(a.clientAPIsHandler.getTermSessions, a.auth.client.User)).Methods("GET")

	mainRouter.HandleFunc("/successteam", a.wrapFunc(a.clientAPIsHandler.getStudentSuccessTeam, a.auth.client.User)).Methods("GET")
	mainRouter.HandleFunc("/successteam/pcp", a.wrapFunc(a.clientAPIsHandler.getPrimaryCareProvider, a.auth.client.User)).Methods("GET")
	mainRouter.HandleFunc("/successteam/advisors", a.wrapFunc(a.clientAPIsHandler.getAcademicAdvisors, a.auth.client.User)).Methods("GET")

	// Admin APIs
	adminRouter := mainRouter.PathPrefix("/admin").Subrouter()
	adminRouter.HandleFunc("/examples/{id}", a.wrapFunc(a.adminAPIsHandler.getExample, a.auth.admin.Permissions)).Methods("GET")
	adminRouter.HandleFunc("/examples", a.wrapFunc(a.adminAPIsHandler.createExample, a.auth.admin.Permissions)).Methods("POST")
	adminRouter.HandleFunc("/examples/{id}", a.wrapFunc(a.adminAPIsHandler.updateExample, a.auth.admin.Permissions)).Methods("PUT")
	adminRouter.HandleFunc("/examples/{id}", a.wrapFunc(a.adminAPIsHandler.deleteExample, a.auth.admin.Permissions)).Methods("DELETE")

	adminRouter.HandleFunc("/configs/{id}", a.wrapFunc(a.adminAPIsHandler.getConfig, a.auth.admin.Permissions)).Methods("GET")
	adminRouter.HandleFunc("/configs", a.wrapFunc(a.adminAPIsHandler.getConfigs, a.auth.admin.Permissions)).Methods("GET")
	adminRouter.HandleFunc("/configs", a.wrapFunc(a.adminAPIsHandler.createConfig, a.auth.admin.Permissions)).Methods("POST")
	adminRouter.HandleFunc("/configs/{id}", a.wrapFunc(a.adminAPIsHandler.updateConfig, a.auth.admin.Permissions)).Methods("PUT")
	adminRouter.HandleFunc("/configs/{id}", a.wrapFunc(a.adminAPIsHandler.deleteConfig, a.auth.admin.Permissions)).Methods("DELETE")

	adminRouter.HandleFunc("/webtools-blacklist", a.wrapFunc(a.adminAPIsHandler.addwebtoolsblacklist, a.auth.admin.Permissions)).Methods("PUT")
	adminRouter.HandleFunc("/webtools-blacklist", a.wrapFunc(a.adminAPIsHandler.getwebtoolsblacklist, a.auth.admin.Permissions)).Methods("GET")
	adminRouter.HandleFunc("/webtools-blacklist", a.wrapFunc(a.adminAPIsHandler.removewebtoolsblacklist, a.auth.admin.Permissions)).Methods("DELETE")
	adminRouter.HandleFunc("/webtools-summary", a.wrapFunc(a.adminAPIsHandler.getWebtoolsSummary, a.auth.admin.Permissions)).Methods("GET")
	//adminRouter.HandleFunc("/legacy-events", a.wrapFunc(a.adminAPIsHandler.legacyEvents, a.auth.admin.Permissions)).Methods("GET")

	// BB APIs
	bbsRouter := mainRouter.PathPrefix("/bbs").Subrouter()
	bbsRouter.HandleFunc("/examples/{id}", a.wrapFunc(a.bbsAPIsHandler.getExample, a.auth.bbs.Permissions)).Methods("GET")
	bbsRouter.HandleFunc("/appointments/units", a.wrapFunc(a.bbsAPIsHandler.getAppointmentUnits, a.auth.bbs.Permissions)).Methods("GET")
	bbsRouter.HandleFunc("/appointments/people", a.wrapFunc(a.bbsAPIsHandler.getAppointmentPeople, a.auth.bbs.Permissions)).Methods("GET")
	bbsRouter.HandleFunc("/appointments/slots", a.wrapFunc(a.bbsAPIsHandler.getAppointmentTimeSlots, a.auth.bbs.Permissions)).Methods("GET")
	bbsRouter.HandleFunc("/appointments/questions", a.wrapFunc(a.bbsAPIsHandler.getAppointmentQuestions, a.auth.bbs.Permissions)).Methods("GET")
	bbsRouter.HandleFunc("/appointments/qands", a.wrapFunc(a.bbsAPIsHandler.getAppointmentOptions, a.auth.bbs.Permissions)).Methods("GET")
	bbsRouter.HandleFunc("/appointments/", a.wrapFunc(a.bbsAPIsHandler.createAppointment, a.auth.bbs.Permissions)).Methods("POST")
	bbsRouter.HandleFunc("/appointments/{id}", a.wrapFunc(a.bbsAPIsHandler.deleteAppointment, a.auth.bbs.Permissions)).Methods("DELETE")
	bbsRouter.HandleFunc("/appointments/", a.wrapFunc(a.bbsAPIsHandler.updateAppointment, a.auth.bbs.Permissions)).Methods("PUT")

	//use api key!!!
	bbsRouter.HandleFunc("/events", a.wrapFunc(a.apiKeyHandler.getLegacyEvents, a.auth.apiKey)).Methods("GET")

	// TPS APIs
	tpsRouter := mainRouter.PathPrefix("/tps").Subrouter()
	tpsRouter.HandleFunc("/examples/{id}", a.wrapFunc(a.tpsAPIsHandler.getExample, a.auth.tps.Permissions)).Methods("GET")
	tpsRouter.HandleFunc("/events", a.wrapFunc(a.tpsAPIsHandler.createEvents, a.auth.tps.Permissions)).Methods("POST")
	tpsRouter.HandleFunc("/events", a.wrapFunc(a.tpsAPIsHandler.deleteEvents, a.auth.tps.Permissions)).Methods("DELETE")

	// System APIs
	systemRouter := mainRouter.PathPrefix("/system").Subrouter()
	systemRouter.HandleFunc("/examples/{id}", a.wrapFunc(a.systemAPIsHandler.getExample, a.auth.system.Permissions)).Methods("GET")

	a.logger.Fatalf("Error serving: %v", http.ListenAndServe(":"+a.port, router))
}

func (a Adapter) serveDoc(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("access-control-allow-origin", "*")

	if a.cachedYamlDoc != nil {
		http.ServeContent(w, r, "", time.Now(), bytes.NewReader([]byte(a.cachedYamlDoc)))
	} else {
		http.ServeFile(w, r, "./driver/web/docs/gen/def.yaml")
	}
}

func (a Adapter) serveDocUI() http.Handler {
	url := fmt.Sprintf("%s/doc", a.baseURL)
	return httpSwagger.Handler(httpSwagger.URL(url))
}

func loadDocsYAML(baseServerURL string) ([]byte, error) {
	data, _ := os.ReadFile("./driver/web/docs/gen/def.yaml")
	yamlMap := yaml.MapSlice{}
	err := yaml.Unmarshal(data, &yamlMap)
	if err != nil {
		return nil, err
	}

	for index, item := range yamlMap {
		if item.Key == "servers" {
			var serverList []interface{}
			if baseServerURL != "" {
				serverList = []interface{}{yaml.MapSlice{yaml.MapItem{Key: "url", Value: baseServerURL}}}
			}

			item.Value = serverList
			yamlMap[index] = item
			break
		}
	}

	yamlDoc, err := yaml.Marshal(&yamlMap)
	if err != nil {
		return nil, err
	}

	return yamlDoc, nil
}

func (a Adapter) wrapFunc(handler handlerFunc, authorization tokenauth.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		logObj := a.logger.NewRequestLog(req)

		logObj.RequestReceived()

		var response logs.HTTPResponse
		if authorization != nil {
			responseStatus, claims, err := authorization.Check(req)
			if err != nil {
				logObj.SendHTTPResponse(w, logObj.HTTPResponseErrorAction(logutils.ActionValidate, logutils.TypeRequest, nil, err, responseStatus, true))
				return
			}

			if claims != nil {
				logObj.SetContext("account_id", claims.Subject)
			}
			response = handler(logObj, req, claims)
		} else {
			response = handler(logObj, req, nil)
		}

		logObj.SendHTTPResponse(w, response)
		logObj.RequestComplete()
	}
}

// NewWebAdapter creates new WebAdapter instance
func NewWebAdapter(baseURL string, port string, serviceID string, apiKey string, app *core.Application, serviceRegManager *authservice.ServiceRegManager, serviceAccountManager *authservice.ServiceAccountManager, logger *logs.Logger) Adapter {
	yamlDoc, err := loadDocsYAML(baseURL)
	if err != nil {
		logger.Fatalf("error parsing docs yaml - %s", err.Error())
	}

	auth, err := NewAuth(serviceRegManager, apiKey)
	if err != nil {
		logger.Fatalf("error creating auth - %s", err.Error())
	}

	defaultAPIsHandler := NewDefaultAPIsHandler(app)
	clientAPIsHandler := NewClientAPIsHandler(app)
	adminAPIsHandler := NewAdminAPIsHandler(app)
	bbsAPIsHandler := NewBBsAPIsHandler(app, serviceAccountManager)
	tpsAPIsHandler := NewTPSAPIsHandler(app)
	apiKeyHandler := NewAPIKeyHandler(app)
	return Adapter{baseURL: baseURL, port: port, serviceID: serviceID, cachedYamlDoc: yamlDoc, auth: auth, defaultAPIsHandler: defaultAPIsHandler,
		clientAPIsHandler: clientAPIsHandler, adminAPIsHandler: adminAPIsHandler, bbsAPIsHandler: bbsAPIsHandler,
		tpsAPIsHandler: tpsAPIsHandler, app: app, apiKeyHandler: apiKeyHandler, logger: logger}
}
