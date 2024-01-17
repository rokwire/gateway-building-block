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

package main

import (
	"application/core"
	"application/core/interfaces"
	"application/driven/eventsbb"
	"application/driven/storage"
	"application/driven/uiucadapters"
	"application/driver/web"
	"strings"

	"github.com/rokwire/core-auth-library-go/v3/authservice"
	"github.com/rokwire/core-auth-library-go/v3/envloader"
	"github.com/rokwire/logging-library-go/v2/logs"
)

var (
	// Version : version of this executable
	Version string
	// Build : build date of this executable
	Build string
)

func main() {
	if len(Version) == 0 {
		Version = "dev"
	}

	serviceID := "gateway"

	loggerOpts := logs.LoggerOpts{
		SensitiveHeaders: []string{"External-Authorization"},
		SuppressRequests: logs.NewStandardHealthCheckHTTPRequestProperties(serviceID + "/version")}
	logger := logs.NewLogger(serviceID, &loggerOpts)
	envLoader := envloader.NewEnvLoader(Version, logger)

	envPrefix := strings.ReplaceAll(strings.ToUpper(serviceID), "-", "_") + "_"
	port := envLoader.GetAndLogEnvVar(envPrefix+"PORT", false, false)
	if len(port) == 0 {
		port = "80"
	}

	// mongoDB adapter
	mongoDBAuth := envLoader.GetAndLogEnvVar(envPrefix+"MONGO_AUTH", true, true)
	mongoDBName := envLoader.GetAndLogEnvVar(envPrefix+"MONGO_DATABASE", true, false)
	mongoTimeout := envLoader.GetAndLogEnvVar(envPrefix+"MONGO_TIMEOUT", false, false)
	storageAdapter := storage.NewStorageAdapter(mongoDBAuth, mongoDBName, mongoTimeout, logger)
	err := storageAdapter.Start()
	if err != nil {
		logger.Fatalf("Cannot start the mongoDB adapter: %v", err)
	}

	// events bb adapter
	eventsBBBaseURL := envLoader.GetAndLogEnvVar(envPrefix+"EVENTS_BB_BASE_URL", true, true)
	eventsBBAPIKey := envLoader.GetAndLogEnvVar(envPrefix+"EVENTS_BB_ROKWIRE_API_KEY", true, true)
	eventsBBAdapter := eventsbb.NewEventsBBAdapter(eventsBBBaseURL, eventsBBAPIKey, logger)

	// appointment adapters
	appointments := make(map[string]interfaces.Appointments)
	appointments["2"] = uiucadapters.NewEngineeringAppontmentsAdapter("KP")
	// application
	application := core.NewApplication(Version, Build, storageAdapter, eventsBBAdapter, appointments, logger)
	err = application.Start()
	if err != nil {
		logger.Fatalf("Cannot start the Application module: %v", err)
	}

	// web adapter
	baseURL := envLoader.GetAndLogEnvVar(envPrefix+"BASE_URL", true, false)
	coreBBBaseURL := envLoader.GetAndLogEnvVar(envPrefix+"CORE_BB_BASE_URL", true, false)

	authService := authservice.AuthService{
		ServiceID:   serviceID,
		ServiceHost: baseURL,
		FirstParty:  true,
		AuthBaseURL: coreBBBaseURL,
	}

	serviceRegLoader, err := authservice.NewRemoteServiceRegLoader(&authService, nil)
	if err != nil {
		logger.Fatalf("Error initializing remote service registration loader: %v", err)
	}

	serviceRegManager, err := authservice.NewServiceRegManager(&authService, serviceRegLoader, !strings.HasPrefix(baseURL, "http://localhost"))
	if err != nil {
		logger.Fatalf("Error initializing service registration manager: %v", err)
	}

	webAdapter := web.NewWebAdapter(baseURL, port, serviceID, application, serviceRegManager, logger)
	webAdapter.Start()
}
