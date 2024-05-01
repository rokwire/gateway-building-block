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
	"strings"
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
		_, err = e.app.storage.InsertLegacyEvents(context, resultList)
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
	e.logger.Info("Web tools timer")

	//cancel if active
	if e.dailyWebToolsTimer != nil {
		e.logger.Info("setupWebToolsTimer -> there is active timer, so cancel it")

		e.timerDone <- true
		e.dailyWebToolsTimer.Stop()
	}

	//wait until it is the correct moment from the day
	location, err := time.LoadLocation("America/Chicago")
	if err != nil {
		e.logger.Errorf("Error getting location:%s\n", err.Error())
	}
	now := time.Now().In(location)
	e.logger.Infof("setupWebToolsTimer -> now - hours:%d minutes:%d seconds:%d\n", now.Hour(), now.Minute(), now.Second())

	nowSecondsInDay := 60*60*now.Hour() + 60*now.Minute() + now.Second()
	desiredMoment := 18000

	var durationInSeconds int
	log.Printf("setupWebToolsTimer -> nowSecondsInDay:%d desiredMoment:%d\n", nowSecondsInDay, desiredMoment)
	if nowSecondsInDay <= desiredMoment {
		e.logger.Infof("setupWebToolsTimer -> not web tools process today, so the first process will be today")
		durationInSeconds = desiredMoment - nowSecondsInDay
	} else {
		e.logger.Infof("setupWebToolsTimer -> the web tools process has already been processed today, so the first process will be tomorrow")
		leftToday := 86400 - nowSecondsInDay
		durationInSeconds = leftToday + desiredMoment // the time which left today + desired moment from tomorrow
	}
	log.Println(durationInSeconds)
	//duration := time.Second * time.Duration(3)
	duration := time.Second * time.Duration(durationInSeconds)
	e.logger.Infof("setupWebToolsTimer -> first call after %s", duration)

	e.dailyWebToolsTimer = time.NewTimer(duration)
	select {
	case <-e.dailyWebToolsTimer.C:
		e.logger.Info("setupWebToolsTimer -> web tools timer expired")
		e.dailyWebToolsTimer = nil

		e.process()
	case <-e.timerDone:
		// timer aborted
		e.logger.Info("setupWebToolsTimer -> web tools timer aborted")
		e.dailyWebToolsTimer = nil
	}
}

func (e eventsLogic) process() {
	e.logger.Info("Webtools process")

	//process work
	e.processWebToolsEvents()

	//generate new processing after 24 hours
	duration := time.Hour * 24
	e.logger.Infof("Webtools process -> next call after %s", duration)
	e.dailyWebToolsTimer = time.NewTimer(duration)
	select {
	case <-e.dailyWebToolsTimer.C:
		e.logger.Info("Webtools process -> timer expired")
		e.dailyWebToolsTimer = nil

		e.process()
	case <-e.timerDone:
		// timer aborted
		e.logger.Info("Webtools process -> timer aborted")
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
	var withimage []model.WebToolsEvent
	for _, l := range allWebToolsEvents {
		if l.LargeImageUploaded != "" {
			withimage = append(withimage, l)
		}
		allWebToolsEvents = withimage
	}

	webToolsCount := len(allWebToolsEvents)
	if webToolsCount == 0 {
		e.logger.Error("web tools are nil")
		return
	}

	e.logger.Infof("we loaded %d web tools events", webToolsCount)

	contentImagesFromTheDataBase, err := e.app.storage.FindImageItems()
	if err != nil {
		e.logger.Error("Error on finding image items")
		return
	}

	images, err := e.app.imageAdapter.ProcessImages(allWebToolsEvents)
	if err != nil {
		e.logger.Error("Error on finding image items")
		return
	}
	for _, t := range contentImagesFromTheDataBase {
		for _, l := range images {
			if t.ID != l.ID && t.ImageURL != l.ImageURL {
				err = e.app.storage.InsertImageItems(l)

			}
		}
	}

	// Create a map to store ImageURLs with corresponding IDs
	imageURLMap := make(map[string]string)
	for _, ciu := range images {
		imageURLMap[ciu.ID] = ciu.ImageURL
	}

	for i := range allWebToolsEvents {
		if allWebToolsEvents[i].LargeImageUploaded == "false" {
			allWebToolsEvents[i].ImageURL = ""
		} else if imageURL, ok := imageURLMap[allWebToolsEvents[i].EventID]; ok {
			allWebToolsEvents[i].ImageURL = imageURL
		}
	}

	now := time.Now()

	//in transaction
	err = e.app.storage.PerformTransaction(func(context storage.TransactionContext) error {
		//1. first we must keep the events ids for the webtools events(sourceId = "0") because we will remove all of them and later recreated with the new ones
		webtoolsItemsFromStorage, err := e.app.storage.FindLegacyEventItemsBySourceID(context, "0")
		if err != nil {
			e.logger.Errorf("error on loading webtools events from the storage - %s", err)
			return err
		}

		existingLegacyIdsMap := make(map[string]string)
		for _, w := range webtoolsItemsFromStorage {
			if len(w.Item.DataSourceEventID) > 0 {
				existingLegacyIdsMap[w.Item.DataSourceEventID] = w.Item.ID
			}
		}

		//2. once we already have the ids then we have to remove all webtools events from the database
		err = e.app.storage.DeleteLegacyEventsBySourceID(context, "0")
		if err != nil {
			e.logger.Errorf("error on deleting legacy events from the storage - %s", err)
			return err
		}

		//at this moment the all webtools items are removed from the database and we can add what comes from webtools

		//3. we have a requirement to ignore events or modify them before applying
		modifiedWebToolsEvents, err := e.modifyWebtoolsEventsList(allWebToolsEvents)
		if err != nil {
			e.logger.Errorf("error on ignoring web tools events - %s", err)
			return err
		}

		//4. now you have to convert all allWebToolsEvents into legacy events
		newLegacyEvents := []model.LegacyEventItem{}
		for _, wt := range modifiedWebToolsEvents {

			//prepare the id
			id := e.prepareID(wt.EventID, existingLegacyIdsMap)

			le := e.constructLegacyEvent(wt, id, now)
			newLegacyEvents = append(newLegacyEvents, le)
		}

		//5. store all them in the database
		_, err = e.app.storage.InsertLegacyEvents(context, newLegacyEvents)
		if err != nil {
			e.logger.Errorf("error on saving events to the storage - %s", err)
			return err
		}
		// It is all!

		return nil
	}, 180000)

	if err != nil {
		e.logger.Errorf("error performing transaction - %s", err)
		return
	}
}

// ignore or modify webtools events
func (e eventsLogic) modifyWebtoolsEventsList(allWebtoolsEvents []model.WebToolsEvent) ([]model.WebToolsEvent, error) {
	modifiedList := []model.WebToolsEvent{}

	ignored := 0
	modified := 0

	//whitelist with categories which we care + map for category conversions
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

	for _, wte := range allWebtoolsEvents {
		currentWte := wte
		category := currentWte.EventType
		lowerCategory := strings.ToLower(category)

		//ignore all day events
		allDay := e.isAllDay(currentWte)
		if allDay {
			e.logger.Info("skipping event as all day is true")
			ignored++
			continue
		}

		//get only the events which have a category from the whitelist
		if newCategory, ok := categoryMap[lowerCategory]; ok {
			currentWte.EventType = newCategory
			e.logger.Infof("modifying event category from %s to %s", category, newCategory)

			modified++
		} else {
			e.logger.Infof("skipping event as category is %s", category)
			ignored++
			continue
		}

		//add it to the modified list
		modifiedList = append(modifiedList, currentWte)
	}

	e.logger.Infof("ignored events count is %d", ignored)
	e.logger.Infof("modified events count is %d", modified)
	e.logger.Infof("final modified list is %d", len(modifiedList))

	return modifiedList, nil
}

func (e eventsLogic) isAllDay(wt model.WebToolsEvent) bool {
	timeType := wt.TimeType
	if timeType == "NONE" {
		return true
	}
	return false
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

	createdBy := g.CreatedBy

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
	location := constructLocation(g.Location)
	con := model.ContactLegacy{ContactName: g.CalendarName, ContactEmail: g.ContactEmail, ContactPhone: g.ContactName}
	var contacts []model.ContactLegacy
	contacts = append(contacts, con)
	contatsLegacy := contactsToDef(contacts)

	modifiedDate := e.formatDate(g.EditedDate)
	createdDate := e.formatDate(g.CreatedDate)

	//start date + end date (+all day)
	allDay := false

	timeType := g.TimeType

	var startDate, startTime, endDate, endTime string
	var startDateObj, endDateObj time.Time

	chicagoLocation, err := time.LoadLocation("America/Chicago")
	if err != nil {
		e.logger.Errorf("cannot get timezone - America/Chicago")
	}

	if timeType == "START_TIME_ONLY" {
		startDate = g.StartDate
		startTime = g.StartTime
		startDateTimeStr := fmt.Sprintf("%s %s", startDate, startTime)
		startDateObj, _ = time.ParseInLocation("1/2/2006 3:04 pm", startDateTimeStr, chicagoLocation)

		endDate = g.EndDate
		endDateTimeStr := fmt.Sprintf("%s 11:59 pm", endDate)
		endDateObj, _ = time.ParseInLocation("1/2/2006 3:04 pm", endDateTimeStr, chicagoLocation)
	} else if timeType == "START_AND_END_TIME" {
		startDate = g.StartDate
		startTime = g.StartTime
		startDateTimeStr := fmt.Sprintf("%s %s", startDate, startTime)
		startDateObj, _ = time.ParseInLocation("1/2/2006 3:04 pm", startDateTimeStr, chicagoLocation)

		endDate = g.EndDate
		endTime = g.EndTime
		endDateTimeStr := fmt.Sprintf("%s %s", endDate, endTime)
		endDateObj, _ = time.ParseInLocation("1/2/2006 3:04 pm", endDateTimeStr, chicagoLocation)
	} else if timeType == "NONE" {
		allDay = true

		startDate = g.StartDate
		endDate = g.EndDate
		startDateTimeStr := fmt.Sprintf("%s 12:00 am", startDate)
		startDateObj, _ = time.ParseInLocation("1/2/2006 3:04 pm", startDateTimeStr, chicagoLocation)

		endDateTimeStr := fmt.Sprintf("%s 11:59 pm", endDate)
		endDateObj, _ = time.ParseInLocation("1/2/2006 3:04 pm", endDateTimeStr, chicagoLocation)
	}

	startDateStr := startDateObj.UTC().Format("Mon, 02 Jan 2006 15:04:05 GMT")
	endDateStr := endDateObj.UTC().Format("Mon, 02 Jan 2006 15:04:05 GMT")

	//end - start date + end date (+all day)

	//tags
	var tags *[]string
	if len(g.Topic) > 0 {
		tagsList := []string{}
		for _, t := range g.Topic {
			tagsList = append(tagsList, t.Name)
		}
		tags = &tagsList
	}
	//end tags

	//target audience
	var targetAudience *[]string

	var targetAudienceList []string
	if g.AudienceFacultyStaff == "true" {
		targetAudienceList = append(targetAudienceList, "faculty", "staff")
	}
	if g.AudienceStudents == "true" {
		targetAudienceList = append(targetAudienceList, "students")
	}
	if g.AudiencePublic == "true" {
		targetAudienceList = append(targetAudienceList, "public")
	}
	if g.AudienceAlumni == "true" {
		targetAudienceList = append(targetAudienceList, "alumni")
	}
	if g.AudienceParents == "true" {
		targetAudienceList = append(targetAudienceList, "parents")
	}

	if len(targetAudienceList) != 0 {
		targetAudience = &targetAudienceList
	}
	//end target audience

	return model.LegacyEventItem{SyncProcessSource: syncProcessSource, SyncDate: now,
		Item: model.LegacyEvent{ID: id, Category: g.EventType, CreatedBy: createdBy, OriginatingCalendarID: g.OriginatingCalendarID, IsVirtial: isVirtual,
			DataModified: modifiedDate, DateCreated: createdDate,
			Sponsor: g.Sponsor, Title: g.Title, CalendarID: g.CalendarID, SourceID: "0", AllDay: allDay, IsEventFree: costFree, Cost: g.Cost, LongDescription: g.Description,
			TitleURL: g.TitleURL, RegistrationURL: g.RegistrationURL, RecurringFlag: Recurrence, IcalURL: icalURL, OutlookURL: outlookURL,
			RecurrenceID: recurrenceID, Location: &location, Contacts: contatsLegacy,
			DataSourceEventID: g.EventID, StartDate: startDateStr, EndDate: endDateStr,
			Tags: tags, TargetAudience: targetAudience, ImageURL: &g.ImageURL}}
}

func (e eventsLogic) formatDate(wtDate string) string {
	dateFormat := "1/2/2006"
	timeFormat := "3:04 pm"

	wtDateTime := wtDate + " 12:00 am"
	dataObj, err := time.Parse(dateFormat+" "+timeFormat, wtDateTime)
	if err != nil {
		return ""
	}

	dataObj = dataObj.Add(5 * time.Hour)

	result := dataObj.Format("2006-01-02T15:04:05")
	return result
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

func constructLocation(location string) model.LocationLegacy {
	description := location
	latitude := 0.0
	longitude := 0.0

	if location == "Davenport 109A" {
		latitude = 40.107335
		longitude = -88.226069
	} else if location == "Nevada Dance Studio (905 W. Nevada St.)" {
		latitude = 40.105825
		longitude = -88.219873
	} else if location == "18th Ave Library, 175 W 18th Ave, Room 205, Oklahoma City, OK" {
		latitude = 36.102183
		longitude = -97.111245
	} else if location == "Champaign County Fairgrounds" {
		latitude = 40.1202191
		longitude = -88.2178757
	} else if location == "Student Union SLC Conference room" {
		latitude = 39.727282
		longitude = -89.617477
	} else if location == "Armory, room 172 (the Innovation Studio)" {
		latitude = 40.104749
		longitude = -88.23195
	} else if location == "Student Union Room 235" {
		latitude = 39.727282
		longitude = -89.617477
	} else if location == "Uni 206, 210, 211" {
		latitude = 40.11314
		longitude = -88.225259
	} else if location == "Uni 205, 206, 210" {
		latitude = 40.11314
		longitude = -88.225259
	} else if location == "Southern Historical Association Combs Chandler 30" {
		latitude = 38.258116
		longitude = -85.756139
	} else if location == "St. Louis, MO" {
		latitude = 38.694237
		longitude = -90.4493
	} else if location == "Student Union SLC" {
		latitude = 39.727282
		longitude = -89.617477
	} else if location == "Purdue University, West Lafayette, Indiana" {
		latitude = 40.425012
		longitude = -86.912645
	} else if location == "MP 7" {
		latitude = 40.100803
		longitude = -88.23604
	} else if location == "116 Roger Adams Lab" {
		latitude = 40.107741
		longitude = -88.224943
	} else if location == "2700 Campus Way 45221" {
		latitude = 39.131894
		longitude = -84.519143
	} else if location == "The Orange Room, Main Library - 1408 W. Gregory Drive, Champaign IL" {
		latitude = 40.1047044
		longitude = -88.22901039999999
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
