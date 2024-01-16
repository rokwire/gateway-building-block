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

package core

import (
	"application/core/model"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/rokwire/logging-library-go/v2/logs"
)

type eventsLogic struct {
	app    *Application
	logger *logs.Logger
}

func (e eventsLogic) start() {
	events, _ := e.getAllEvents()

	fmt.Println(events)
}

func (e eventsLogic) getAllEvents() ([]model.ResponseWS, error) {
	var allevents []model.ResponseWS
	var events []model.ResponseWS
	var legacyEvent []model.LegacyEvent
	page := 0
	for {
		resp, err := http.Get(fmt.Sprintf("https://xml.calendars.illinois.edu/eventXML16/6991.xml?pageNumber=%d", page))
		if err != nil {
			log.Printf("error: %s", err)
			break
		}

		data, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("error: %s", err)
			break
		}

		var parsed model.ResponseWS
		err = xml.Unmarshal(data, &parsed)
		if err != nil {
			log.Printf("error: %s", err)
			break
		}

		count := len(parsed.PublicEventWS)
		log.Printf("Events count: %d", count)

		//io.Write(file)
		if count == 0 {
			break
		}
		page++
		events = append(events, parsed)
		allevents = append(allevents, events...)
	}

	for _, w := range allevents {
		if w.PublicEventWS != nil {
			for _, g := range w.PublicEventWS {

				var isVirtual bool
				if g.VirtualEvent == "false" {
					isVirtual = false
				} else if g.VirtualEvent == "true" {
					isVirtual = true
				}

				var IsEventFree bool
				if g.CostFree == "false" {
					IsEventFree = false
				} else if g.CostFree == "true" {
					IsEventFree = true
				}

				var Recurrence bool
				if g.Recurrence == "false" {
					Recurrence = false
				} else if g.Recurrence == "true" {
					Recurrence = true
				}

				event := model.LegacyEvent{RecurringFlag: Recurrence, RegistrationURL: &g.RegistrationURL, TitleURL: &g.TitleURL,
					LongDescription: g.Description, IsEventFree: IsEventFree, AllDay: false, SourceID: "0",
					CalendarID: g.CalendarId, Title: g.Title, Sponsor: g.Sponsor, DataSourceEventID: g.EventId,
					IsVirtial: isVirtual, OriginatingCalendarID: g.OriginatingCalendarId, Category: g.EventType}
				legacyEvent = append(legacyEvent, event)
			}
		}
	}

	le := e.app.storage.SaveLegacyEvents(legacyEvent)
	fmt.Println(le)

	return allevents, nil
}

// newAppEventsLogic creates new appShared
func newAppEventsLogic(app *Application) eventsLogic {
	return eventsLogic{app: app}
}
