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

package web

import (
	"apigateway/core"
	"apigateway/driver/web/rest"
	"apigateway/utils"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"
)

//Adapter entity
type Adapter struct {
	host string
	port string

	apisHandler       rest.ApisHandler
	adminApisHandler  rest.AdminApisHandler
	laundryapiHandler rest.LaundryApisHandler
	tokenAuth         *TokenAuth
	app               *core.Application
}

// @title Rokwire Gatewauy Building Block API
// @description Rokwire Rokwire Building Block API Documentation.
// @version 0.1.0
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @host localhost
// @BasePath /notifications/api
// @schemes https

// @securityDefinitions.apikey RokwireAuth
// @in header
// @name ROKWIRE-API-KEY

// @securityDefinitions.apikey InternalAuth
// @in header
// @name INTERNAL-API-KEY

// @securityDefinitions.apikey UserAuth
// @in header (add client id token with Bearer prefix to the Authorization value)
// @name Authorization

// @securityDefinitions.apikey AdminUserAuth
// @in header (add admin id token with Bearer prefix to the Authorization value)
// @name Authorization

// Start starts the module
func (we Adapter) Start() {

	router := mux.NewRouter().StrictSlash(true)

	// handle apis
	//do i need a different adapter for each "endpoint" (laundry, courselist, wayfinding, etc)
	//or can I set different routers for different router path prefixise (/laundry, /courselist, ...)
	//still learning the gorilla mux library
	mainRouter := router.PathPrefix("/laundry/api").Subrouter()
	mainRouter.PathPrefix("/doc/ui").Handler(we.serveDocUI())
	mainRouter.HandleFunc("/doc", we.serveDoc)
	mainRouter.HandleFunc("/version", we.wrapFunc(we.apisHandler.Version)).Methods("GET")

	// Client APIs
	mainRouter.HandleFunc("/record", we.tokenAuthWrapFunc(we.apisHandler.StoreRecord)).Methods("POST")
	mainRouter.HandleFunc("/rooms", we.wrapFunc(we.laundryapiHandler.GetLaundryRooms)).Methods("GET")
	mainRouter.HandleFunc("/room", we.wrapFunc(we.laundryapiHandler.GetRoomDetails)).Methods("GET")
	mainRouter.HandleFunc("/initrequest", we.wrapFunc(we.laundryapiHandler.InitServiceRequest)).Methods("GET")
	mainRouter.HandleFunc("/requestservice", we.wrapFunc(we.laundryapiHandler.SubmitServiceRequest)).Methods("POST")

	log.Fatal(http.ListenAndServe(":"+we.port, router))
}

func (we Adapter) serveDoc(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("access-control-allow-origin", "*")
	http.ServeFile(w, r, "./docs/swagger.yaml")
}

func (we Adapter) serveDocUI() http.Handler {
	url := fmt.Sprintf("%s/notifications/api/doc", we.host)
	return httpSwagger.Handler(httpSwagger.URL(url))
}

//functions with no authentication at all
func (we Adapter) wrapFunc(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		utils.LogRequest(req)

		handler(w, req)
	}
}

func (we Adapter) tokenAuthWrapFunc(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		//authenticate token
		claims, err := we.tokenAuth.tokenAuth.CheckRequestTokens(req)
		if err != nil {
			log.Printf("Authentication error: %v\n", err)
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		err = we.tokenAuth.tokenAuth.AuthorizeRequestScope(claims, req)
		if err != nil {
			log.Printf("Scope error: %v\n", err)
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}
		log.Printf("Authentication successful for user: %v", claims)
		handler(w, req)
	}
}

//NewWebAdapter creates new WebAdapter instance
func NewWebAdapter(host string, port string, app *core.Application, tokenAuth *TokenAuth) Adapter {

	apisHandler := rest.NewApisHandler(app)
	adminApisHandler := rest.NewAdminApisHandler(app)
	laundryapiHandler := rest.NewLaundryApisHandler(app)
	return Adapter{host: host, port: port,
		apisHandler: apisHandler, adminApisHandler: adminApisHandler, app: app, laundryapiHandler: laundryapiHandler, tokenAuth: tokenAuth}
}

//AppListener implements core.ApplicationListener interface
type AppListener struct {
	adapter *Adapter
}
