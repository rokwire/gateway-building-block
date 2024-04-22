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
	"strings"
)

// appTPS contains BB implementations
type appTPS struct {
	app *Application
}

// GetExample gets an Example by ID
func (a appTPS) GetExample(orgID string, appID string, id string) (*model.Example, error) {
	return a.app.shared.getExample(orgID, appID, id)
}

// CreateEvents creates events
func (a appTPS) CreateEvents(event []model.LegacyEventItem) ([]model.LegacyEventItem, error) {
	modifiedLegacyEvents, err := a.modifyLegacyEventsList(event)
	if err != nil {
		a.app.logger.Errorf("error on ignoring legacy events - %s", err)
		return nil, err
	}
	return a.app.storage.InsertLegacyEvents(nil, modifiedLegacyEvents)
}

// DeleteEvents deletes legacy events by ids and creator
func (a appTPS) DeleteEvents(ids []string, accountID string) error {
	return a.app.storage.DeleteLegacyEventsByIDsAndCreator(nil, ids, accountID)
}

// ignore or modify legacy events
func (a appTPS) modifyLegacyEventsList(legacyEvents []model.LegacyEventItem) ([]model.LegacyEventItem, error) {
	modifiedList := []model.LegacyEventItem{}
	modified := 0

	//map for category conversions
	categoryMap := map[string]string{
		"exhibition":               "Exhibits",
		"festival/celebration":     "Festivals and Celebrations",
		"film screening":           "Film Screenings",
		"performance":              "Performances",
		"lecture":                  "Speakers and Seminars",
		"seminar/symposium":        "Speakers and Seminars",
		"conference/workshop":      "Conferences and Workshops",
		"reception/open house":     "Receptions and Open House Events",
		"social/informal event":    "Social and Informal Events",
		"professional development": "Career Development",
		"health/fitness":           "Recreation, Health and Fitness",
		"sporting event":           "Club Athletics",
		"sidearm":                  "Big 10 Athletics",
	}

	for _, wte := range legacyEvents {
		currentWte := wte
		category := currentWte.Item.Category
		lowerCategory := strings.ToLower(category)

		//modify some categories
		if newCategory, ok := categoryMap[lowerCategory]; ok {
			currentWte.Item.Category = newCategory
			a.app.logger.Infof("modifying event category from %s to %s", category, newCategory)

			modified++
		}

		//add it to the modified list
		modifiedList = append(modifiedList, currentWte)
	}
	a.app.logger.Infof("events count is %d", modified)
	a.app.logger.Infof("final list is %d", len(modifiedList))

	return modifiedList, nil
}

// newAppTPS creates new appTPS
func newAppTPS(app *Application) appTPS {
	return appTPS{app: app}
}
