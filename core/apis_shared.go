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
)

// appShared contains shared implementations
type appShared struct {
	app *Application
}

// getExample gets an Example by ID
func (a appShared) getExample(orgID string, appID string, id string) (*model.Example, error) {
	return a.app.storage.FindExample(orgID, appID, id)
}

// newAppShared creates new appShared
func newAppShared(app *Application) appShared {
	return appShared{app: app}
}

// getBuildingFeatures returns all building features
func (a appShared) getBuildingFeatures() ([]model.AppBuildingFeature, error) {
	return a.app.storage.LoadAppBuildingFeatures()
}

// getFloorPlanMarkup returns the floor plan markup needed by the app
func (a appShared) getFloorPlanMarkup() (*model.FloorPlanMarkup, error) {
	return a.app.storage.LoadFloorPlanMarkup()
}
