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

// Application represents the core application code based on hexagonal architecture
type Application struct {
	version string
	build   string

	Services Services // expose to the drivers adapters

	storage            Storage
	laundry            Laundry
	locationAdapter    BuildingLocation
	contactInfoAdapter ContactInformation
	giesCourseAdapter  GiesCourses
}

// Start starts the core part of the application
func (app *Application) Start() {
}

// NewApplication creates new Application
func NewApplication(version string, build string, storage Storage, laundry Laundry, bldgloc BuildingLocation, contacts ContactInformation, gcAdapter GiesCourses) *Application {

	application := Application{version: version, build: build, storage: storage, laundry: laundry, locationAdapter: bldgloc, contactInfoAdapter: contacts, giesCourseAdapter: gcAdapter}

	//add the drivers ports/interfaces
	application.Services = &servicesImpl{app: &application}

	return &application
}
