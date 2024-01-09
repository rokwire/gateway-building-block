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

package core

import (
	"application/core/interfaces"
	"application/core/model"

	"github.com/rokwire/core-auth-library-go/v3/authutils"
	"github.com/rokwire/logging-library-go/v2/errors"
	"github.com/rokwire/logging-library-go/v2/logs"
	"github.com/rokwire/logging-library-go/v2/logutils"
)

type storageListener struct {
	app *Application
	model.DefaultStorageListener
}

// OnExampleUpdated notifies that the example collection has changed
func (s *storageListener) OnExampleUpdated() {
	s.app.logger.Infof("OnExampleUpdated")

	// TODO: Implement listener
}

// Application represents the core application code based on hexagonal architecture
type Application struct {
	version string
	build   string

	Default interfaces.Default // expose to the drivers adapters
	Client  interfaces.Client  // expose to the drivers adapters
	Admin   interfaces.Admin   // expose to the drivers adapters
	BBs     interfaces.BBs     // expose to the drivers adapters
	TPS     interfaces.TPS     // expose to the drivers adapters
	System  interfaces.System  // expose to the drivers adapters
	shared  Shared

	AppointmentAdapters map[string]interfaces.Appointments //expose to the different vendor specific appointment adapters

	logger *logs.Logger

	storage interfaces.Storage

	//events logic
	eventsLogic eventsLogic
}

// Start starts the core part of the application
func (a *Application) Start() {
	//set storage listener
	storageListener := storageListener{app: a}
	a.storage.RegisterStorageListener(&storageListener)

	a.eventsLogic.start()
}

// GetEnvConfigs retrieves the cached database env configs
func (a *Application) GetEnvConfigs() (*model.EnvConfigData, error) {
	// Load env configs from database
	config, err := a.storage.FindConfig(model.ConfigTypeEnv, authutils.AllApps, authutils.AllOrgs)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionGet, model.TypeConfig, nil, err)
	}
	if config == nil {
		return nil, errors.ErrorData(logutils.StatusMissing, model.TypeConfig, &logutils.FieldArgs{"type": model.ConfigTypeEnv, "app_id": authutils.AllApps, "org_id": authutils.AllOrgs})
	}
	return model.GetConfigData[model.EnvConfigData](*config)
}

// NewApplication creates new Application
func NewApplication(version string, build string, storage interfaces.Storage, appntAdapters map[string]interfaces.Appointments, logger *logs.Logger) *Application {
	application := Application{version: version, build: build, storage: storage, logger: logger, AppointmentAdapters: appntAdapters}

	//add the drivers ports/interfaces
	application.Default = newAppDefault(&application)
	application.Client = newAppClient(&application)
	application.Admin = newAppAdmin(&application)
	application.BBs = newAppBBs(&application)
	application.TPS = newAppTPS(&application)
	application.System = newAppSystem(&application)
	application.shared = newAppShared(&application)
	application.eventsLogic = newAppEventsLogic(&application)

	return &application
}
