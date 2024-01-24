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
		//	return err
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

func (e eventsLogic) setupWebToolsTimer() {
	log.Println("Web tools timer")

	//cancel if active
	if e.dailyWebToolsTimer != nil {
		log.Println("setupWebToolsTimer -> there is active timer, so cancel it")

		e.timerDone <- true
		e.dailyWebToolsTimer.Stop()
	}

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
	}
	//log.Println(durationInSeconds)
	//duration := time.Second * time.Duration(20)
	duration := time.Second * time.Duration(durationInSeconds)
	log.Printf("setupWebToolsTimer -> first call after %s", duration)

	e.dailyWebToolsTimer = time.NewTimer(duration)
	select {
	case <-e.dailyWebToolsTimer.C:
		log.Println("setupWebToolsTimer -> web tools timer expired")
		e.dailyWebToolsTimer = nil

		e.getAllEvents()
	case <-e.timerDone:
		// timer aborted
		log.Println("setupWebToolsTimer -> web tools timer aborted")
		e.dailyWebToolsTimer = nil
	}
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
	/*  for pe in XML2JSON:
	    # decide not to skip if location not exist or empty.
	    # if not pe.get('location'):
	    #     continue
	    entry = dict()

	    if pe['timeType'] == "ALL_DAY":
	        # skip all day event. (https://github.com/rokwire/events-manager/issues/1086)
	        continue

	    if pe.get("shareWithIllinoisMobileApp", "false") == "false":
	        dataSourceEventId = pe.get("eventId", "")
	        result = find_one(
	            current_app.config['EVENT_COLLECTION'],
	            condition={'dataSourceEventId': dataSourceEventId}
	        )
	        if result:
	            notSharedWithMobileList.append(result["_id"])
	        continue


	    if 'virtualEventURL' in pe:
	        entry['virtualEventUrl'] = pe['virtualEventURL']


	    # Required Field

	    # find geographical location
	    skip_google_geoservice = False
	    if not pe.get('location'):
	        skip_google_geoservice = True
	    # flag for checking online event
	    found_online_event = False
	    # compare with the existing location
	    existing_event = find_one(current_app.config['EVENT_COLLECTION'], condition={'dataSourceEventId': entry[
	        'dataSourceEventId'
	    ]})
	    existing_location = existing_event.get('location')

	    # filter out online location
	    if pe.get('location'):
	        pe_location_lower_case = pe["location"].lower()
	        for excluded_location in Config.EXCLUDED_LOCATION:
	            if excluded_location.lower() in pe_location_lower_case:
	                skip_google_geoservice = True
	                found_online_event = True
	                entry["location"] = {
	                    "description": pe["location"]
	                }
	                break

	    if existing_location:
	        # mark previouly unidentified online events
	        if found_online_event:
	            if existing_location.get('latitude') and existing_location.get('longitude'):
	                entry["replace_event"] = True
	        else:
	            existing_description = existing_location.get('description')
	            if existing_description == pe.get('location'):
	                if existing_location.get('latitude') and existing_location.get('longitude'):
	                    skip_google_geoservice = True
	                    lat = existing_location.get('latitude')
	                    lng = existing_location.get('longitude')
	                    GeoInfo = {
	                        'latitude': lat,
	                        'longitude': lng,
	                        'description': pe['location']
	                    }
	                    entry['location'] = GeoInfo

	    if not entry.get('isVirtual') or not skip_google_geoservice:
	        location = pe.get('location')
	        calendarName = pe['calendarName']
	        sponsor = pe['sponsor']

	        if location in predefined_locations:
	            entry['location'] = predefined_locations[location]
	            __logger.info("assign predefined geolocation: calendarId: " + str(entry['calendarId']) + ", dataSourceEventId: " + str(entry['dataSourceEventId']))
	        elif location:
	            (found, GeoInfo) = search_static_location(calendarName, sponsor, location)
	            if found:
	                entry['location'] = GeoInfo
	            else:
	                try:
	                    GeoResponse = gmaps.geocode(address=location+',Urbana', components={'administrative_area': 'Urbana', 'country': "US"})
	                except googlemaps.exceptions.ApiError as e:
	                    __logger.error("API Key Error: {}".format(e))
	                    entry['location'] = {'description': pe['location']}
	                    xmltoMongoDB.append(entry)
	                    continue

	                if len(GeoResponse) != 0:
	                    lat = GeoResponse[0]['geometry']['location']['lat']
	                    lng = GeoResponse[0]['geometry']['location']['lng']
	                    GeoInfo = {
	                        'latitude': lat,
	                        'longitude': lng,
	                        'description': pe['location']
	                    }
	                    entry['location'] = GeoInfo
	                else:
	                    entry['location'] = {'description': pe['location']}
	                    __logger.error("calendarId: %s, dataSourceEventId: %s,  location: %s geolocation not found" %
	                          (entry.get('calendarId'), entry.get('dataSourceEventId'), entry.get('location')))
	        else:
	            entry['location'] = {
	                        'description': ""
	                    }
	    else:
	        entry['location'] = {
	            'description': ""
	        }
	    entry_location = entry['location']
	    if pe['timeType'] == "START_TIME_ONLY":
	        startDate = pe['startDate']
	        startTime = pe['startTime']
	        startDateObj = datetime.strptime(startDate + ' ' + startTime + '', '%m/%d/%Y %I:%M %p')
	        endDate = pe['endDate']
	        endDateObj = datetime.strptime(endDate + ' 11:59 pm', '%m/%d/%Y %I:%M %p')
	        # normalize event datetime to UTC
	        # TODO: current default time zone is CDT
	        entry['startDate'] = event_time_conversion.utctime(startDateObj, entry_location.get('latitude', 40.1153287), entry_location.get('longitude', -88.2280659))
	        entry['endDate'] = event_time_conversion.utctime(endDateObj, entry_location.get('latitude', 40.1153287), entry_location.get('longitude', -88.2280659))

	    elif pe['timeType'] == "START_AND_END_TIME":
	        startDate = pe['startDate']
	        startTime = pe['startTime']
	        endDate = pe['endDate']
	        endTime = pe['endTime']
	        startDateObj = datetime.strptime(startDate + ' ' + startTime, '%m/%d/%Y %I:%M %p')
	        endDateObj = datetime.strptime(endDate + ' ' + endTime, '%m/%d/%Y %I:%M %p')
	        # normalize event datetime to UTC
	        # TODO: current default time zone is CDT
	        entry['startDate'] = event_time_conversion.utctime(startDateObj, entry_location.get('latitude', 40.1153287), entry_location.get('longitude', -88.2280659))
	        entry['endDate'] = event_time_conversion.utctime(endDateObj, entry_location.get('latitude', 40.1153287), entry_location.get('longitude', -88.2280659))

	    # when time type is None, usually happens in calendar 468
	    elif pe['timeType'] == "NONE":
	        entry['allDay'] = True
	        startDate = pe['startDate']
	        endDate = pe['endDate']
	        startDateObj = datetime.strptime(startDate + ' 12:00 am', '%m/%d/%Y %I:%M %p')
	        endDateObj = datetime.strptime(endDate + ' 11:59 pm', '%m/%d/%Y %I:%M %p')
	        # normalize event datetime to UTC
	        # TODO: current default time zone is CDT
	        entry['startDate'] = event_time_conversion.utctime(startDateObj, entry_location.get('latitude', 40.1153287), entry_location.get('longitude', -88.2280659))
	        entry['endDate'] = event_time_conversion.utctime(endDateObj, entry_location.get('latitude', 40.1153287), entry_location.get('longitude', -88.2280659))

	    # Optional Field

	    if 'recurrenceId' in pe:
	        entry['recurrenceId'] = int(pe['recurrenceId'])

	    targetAudience = []
	    targetAudience.extend(["faculty", "staff"]) if pe['audienceFacultyStaff'] == "true" else None
	    targetAudience.append("students") if pe['audienceStudents'] == "true" else None
	    targetAudience.append("public") if pe['audiencePublic'] == "true" else None
	    targetAudience.append("alumni") if pe['audienceAlumni'] == "true" else None
	    targetAudience.append("parents") if pe['audienceParents'] == "true" else None
	    if len(targetAudience) != 0:
	        entry['targetAudience'] = targetAudience

	    contacts = []
	    contact = {}
	    if 'contactName' in pe:
	        name_list = pe['contactName'].split(' ')
	        contact['firstName'] = pe['contactName'].split(' ')[0].rstrip(',')
	        if len(name_list) > 1:
	            contact['lastName'] = pe['contactName'].split(' ')[1]
	        else:
	            contact['lastName'] = ""
	    if 'contactEmail' in pe:
	        contact['email'] = pe['contactEmail']
	    if 'contactPhone' in pe:
	        contact['phone'] = pe['contactPhone']
	    contacts.append(contact) if len(contact) != 0 else None
	    if len(contacts) != 0:
	        entry['contacts']= contacts

	    # creation information
	    dateCreatedObj = datetime.strptime(pe['createdDate'] + ' 12:00 am', '%m/%d/%Y %I:%M %p')
	    entry['dateCreated'] = (dateCreatedObj+timedelta(hours=5)).strftime('%Y-%m-%dT%H:%M:%S')
	    if 'createdBy' in pe:
	        entry['createdBy'] = pe['createdBy']

	    # edit information
	    dataModifiedObj = datetime.strptime(pe['editedDate'] + ' 12:00 am', '%m/%d/%Y %I:%M %p')
	    entry['dataModified'] = (dataModifiedObj+timedelta(hours=5)).strftime('%Y-%m-%dT%H:%M:%S')

	    xmltoMongoDB.append(entry)
	__logger.info("Get {} parsed events".format(len(xmltoMongoDB)))
	__logger.info("Get {} not shareWithIllinoisMobileApp events".format(len(notSharedWithMobileList)))
	return (xmltoMongoDB, notSharedWithMobileList)

	*/

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

	return model.LegacyEvent{Category: g.EventType, OriginatingCalendarID: g.OriginatingCalendarID, IsVirtial: isVirtual, DataModified: g.EventID,
		Sponsor: g.Sponsor, Title: g.Title, CalendarID: g.CalendarID, SourceID: "0", AllDay: false, IsEventFree: costFree, LongDescription: g.Description,
		TitleURL: g.TitleURL, RegistrationURL: g.RegistrationURL, RecurringFlag: Recurrence, IcalURL: icalURL, OutlookURL: outlookURL,
		RecurrenceID: recurrenceID}
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

// newAppEventsLogic creates new appShared
func newAppEventsLogic(app *Application, eventsBBAdapter EventsBBAdapter, logger logs.Logger) eventsLogic {
	timerDone := make(chan bool)
	return eventsLogic{app: app, eventsBBAdapter: eventsBBAdapter, timerDone: timerDone, logger: logger}
}
