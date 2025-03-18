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

package model

import (
	"encoding/xml"
	"time"

	"github.com/rokwire/logging-library-go/v2/logutils"
)

const (
	//TypeLegacyEvents type
	TypeLegacyEvents logutils.MessageDataType = "legacy_events"
)

// WebToolsResponse represents web tools response item
type WebToolsResponse struct {
	XMLName        xml.Name        `xml:"responseWS"`
	Text           string          `xml:",chardata"`
	Deprecated     string          `xml:"deprecated"`
	MaxPageSize    string          `xml:"maxPageSize"`
	WebToolsEvents []WebToolsEvent `xml:"publicEventWS"`
}

// WebToolsEvent represents web tools event entity
type WebToolsEvent struct {
	Text                               string `xml:",chardata"`
	CalendarID                         string `xml:"calendarId"`
	CalendarName                       string `xml:"calendarName"`
	EventID                            string `xml:"eventId"`
	Recurrence                         string `xml:"recurrence"`
	RecurrenceID                       string `xml:"recurrenceId"`
	OriginatingCalendarID              string `xml:"originatingCalendarId"`
	OriginatingCalendarName            string `xml:"originatingCalendarName"`
	Title                              string `xml:"title"`
	TitleURL                           string `xml:"titleURL"`
	EventType                          string `xml:"eventType"`
	Sponsor                            string `xml:"sponsor"`
	DateDisplay                        string `xml:"dateDisplay"`
	StartDate                          string `xml:"startDate"`
	EndDate                            string `xml:"endDate"`
	TimeType                           string `xml:"timeType"`
	StartTime                          string `xml:"startTime"`
	EndTime                            string `xml:"endTime"`
	EndTimeLabel                       string `xml:"endTimeLabel"`
	InPersonEvent                      string `xml:"inPersonEvent"`
	Location                           string `xml:"location"`
	Description                        string `xml:"description"`
	Speaker                            string `xml:"speaker"`
	RegistrationLabel                  string `xml:"registrationLabel"`
	RegistrationURL                    string `xml:"registrationURL"`
	ContactName                        string `xml:"contactName"`
	ContactEmail                       string `xml:"contactEmail"`
	ContactPhone                       string `xml:"contactPhone"`
	CostFree                           string `xml:"costFree"`
	Cost                               string `xml:"cost"`
	CreatedBy                          string `xml:"createdBy"`
	CreatedDate                        string `xml:"createdDate"`
	EditedBy                           string `xml:"editedBy"`
	EditedDate                         string `xml:"editedDate"`
	Summary                            string `xml:"summary"`
	AudienceFacultyStaff               string `xml:"audienceFacultyStaff"`
	AudienceStudents                   string `xml:"audienceStudents"`
	AudiencePublic                     string `xml:"audiencePublic"`
	AudienceAlumni                     string `xml:"audienceAlumni"`
	AudienceParents                    string `xml:"audienceParents"`
	ShareWithUrbanaEventsInChicagoArea string `xml:"shareWithUrbanaEventsInChicagoArea"`
	ShareWithResearch                  string `xml:"shareWithResearch"`
	ShareWithSpeakers                  string `xml:"shareWithSpeakers"`
	ShareWithIllinoisMobileApp         string `xml:"shareWithIllinoisMobileApp"`
	ThumbImageUploaded                 string `xml:"thumbImageUploaded"`
	LargeImageUploaded                 string `xml:"largeImageUploaded"`
	LargeImageSize                     string `xml:"largeImageSize"`
	VirtualEvent                       string `xml:"virtualEvent"`
	VirtualEventURL                    string `xml:"virtualEventURL"`
	Topic                              []struct {
		Text string `xml:",chardata"`
		ID   string `xml:"id"`
		Name string `xml:"name"`
	} `xml:"topic"`
}

// Blacklist represents web tools blacklist ids
type Blacklist struct {
	Name string   `json:"name" bson:"name"`
	Data []string `json:"data" bson:"data"`
}

// WebToolsItems represents web tools originating calendar ids
type WebToolsItems struct {
	Count int    `json:"count"`
	ID    string `json:"id" bson:"originatingCalendarId"`
	Name  string `json:"Name" bson:"originatingCalendarName"`
}

// WebToolsItems represents web tools originating calendar ids
type WebToolsSource struct {
	Count         int             `json:"count"`
	WebToolsItems []WebToolsItems `json:"webtools-source"`
}

// TPsItems represents web tools originating calendar ids
type TPsItems struct {
	Count int    `json:"count"`
	ID    string `json:"id" bson:"originatingCalendarId"`
	Name  string `json:"Name" bson:"originatingCalendarName"`
}

// TPsSource represents web tools originating calendar ids
type TPsSource struct {
	Count         int        `json:"count"`
	WebToolsItems []TPsItems `json:"tps_api"`
}

// WebToolsSummary represents web tools summary
type WebToolsSummary struct {
	AllEventsCount            int         `json:"all_events_count"`
	ValidEventsCount          int         `json:"valid_events_count"`
	IgnoredEventsCount        int         `json:"ignored_events_count"`
	TotalOriginatingCalendars int         `json:"total_originating_calendars"`
	Valid                     Valid       `json:"valid"`
	Ignored                   Ignored     `json:"ignored"`
	Blacklists                []Blacklist `json:"blacklists"`
}

type Valid struct {
	WebtoolsSource WebToolsSource `json:"webtools_source"`
	TpsAPI         TPsSource      `json:"tps_api"`
}

type Ignored struct {
	WebtoolsSource WebToolsSource `json:"webtools_source"`
	TpsAPI         TPsSource      `json:"tps_api"`
}

// LegacyEvent wrapper
type LegacyEvent struct {
	AllDay                  bool            `json:"allDay" bson:"allDay"`
	CalendarID              string          `json:"calendarId" bson:"calendarId"`
	Category                string          `json:"category" bson:"category"`
	Subcategory             string          `json:"subcategory" bson:"subcategory"`
	CreatedBy               string          `json:"createdBy" bson:"createdBy"`
	LongDescription         string          `json:"longDescription" bson:"longDescription"`
	DataModified            string          `json:"dataModified" bson:"dataModified"`
	DataSourceEventID       string          `json:"dataSourceEventId" bson:"dataSourceEventId"`
	DateCreated             string          `json:"dateCreated" bson:"dateCreated"`
	EndDate                 string          `json:"endDate" bson:"endDate"`
	EventID                 string          `json:"eventId" bson:"eventId"`
	IcalURL                 string          `json:"icalUrl" bson:"icalUrl"`
	ID                      string          `json:"id" bson:"id"`
	ImageURL                *string         `json:"imageURL" bson:"imageURL"`
	IsEventFree             bool            `json:"isEventFree" bson:"isEventFree"`
	IsVirtial               bool            `json:"isVirtual" bson:"isVirtual"`
	Location                *LocationLegacy `json:"location" bson:"location"`
	OriginatingCalendarID   string          `json:"originatingCalendarId" bson:"originatingCalendarId"`
	OriginatingCalendarName string          `json:"originatingCalendarName" bson:"originatingCalendarName"`
	OutlookURL              string          `json:"outlookUrl" bson:"outlookUrl"`
	RecurrenceID            *int            `json:"recurrenceId" bson:"recurrenceId"`
	IsSuperEvent            bool            `json:"isSuperEvent" bson:"isSuperEvent"`
	RecurringFlag           bool            `json:"recurringFlag" bson:"recurringFlag"`
	SourceID                string          `json:"sourceId" bson:"sourceId"`
	Sponsor                 string          `json:"sponsor" bson:"sponsor"`
	StartDate               string          `json:"startDate" bson:"startDate"`
	Title                   string          `json:"title" bson:"title"`
	TitleURL                string          `json:"titleURL" bson:"titleURL"`
	Tags                    *[]string       `json:"tags" bson:"tags"`
	TargetAudience          *[]string       `json:"targetAudience" bson:"targetAudience"`
	RegistrationURL         string          `json:"registrationURL" bson:"registrationURL"`
	Contacts                []ContactLegacy `json:"contacts" bson:"contacts"`
	SubEvents               []SubEvents     `json:"subEvents" bson:"subEvents"`
	Cost                    string          `json:"cost" bson:"cost"`
}

// LocationLegacy represents event legacy location
type LocationLegacy struct {
	Description string  `json:"description" bson:"description"`
	Latitude    float64 `json:"latitude" bson:"latitude"`
	Longitude   float64 `json:"longitude" bson:"longitude"`
	Address     string  `json:"address" bson:"address"`
	Building    string  `json:"building" bson:"building"`
	Floor       int     `json:"floor" bson:"floor"`
	Room        string  `json:"room" bson:"room"`
}

// SubEvents represents the sub events
type SubEvents struct {
	ID         string `json:"id" bson:"id"`
	IsFeatured bool   `json:"isFeatured" bson:"isFeatured"`
	Track      string `json:"track" bson:"track"`
}

// LegacyEventItem represents legacy event entity which contains legacy event + other sync info
type LegacyEventItem struct {
	SyncProcessSource string            `bson:"sync_process_source"` //webtools-direct or events-bb-initial or events-tps-api
	SyncDate          time.Time         `bson:"sync_date"`
	Status            LegacyEventStatus `bson:"status"`

	Item LegacyEvent `bson:"item"`

	CreateInfo *CreateInfo `bson:"create_info"`
}

// LegacyEventStatus represents legacy event status
type LegacyEventStatus struct {
	Name          string  `bson:"name"` //valid or ignored
	ReasonIgnored *string `bson:"reason_ignored"`
}

// ContactLegacy represents event legacy contacts
type ContactLegacy struct {
	ContactName  string `json:"contactName" bson:"contactName"`
	ContactEmail string `json:"contactEmail" bson:"contactEmail"`
	ContactPhone string `json:"contactPhone" bson:"contactPhone"`
}

// CreateInfo represents entity creation info
type CreateInfo struct {
	Time      time.Time `json:"time" bson:"time"`
	AccountID string    `json:"account_id" bson:"account_id"`
}
