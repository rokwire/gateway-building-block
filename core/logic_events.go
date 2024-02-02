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
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/rokwire/logging-library-go/v2/logs"
)

type eventsLogic struct {
	app    *Application
	logger logs.Logger

	eventsBBAdapter EventsBBAdapter

	//web tools timer
	dailyWebToolsTimer *time.Timer
	timerDone          chan bool
}

func (e eventsLogic) start() error {

	//1. check if the initial import must be applied - it happens only once!
	err := e.importInitialEventsFromEventsBB()
	if err != nil {
		return err
	}

	//2. set up web tools timer
	go e.setupWebToolsTimer()

	//3. initialize event locations db if needs
	go e.initializeDB()

	return nil
}

func (e eventsLogic) initializeDB() {
	e.logger.Info("InitializeLegacyLocations started")
	defer e.logger.Info("InitializeLegacyLocations ended")
	err := e.app.storage.InitializeLegacyLocations()
	if err != nil {
		e.logger.Errorf("error on initialzing legacy locations db: %s", err)
	}
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

		//there are a lot of duplicate items(dataSourceEventId), so we need to fix them
		fixedEvents := []model.LegacyEvent{}
		addedItemMap := map[string]bool{}
		for _, item := range events {
			if len(item.DataSourceEventID) == 0 {
				//there are a lot of such items absiously they are used, so add them
				fixedEvents = append(fixedEvents, item)
			} else {
				if _, exists := addedItemMap[item.DataSourceEventID]; exists {
					e.logger.Infof("Already added %s, so do nothing", item.DataSourceEventID)
				} else {
					//not added, so adding it
					fixedEvents = append(fixedEvents, item)

					//mark it as added
					addedItemMap[item.DataSourceEventID] = true
				}
			}
		}

		e.logger.Infof("Got %d events after the events fix", len(fixedEvents))

		//prepare the list which we will store
		syncProcessSource := "events-bb-initial"
		now := time.Now()
		resultList := make([]model.LegacyEventItem, len(fixedEvents))
		for i, le := range fixedEvents {
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

func (e eventsLogic) setupWebToolsTimer() {
	log.Println("Web tools timer")

	//cancel if active
	if e.dailyWebToolsTimer != nil {
		log.Println("setupWebToolsTimer -> there is active timer, so cancel it")

		e.timerDone <- true
		e.dailyWebToolsTimer.Stop()
	}
	/*
		//wait until it is the correct moment from the day
		location, err := time.LoadLocation("America/Chicago")
		if err != nil {
			log.Printf("Error getting location:%s\n", err.Error())
		}
		now := time.Now().In(location)
		log.Printf("setupWebToolsTimer -> now - hours:%d minutes:%d seconds:%d\n", now.Hour(), now.Minute(), now.Second())

		nowSecondsInDay := 60*60*now.Hour() + 60*now.Minute() + now.Second()
		desiredMoment := 18000

		var durationInSeconds int
		log.Printf("setupWebToolsTimer -> nowSecondsInDay:%d desiredMoment:%d\n", nowSecondsInDay, desiredMoment)
		if nowSecondsInDay <= desiredMoment {
			log.Println("setupWebToolsTimer -> not web tools process today, so the first process will be today")
			durationInSeconds = desiredMoment - nowSecondsInDay
		} else {
			log.Println("setupWebToolsTimer -> the web tools process has already been processed today, so the first process will be tomorrow")
			leftToday := 86400 - nowSecondsInDay
			durationInSeconds = leftToday + desiredMoment // the time which left today + desired moment from tomorrow
		}*/
	//log.Println(durationInSeconds)
	duration := time.Second * time.Duration(0)
	//duration := time.Second * time.Duration(durationInSeconds)
	log.Printf("setupWebToolsTimer -> first call after %s", duration)

	e.dailyWebToolsTimer = time.NewTimer(duration)
	select {
	case <-e.dailyWebToolsTimer.C:
		log.Println("setupWebToolsTimer -> web tools timer expired")
		e.dailyWebToolsTimer = nil

		e.processWebToolsEvents()
	case <-e.timerDone:
		// timer aborted
		log.Println("setupWebToolsTimer -> web tools timer aborted")
		e.dailyWebToolsTimer = nil
	}
}

func (e eventsLogic) processWebToolsEvents() {
	//load all web tools events
	allWebToolsEvents, err := e.loadAllWebToolsEvents()
	if err != nil {
		e.logger.Errorf("error on loading web tools events - %s", err)
		return
	}
	webToolsCount := len(allWebToolsEvents)
	if webToolsCount == 0 {
		e.logger.Error("web tools are nil")
		return
	}

	e.logger.Infof("we loaded %d web tools events", webToolsCount)

	now := time.Now()

	//in transaction
	err = e.app.storage.PerformTransaction(func(context storage.TransactionContext) error {
		//1. first find which events are already in the database. You have to compare by dataSourceEventId field.
		legacyEventItemFromStorage, err := e.app.storage.FindLegacyEventItems(context)
		if err != nil {
			e.logger.Errorf("error on loading events from the storage - %s", err)
			return nil
		}

		var leExist []model.LegacyEventItem
		for _, w := range allWebToolsEvents {
			for _, l := range legacyEventItemFromStorage {
				if w.EventID == l.Item.DataSourceEventID {
					leExist = append(leExist, l)
				}
			}
		}

		//1.1 before to execute point 2(i.e. remove all of them) you must keep their IDs so that to put them again on point 3
		existingLegacyIdsMap := make(map[string]string)
		for _, w := range leExist {
			if w.Item.DataSourceEventID != "" {
				existingLegacyIdsMap[w.Item.DataSourceEventID] = w.Item.ID
			}
		}

		//2. Once you know which are already in the datatabse then you must remove all of them
		err = e.app.storage.DeleteLegacyEventsByIDs(context, existingLegacyIdsMap)
		if err != nil {
			e.logger.Errorf("error on deleting events from the storage - %s", err)
			return nil
		}

		//3. Now you have to convert all allWebToolsEvents into legacy events
		newLegacyEvents := []model.LegacyEventItem{}
		for _, wt := range allWebToolsEvents {

			//prepare the id
			id := e.prepareID(wt.EventID, existingLegacyIdsMap)

			le := e.constructLegacyEvent(wt, id, now)
			newLegacyEvents = append(newLegacyEvents, le)
		}

		//4. Store all them in the database
		err = e.app.storage.InsertLegacyEvents(context, newLegacyEvents)
		if err != nil {
			e.logger.Errorf("error on saving events to the storage - %s", err)
			return nil
		}
		// It is all!

		//* keep the already exisiting events IDS THE SAME!

		return nil
	}, 60000)

	if err != nil {
		e.logger.Errorf("error performing transaction - %s", err)
		return
	}
}

func (e eventsLogic) prepareID(currentWTEventID string, existingLegacyIdsMap map[string]string) string {
	if value, exists := existingLegacyIdsMap[currentWTEventID]; exists {
		return value
	}
	return uuid.NewString()
}

func (e eventsLogic) loadAllWebToolsEvents() ([]model.WebToolsEvent, error) {
	allWebToolsEvents := []model.WebToolsEvent{}

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

		var responseData model.WebToolsResponse
		err = xml.Unmarshal(data, &responseData)
		if err != nil {
			log.Printf("error: %s", err)
			break
		}

		count := len(responseData.WebToolsEvents)
		log.Printf("page:%d events count: %d", page, count)

		//io.Write(file)
		if count == 0 {
			break
		}
		page++

		currentItems := responseData.WebToolsEvents
		allWebToolsEvents = append(allWebToolsEvents, currentItems...)
	}

	return allWebToolsEvents, nil
}

func (e eventsLogic) constructLegacyEvent(g model.WebToolsEvent, id string, now time.Time) model.LegacyEventItem {
	syncProcessSource := "webtools-direct"

	var costFree bool
	if g.CostFree == "false" {
		costFree = false
	} else if g.CostFree == "true" {
		costFree = true
	}

	var isVirtual bool
	if g.VirtualEvent == "false" {
		isVirtual = false
	} else if g.VirtualEvent == "true" {
		isVirtual = true
	}

	var Recurrence bool
	if g.Recurrence == "false" {
		Recurrence = false
	} else if g.Recurrence == "true" {
		Recurrence = true
	}
	icalURL := fmt.Sprintf("https://calendars.illinois.edu/ical/%s/%s.ics", g.CalendarID, g.EventID)
	outlookURL := fmt.Sprintf("https://calendars.illinois.edu/outlook2010/%s/%s.ics", g.CalendarID, g.EventID)

	recurrenceID, _ := recurenceIDtoInt(g.RecurrenceID)
	location := locationToDef(g.Location)
	con := model.ContactLegacy{ContactName: g.CalendarName, ContactEmail: g.ContactEmail, ContactPhone: g.ContactName}
	var contacts []model.ContactLegacy
	contacts = append(contacts, con)
	contatsLegacy := contactsToDef(contacts)

	return model.LegacyEventItem{SyncProcessSource: syncProcessSource, SyncDate: now,
		Item: model.LegacyEvent{ID: id, Category: g.EventType, OriginatingCalendarID: g.OriginatingCalendarID, IsVirtial: isVirtual, DataModified: g.EventID,
			Sponsor: g.Sponsor, Title: g.Title, CalendarID: g.CalendarID, SourceID: "0", AllDay: false, IsEventFree: costFree, LongDescription: g.Description,
			TitleURL: g.TitleURL, RegistrationURL: g.RegistrationURL, RecurringFlag: Recurrence, IcalURL: icalURL, OutlookURL: outlookURL,
			RecurrenceID: recurrenceID, Location: &location, Contacts: contatsLegacy,
			DataSourceEventID: g.EventID, StartDate: g.StartDate}}
}

func (e eventsLogic) consLegacyEventItem(g model.LegacyEventItem) model.LegacyEventItem {

	return model.LegacyEventItem{SyncProcessSource: g.SyncProcessSource, SyncDate: g.SyncDate,
		Item: model.LegacyEvent{AllDay: g.Item.AllDay, CalendarID: g.Item.CalendarID, Category: g.Item.Category,
			Subcategory: g.Item.Subcategory, CreatedBy: g.Item.CreatedBy, LongDescription: g.Item.LongDescription,
			DataModified: g.Item.DataModified, DataSourceEventID: g.Item.DataSourceEventID, DateCreated: g.Item.DateCreated,
			EndDate: g.Item.EndDate, EventID: g.Item.EventID, IcalURL: g.Item.IcalURL, IsEventFree: g.Item.IsEventFree,
			IsVirtial: g.Item.IsVirtial, Location: g.Item.Location, OriginatingCalendarID: g.Item.OriginatingCalendarID,
			OutlookURL: g.Item.OutlookURL, RecurrenceID: g.Item.RecurrenceID, IsSuperEvent: g.Item.IsSuperEvent,
			RecurringFlag: g.Item.RecurringFlag, SourceID: g.Item.SourceID, Sponsor: g.Item.Sponsor, StartDate: g.Item.StartDate,
			Title: g.Item.Title, TitleURL: g.Item.TitleURL, RegistrationURL: g.Item.RegistrationURL, Contacts: g.Item.Contacts, SubEvents: g.Item.SubEvents}}
}
func (e eventsLogic) consLegacyEventsItems(items []model.LegacyEventItem) []model.LegacyEventItem {
	defs := make([]model.LegacyEventItem, len(items))
	for index := range items {
		defs[index] = e.consLegacyEventItem(items[index])
	}
	return defs
}

func recurenceIDtoInt(s string) (*int, error) {
	// Parse string to int
	parsedInt, err := strconv.Atoi(s)
	if err != nil {
		return nil, err
	}

	// Create a pointer to the parsed integer
	result := new(int)
	*result = parsedInt

	return result, nil
}

// Location
func locationToDef(location string) model.LocationLegacy {
	var description string
	var latitude float32
	var longitude float32
	if location != "" {
		description = location
		latitude = 0
		longitude = 0
	} else if location == "Davenport 109A" {
		description = location
		latitude = 40.107335
		longitude = -88.226069
	} else if location == "Nevada Dance Studio (905 W. Nevada St.)" {
		latitude = 40.105825
		longitude = -88.219873
		description = location
	} else if location == "18th Ave Library, 175 W 18th Ave, Room 205, Oklahoma City, OK" {
		latitude = 36.102183
		longitude = -97.111245
		description = location
	} else if location == "Champaign County Fairgrounds" {
		latitude = 40.1202191
		longitude = -88.2178757
		description = location
	} else if location == "Student Union SLC Conference room" {
		latitude = 39.727282
		longitude = -89.617477
		description = location
	} else if location == "Armory, room 172 (the Innovation Studio)" {
		latitude = 40.104749
		longitude = -88.23195
		description = location
	} else if location == "Student Union Room 235" {
		latitude = 39.727282
		longitude = -89.617477
		description = location
	} else if location == "Uni 206, 210, 211" {
		latitude = 40.11314
		longitude = -88.225259
		description = location
	} else if location == "Uni 205, 206, 210" {
		latitude = 40.11314
		longitude = -88.225259
		description = location
	} else if location == "Southern Historical Association Combs Chandler 30" {
		latitude = 38.258116
		longitude = -85.756139
		description = location
	} else if location == "St. Louis, MO" {
		latitude = 38.694237
		longitude = -90.4493
		description = location
	} else if location == "Student Union SLC" {
		latitude = 39.727282
		longitude = -89.617477
		description = location
	} else if location == "Purdue University, West Lafayette, Indiana" {
		latitude = 40.425012
		longitude = -86.912645
		description = location
	} else if location == "MP 7" {
		latitude = 40.100803
		longitude = -88.23604
		description = location
	} else if location == "116 Roger Adams Lab" {
		latitude = 40.107741
		longitude = -88.224943
		description = location
	} else if location == "2700 Campus Way 45221" {
		description = location
		latitude = 39.131894
		longitude = -84.519143
	}

	return model.LocationLegacy{Description: description, Latitude: float64(latitude), Longitude: float64(longitude)}
}

// Contacts
func contactToDef(items model.ContactLegacy) model.ContactLegacy {
	return model.ContactLegacy{ContactName: items.ContactName, ContactEmail: items.ContactEmail, ContactPhone: items.ContactPhone}
}
func contactsToDef(items []model.ContactLegacy) []model.ContactLegacy {
	defs := make([]model.ContactLegacy, len(items))
	for index := range items {
		defs[index] = contactToDef(items[index])
	}
	return defs
}

// newAppEventsLogic creates new appShared
func newAppEventsLogic(app *Application, eventsBBAdapter EventsBBAdapter, logger logs.Logger) eventsLogic {
	timerDone := make(chan bool)
	return eventsLogic{app: app, eventsBBAdapter: eventsBBAdapter, timerDone: timerDone, logger: logger}
}
