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
	"application/core/model"
	"time"

	"github.com/rokwire/rokwire-building-block-sdk-go/utils/errors"
	"github.com/rokwire/rokwire-building-block-sdk-go/utils/logging/logs"
	"github.com/rokwire/rokwire-building-block-sdk-go/utils/logging/logutils"
	"github.com/rokwire/rokwire-building-block-sdk-go/utils/rokwireutils"
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

	Default Default // expose to the drivers adapters
	Client  Client  // expose to the drivers adapters
	Admin   Admin   // expose to the drivers adapters
	BBs     BBs     // expose to the drivers adapters
	TPS     TPS     // expose to the drivers adapters
	System  System  // expose to the drivers adapters
	shared  Shared

	CampusBuildings  model.CachedBuildings               //caches a list of all campus building data
	AppBLdgFeatures  map[string]model.AppBuildingFeature //caches the configured set of building features
	FloorPlanWrapper model.FloorPlanMarkup               //caches the floor plan markup

	AppointmentAdapters map[string]Appointments //expose to the different vendor specific appointment adapters

	logger *logs.Logger

	storage Storage

	eventsBBAdapter EventsBBAdapter
	imageAdapter    ImageAdapter
	geoBBAdapter    GeoAdapter

	//events logic
	eventsLogic eventsLogic
}

// Start starts the core part of the application
func (a *Application) Start() error {
	//set storage listener
	storageListener := storageListener{app: a}
	a.storage.RegisterStorageListener(&storageListener)

	err := a.eventsLogic.start()
	if err != nil {
		return err
	}

	//no error
	return nil
}

// GetEnvConfigs retrieves the cached database env configs
func (a *Application) GetEnvConfigs() (*model.EnvConfigData, error) {
	// Load env configs from database
	config, err := a.storage.FindConfig(model.ConfigTypeEnv, rokwireutils.AllApps, rokwireutils.AllOrgs)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionGet, model.TypeConfig, nil, err)
	}
	if config == nil {
		return nil, errors.ErrorData(logutils.StatusMissing, model.TypeConfig, &logutils.FieldArgs{"type": model.ConfigTypeEnv, "app_id": rokwireutils.AllApps, "org_id": rokwireutils.AllOrgs})
	}
	return model.GetConfigData[model.EnvConfigData](*config)
}

// NewApplication creates new Application
func NewApplication(version string, build string,
	storage Storage,
	eventsBBAdapter EventsBBAdapter,
	imageAdapter ImageAdapter,
	geoBBAdapter GeoAdapter,
	appntAdapters map[string]Appointments,
	logger *logs.Logger) *Application {
	application := Application{version: version, build: build, storage: storage, eventsBBAdapter: eventsBBAdapter, imageAdapter: imageAdapter, logger: logger, AppointmentAdapters: appntAdapters}

	//add the drivers ports/interfaces
	application.Default = newAppDefault(&application)
	application.Client = newAppClient(&application)
	application.Admin = newAppAdmin(&application)
	application.BBs = newAppBBs(&application)
	application.TPS = newAppTPS(&application)
	application.System = newAppSystem(&application)
	application.shared = newAppShared(&application)
	application.eventsLogic = newAppEventsLogic(&application, eventsBBAdapter, geoBBAdapter, *logger)

	fmpw, fmerr := application.shared.getFloorPlanMarkup()
	if fmerr != nil {
		application.logger.Errorf("Error getting floor plan markup: %v", fmerr)
	} else {
		application.logger.Infof("Floor plan markup loaded successfully")
		application.FloorPlanWrapper = *fmpw
	}

	bldfeatures, blderr := application.shared.getBuildingFeatures()
	if blderr != nil {

	}
	application.AppBLdgFeatures = make(map[string]model.AppBuildingFeature)
	for _, bf := range bldfeatures {
		application.AppBLdgFeatures[bf.CampusCode] = bf
	}

	//	_, err := application.Client.GetBuildings()
	//	if err != nil {
	//set to one day ago to force a retry and refresh
	application.CampusBuildings.LoadDate = time.Now().AddDate(0, 0, -1)
	//}

	return &application
}
