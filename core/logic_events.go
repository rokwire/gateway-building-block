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
	"application/driven/storage"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/rokwire/logging-library-go/v2/logs"
)

type eventsLogic struct {
	app    *Application
	logger logs.Logger

	eventsBBAdapter EventsBBAdapter
}

func (e eventsLogic) start() error {

	//1. check if the initial import must be applied - it happens only once!
	err := e.importInitialEventsFromEventsBB()
	if err != nil {
		return err
	}

	//hold on for now.. use timer
	//events, _ := e.getAllEvents()
	//fmt.Println(events)

	return nil
}

func (e eventsLogic) importInitialEventsFromEventsBB() error {
	importProcessed := false

	//in transaction
	err := e.app.storage.PerformTransaction(func(context storage.TransactionContext) error {

		//first check if need to import
		config, err := e.app.storage.FindGlobalConfig(context, "initial-legacy-events-import")
		if err != nil {
			return err
		}
		if config == nil {
			return errors.New("no initial legacy events import config added")
		}
		processed := config.Data["processed"].(bool)
		if processed {
			importProcessed = true
			return nil //no need to execute processing
		}

		// we make initial import

		//load the events
		events, err := e.eventsBBAdapter.LoadAllLegacyEvents()
		if err != nil {
			return err
		}

		//they cannot be 0
		eventsCount := len(events)
		if eventsCount == 0 {
			return errors.New("cannot have 0 events, there is an error")
		}

		e.logger.Infof("Got %d events from events BB", eventsCount)

		//prepare the list which we will store
		syncProcessSource := "events-bb-initial"
		now := time.Now()
		resultList := make([]model.LegacyEventItem, eventsCount)
		for i, le := range events {
			leItem := model.LegacyEventItem{SyncProcessSource: syncProcessSource, SyncDate: now, Item: le}
			resultList[i] = leItem
		}

		//insert the initial events
		err = e.app.storage.InsertLegacyEvents(context, resultList)
		if err != nil {
			return err
		}

		//mark as processed
		config.Data["processed"] = true
		err = e.app.storage.SaveGlobalConfig(context, *config)
		if err != nil {
			return err
		}

		return nil
	}, 60000)

	if err != nil {
		return err
	}

	if importProcessed {
		e.logger.Info("Initial events already imported")
	} else {
		e.logger.Info("Successfuly imported initial events")
	}
	return nil
}

func (e eventsLogic) getAllEvents() ([]model.WebToolsEventItem, error) {
	var allevents []model.WebToolsEventItem
	var events []model.WebToolsEventItem
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

		var parsed model.WebToolsEventItem
		err = xml.Unmarshal(data, &parsed)
		if err != nil {
			log.Printf("error: %s", err)
			break
		}

		count := len(parsed.WebToolsEvent)
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
		if w.WebToolsEvent != nil {
			for _, g := range w.WebToolsEvent {

				event := e.constructLegacyEvents(g)
				legacyEvent = append(legacyEvent, event)
			}
		}
	}

	//le := e.app.storage.SaveLegacyEvents(legacyEvent)
	//fmt.Println(le)

	return allevents, nil
}

func (e eventsLogic) constructLegacyEvents(g model.WebToolsEvent) model.LegacyEvent {

	// For Stefan:
	//apply the needed processing so that to convert the web tools event in legacy event 100% how it happens in the old code!

	/*
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

		event := model.LegacyEvent{RecurringFlag: Recurrence, /* RegistrationURL: &g.RegistrationURL,  TitleURL: &g.TitleURL*/
	//	LongDescription: g.Description, IsEventFree: IsEventFree, AllDay: false, SourceID: "0",
	//	CalendarID: g.CalendarID, Title: g.Title, Sponsor: g.Sponsor, DataSourceEventID: g.EventID,
	//	IsVirtial: isVirtual, OriginatingCalendarID: g.OriginatingCalendarID, Category: g.EventType} */

	return model.LegacyEvent{}
}

// newAppEventsLogic creates new appShared
func newAppEventsLogic(app *Application, eventsBBAdapter EventsBBAdapter, logger logs.Logger) eventsLogic {
	return eventsLogic{app: app, eventsBBAdapter: eventsBBAdapter, logger: logger}
}
