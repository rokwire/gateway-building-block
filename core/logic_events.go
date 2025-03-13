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

var whitelistCategoryMap = map[string]string{
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

type eventsLogic struct {
	app    *Application
	logger logs.Logger

	eventsBBAdapter EventsBBAdapter
	geoBBAdapter    GeoAdapter

	//web tools timer
	dailyWebToolsTimer *time.Timer
	timerDone          chan bool
}

func (e eventsLogic) start() error {

	//1. set up web tools timer
	go e.setupWebToolsTimer()

	//2. initialize event locations db if needs
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

func (e eventsLogic) setupWebToolsTimer() {
	e.logger.Info("Web tools timer")

	//cancel if active
	if e.dailyWebToolsTimer != nil {
		e.logger.Info("setupWebToolsTimer -> there is active timer, so cancel it")

		e.timerDone <- true
		e.dailyWebToolsTimer.Stop()
	}

	//wait until it is the correct moment from the day
	/*location, err := time.LoadLocation("America/Chicago")
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
	} */
	//log.Println(durationInSeconds)
	duration := time.Second * time.Duration(3)
	//duration := time.Second * time.Duration(durationInSeconds)
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

	webToolsCount := len(allWebToolsEvents)
	if webToolsCount == 0 {
		e.logger.Error("web tools are nil")
		return
	}

	e.logger.Infof("we loaded %d web tools events", webToolsCount)

	//process the images before the main processing
	imagesData, err := e.processImages(allWebToolsEvents)
	if err != nil {
		e.logger.Errorf("error on processing images - %s", err)
		return
	}

	//process the locations before the main processing
	locationsData, err := e.processLocations(allWebToolsEvents)
	if err != nil {
		e.logger.Errorf("error on processing locations - %s", err)
		return
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

		//3. apply rules
		statuses, err := e.applyRules(context, allWebToolsEvents)
		if err != nil {
			e.logger.Errorf("error on apply rules web tools events - %s", err)
			return err
		}

		//4. now you have to convert all allWebToolsEvents into legacy events
		newLegacyEvents := []model.LegacyEventItem{}
		for _, wt := range allWebToolsEvents {

			//prepare the id
			id := e.prepareID(wt.EventID, existingLegacyIdsMap)

			//get status
			status, exists := statuses[wt.EventID]
			if !exists {
				return errors.New("status not found for " + wt.EventID)
			}

			le := e.constructLegacyEvent(wt, id, status, now, imagesData, locationsData)
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

func (e eventsLogic) processImages(allWebtoolsEvents []model.WebToolsEvent) ([]model.ContentImagesURL, error) {
	//get the events for images processing
	forProcessingEvents, err := e.getEventsForImagesProcessing(allWebtoolsEvents)
	if err != nil {
		return nil, err
	}

	e.logger.Infof("there are %d events for images processing", len(forProcessingEvents))

	//get the events which are not processed
	notProccesed, err := e.getNotProcessedEvents(forProcessingEvents)
	if err != nil {
		return nil, err
	}

	e.logger.Infof("there are %d events to be image processed as not proccesed", len(notProccesed))

	//process the images which have not been processed
	err = e.applyProcessImages(notProccesed)
	if err != nil {
		e.logger.Error("Error on processing images")
		return nil, err
	}

	//as we already ahve processed all iages just return this dtaa to be used
	imagesData, err := e.app.storage.FindImageItems()
	if err != nil {
		return nil, err
	}

	return imagesData, nil
}

func (e eventsLogic) getEventsForImagesProcessing(allWebtoolsEvents []model.WebToolsEvent) ([]model.WebToolsEvent, error) {
	res := []model.WebToolsEvent{}
	for _, w := range allWebtoolsEvents {
		if w.LargeImageUploaded == "true" {
			res = append(res, w)
		}

	}
	return res, nil
}

func (e eventsLogic) getNotProcessedEvents(eventsForProcessing []model.WebToolsEvent) ([]model.WebToolsEvent, error) {
	allProcessed, err := e.app.storage.FindImageItems()
	if err != nil {
		return nil, err
	}

	processedMap := make(map[string]bool) // map to keep track of processed events
	for _, item := range allProcessed {
		processedMap[item.ID] = true
	}

	var notProcessedEvents []model.WebToolsEvent
	for _, event := range eventsForProcessing {
		if _, processed := processedMap[event.EventID]; !processed {
			notProcessedEvents = append(notProcessedEvents, event)
		}
	}

	return notProcessedEvents, nil
}

func (e eventsLogic) applyProcessImages(item []model.WebToolsEvent) error {
	i := 0
	for _, w := range item {

		//process image
		res, err := e.app.imageAdapter.ProcessImage(w)
		if err != nil {
			return err
		}

		if res == nil {
			continue
		}

		//mark as processed
		err = e.app.storage.InsertImageItem(*res)
		if err != nil {
			return err
		}

		e.logger.Infof("%d - %s image was processed: %s", i, res.ID, res.ImageURL)

		i++
	}
	return nil

}

// valid or ignored
func (e eventsLogic) applyRules(context storage.TransactionContext, allWebtoolsEvents []model.WebToolsEvent) (map[string]model.LegacyEventStatus, error) {
	statuses := map[string]model.LegacyEventStatus{}

	for _, wte := range allWebtoolsEvents {
		statusName := "valid"
		var reasonIgnored *string

		//all day rule
		allDay := e.applyAllDayRule(wte)
		if allDay {
			statusName = "ignored"
			reason := "skipping event as all day is true"
			reasonIgnored = &reason
		}

		//whitelisted categories
		inWhitelist, reason := e.applyWhitelistCategoriesRule(wte)
		if !inWhitelist {
			statusName = "ignored"
			reasonIgnored = reason
		}

		status := model.LegacyEventStatus{Name: statusName, ReasonIgnored: reasonIgnored}
		statuses[wte.EventID] = status
	}

	return statuses, nil
	/*
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
	*/
}

func (e eventsLogic) applyAllDayRule(wt model.WebToolsEvent) bool {
	timeType := wt.TimeType
	return timeType == "NONE"
}

// returns reason
func (e eventsLogic) applyWhitelistCategoriesRule(wt model.WebToolsEvent) (bool, *string) {
	category := wt.EventType
	lowerCategory := strings.ToLower(category)

	_, exists := whitelistCategoryMap[lowerCategory]
	if !exists {
		reason := fmt.Sprintf("skipping event as category is %s", category)
		return false, &reason
	}
	return true, nil
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

func (e eventsLogic) constructLegacyEvent(g model.WebToolsEvent, id string, status model.LegacyEventStatus,
	now time.Time, imagesData []model.ContentImagesURL, locationsData []model.LegacyLocation) model.LegacyEventItem {

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
	var startDateObj, endDateObj *time.Time

	chicagoLocation, err := time.LoadLocation("America/Chicago")
	if err != nil {
		e.logger.Errorf("cannot get timezone - America/Chicago")
	}

	if timeType == "START_TIME_ONLY" {
		startDate = g.StartDate
		startTime = g.StartTime
		startDateTimeStr := fmt.Sprintf("%s %s", startDate, startTime)
		startDateObjTmp, _ := time.ParseInLocation("1/2/2006 3:04 pm", startDateTimeStr, chicagoLocation)
		startDateObj = &startDateObjTmp

		/*endDate = g.EndDate
		endDateTimeStr := fmt.Sprintf("%s 11:59 pm", endDate)
		endDateObjTmp, _ := time.ParseInLocation("1/2/2006 3:04 pm", endDateTimeStr, chicagoLocation)
		endDateObj = &endDateObjTmp*/
	} else if timeType == "START_AND_END_TIME" {
		startDate = g.StartDate
		startTime = g.StartTime
		startDateTimeStr := fmt.Sprintf("%s %s", startDate, startTime)
		startDateObjTmp, _ := time.ParseInLocation("1/2/2006 3:04 pm", startDateTimeStr, chicagoLocation)
		startDateObj = &startDateObjTmp

		endDate = g.EndDate
		endTime = g.EndTime
		endDateTimeStr := fmt.Sprintf("%s %s", endDate, endTime)
		endDateObjTmp, _ := time.ParseInLocation("1/2/2006 3:04 pm", endDateTimeStr, chicagoLocation)
		endDateObj = &endDateObjTmp
	} else if timeType == "NONE" {
		allDay = true

		startDate = g.StartDate
		startDateTimeStr := fmt.Sprintf("%s 12:00 am", startDate)
		startDateObjTmp, _ := time.ParseInLocation("1/2/2006 3:04 pm", startDateTimeStr, chicagoLocation)
		startDateObj = &startDateObjTmp

		/*	endDateTimeStr := fmt.Sprintf("%s 11:59 pm", endDate)
			endDateObjTmp, _ := time.ParseInLocation("1/2/2006 3:04 pm", endDateTimeStr, chicagoLocation)
			endDateObj = &endDateObjTmp */
	}

	startDateStr := ""
	endDateStr := ""

	if startDateObj != nil {
		startDateStr = startDateObj.UTC().Format("Mon, 02 Jan 2006 15:04:05 GMT")
	}
	if endDateObj != nil {
		endDateStr = endDateObj.UTC().Format("Mon, 02 Jan 2006 15:04:05 GMT")
	}

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

	//image url
	imageURL := e.getImageURL(g.EventID, imagesData)
	loc := constructLocation(g, locationsData)

	return model.LegacyEventItem{SyncProcessSource: syncProcessSource, SyncDate: now, Status: status,
		Item: model.LegacyEvent{ID: id, Category: g.EventType, CreatedBy: createdBy,
			OriginatingCalendarID: g.OriginatingCalendarID, OriginatingCalendarName: g.OriginatingCalendarName,
			IsVirtial: isVirtual, DataModified: modifiedDate, DateCreated: createdDate,
			Sponsor: g.Sponsor, Title: g.Title, CalendarID: g.CalendarID, SourceID: "0", AllDay: allDay, IsEventFree: costFree, Cost: g.Cost, LongDescription: g.Description,
			TitleURL: g.TitleURL, RegistrationURL: g.RegistrationURL, RecurringFlag: Recurrence, IcalURL: icalURL, OutlookURL: outlookURL,
			RecurrenceID: recurrenceID, Location: loc, Contacts: contatsLegacy,
			DataSourceEventID: g.EventID, StartDate: startDateStr, EndDate: endDateStr,
			Tags: tags, TargetAudience: targetAudience, ImageURL: imageURL}}
}

func (e eventsLogic) getImageURL(eventID string, imageData []model.ContentImagesURL) *string {
	for _, image := range imageData {
		if image.ID == eventID {
			return &image.ImageURL
		}
	}
	return nil
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

func (e eventsLogic) processLocations(allWebtoolsEvents []model.WebToolsEvent) ([]model.LegacyLocation, error) {
	//get the locations for processing
	forProcessingLocations, err := e.getLocationsForProcessing(allWebtoolsEvents)
	if err != nil {
		return nil, err
	}

	e.logger.Infof("there are %d locations for processing", len(forProcessingLocations))

	//get the locations which are not processed
	notProccesed, err := e.getNotProcessedLocations(forProcessingLocations)
	if err != nil {
		return nil, err
	}

	e.logger.Infof("there are %d locations to be processed as not proccesed", len(notProccesed))

	//process the locations which have not been processed
	err = e.applyProcessLocations(notProccesed)
	if err != nil {
		e.logger.Error("Error on processing locations")
		return nil, err
	}

	//as we already have processed all locations just return this data to be used
	locationsData, err := e.app.storage.FindLegacyLocationItems()
	if err != nil {
		return nil, err
	}

	return locationsData, nil
}

func (e eventsLogic) getLocationsForProcessing(allWebtoolsEvents []model.WebToolsEvent) ([]string, error) {
	locationsMap := make(map[string]bool)
	for _, event := range allWebtoolsEvents {
		if len(event.Location) == 0 {
			continue
		}

		locationsMap[event.Location] = true
	}

	res := []string{}
	for location := range locationsMap {
		res = append(res, location)
	}
	return res, nil
}

func (e eventsLogic) getNotProcessedLocations(locationsForProcessing []string) ([]string, error) {
	allProcessed, err := e.app.storage.FindLegacyLocations()
	if err != nil {
		return nil, err
	}

	processedMap := make(map[string]bool) // map to keep track of processed events
	for _, item := range allProcessed {
		processedMap[item.Name] = true
	}

	var notProcessedEvents []string
	for _, loc := range locationsForProcessing {
		if _, processed := processedMap[loc]; !processed {
			notProcessedEvents = append(notProcessedEvents, loc)
		}
	}

	return notProcessedEvents, nil
}

func (e eventsLogic) applyProcessLocations(locations []string) error {
	i := 0
	for _, loc := range locations {

		//process the location
		founded, err := e.geoBBAdapter.FindLocation(loc)
		if err != nil {
			return err
		}

		if founded == nil {
			e.logger.Infof("%d - %s NOT found", i, loc)
			continue
		}

		//mark as processed
		err = e.app.storage.InsertLegacyLocationItem(*founded)
		if err != nil {
			return err
		}

		e.logger.Infof("%d - %s WAS found", i, loc)

		i++
	}
	return nil

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

func constructLocation(event model.WebToolsEvent, locations []model.LegacyLocation) *model.LocationLegacy {
	eventLocation := event.Location

	//check for empty location
	if len(eventLocation) == 0 {
		emptyLocation := constructEmptyLocation(eventLocation)
		return &emptyLocation
	}

	//in some cases we do not use the founded locations directly by the location key as in some cases the geo service(Google)
	//gives bad results. In this case we have defined a new correct search key.

	//try if need to change the key
	searchKey := usePredefinedLocationKey(event.CalendarName, event.Sponsor, event.Location)
	if searchKey == nil {
		searchKey = &event.Location
	}

	//we have a search key, so try to find a location
	founded := findLocation(*searchKey, locations)
	if founded != nil {
		//return it

		return &model.LocationLegacy{Description: founded.Description,
			Latitude: *founded.Lat, Longitude: *founded.Long}
	}

	//not found so return empty location
	emptyLocation := constructEmptyLocation(eventLocation)
	return &emptyLocation
}

func constructEmptyLocation(location string) model.LocationLegacy {
	description := location
	latitude := 0.0
	longitude := 0.0
	return model.LocationLegacy{Description: description, Latitude: float64(latitude), Longitude: float64(longitude)}
}

// usePredefinedLocationKey looks for a static location based on the calendar name, sponsor, and location description
func usePredefinedLocationKey(calendarName, sponsor, location string) *string {

	type tip struct {
		CalendarName    string
		SponsorKeyword  string
		LocationKeyword string
		AccessName      string
	}

	var tip4CalALoc = []tip{
		{CalendarName: "Krannert Center", SponsorKeyword: "", LocationKeyword: "studio", AccessName: "Krannert Center"},
		{CalendarName: "Krannert Center", SponsorKeyword: "", LocationKeyword: "stage", AccessName: "Krannert Center"},
		{CalendarName: "General Events", SponsorKeyword: "ncsa", LocationKeyword: "ncsa", AccessName: "NCSA"},
		{CalendarName: "National Center for Supercomputing Applications master calendar", SponsorKeyword: "", LocationKeyword: "ncsa", AccessName: "NCSA"},
	}

	for _, tip := range tip4CalALoc {
		if tip.CalendarName == calendarName &&
			strings.Contains(strings.ToLower(sponsor), tip.SponsorKeyword) &&
			strings.Contains(strings.ToLower(location), tip.LocationKeyword) {
			return &tip.AccessName
		}
	}

	return nil
}

func findLocation(loc string, locations []model.LegacyLocation) *model.LegacyLocation {
	for _, location := range locations {
		if location.Name == loc {
			return &location
		}
	}
	return nil
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
func newAppEventsLogic(app *Application, eventsBBAdapter EventsBBAdapter, geoBBAdapter GeoAdapter, logger logs.Logger) eventsLogic {
	timerDone := make(chan bool)
	return eventsLogic{app: app, eventsBBAdapter: eventsBBAdapter, geoBBAdapter: geoBBAdapter, timerDone: timerDone, logger: logger}
}
