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

	"github.com/rokwire/logging-library-go/v2/logutils"
)

const (
	//TypeLegacyEvents type
	TypeLegacyEvents logutils.MessageDataType = "legacy_events"
)

// WebToolsEvent represents web tools event
type WebToolsEvent struct {
	XMLName       xml.Name `xml:"responseWS"`
	Text          string   `xml:",chardata"`
	Deprecated    string   `xml:"deprecated"`
	MaxPageSize   string   `xml:"maxPageSize"`
	PublicEventWS []struct {
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
	} `xml:"publicEventWS"`
}

// LegacyEvent wrapper
type LegacyEvent struct {
	AllDay                bool            `json:"allDay"`
	CalendarID            string          `json:"calendarId"`
	Category              string          `json:"category"`
	Subcategory           string          `json:"subcategory"`
	CreatedBy             string          `json:"createdBy"`
	LongDescription       string          `json:"longDescription"`
	DataModified          string          `json:"dataModified"`
	DataSourceEventID     string          `json:"dataSourceEventId"`
	DateCreated           string          `json:"dateCreated"`
	EndDate               string          `json:"endDate"`
	EventID               string          `json:"eventId"`
	IcalURL               string          `json:"icalUrl"`
	ID                    string          `json:"id"`
	ImageURL              *string         `json:"imageURL"`
	IsEventFree           bool            `json:"isEventFree"`
	IsVirtial             bool            `json:"isVirtual"`
	Location              *LocationLegacy `json:"location"`
	OriginatingCalendarID string          `json:"originatingCalendarId"`
	OutlookURL            string          `json:"outlookUrl"`
	RecurrenceID          *int            `json:"recurrenceId"`
	IsSuperEvent          bool            `json:"isSuperEvent"`
	RecurringFlag         bool            `json:"recurringFlag"`
	SourceID              string          `json:"sourceId"`
	Sponsor               string          `json:"sponsor"`
	StartDate             string          `json:"startDate"`
	Title                 string          `json:"title"`
	TitleURL              string          `json:"titleURL"`
	RegistrationURL       string          `json:"registrationURL"`
	SubEvents             []SubEvents     `json:"subEvents"`
}

// LocationLegacy represents event legacy location
type LocationLegacy struct {
	Description string  `json:"description"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
}

// SubEvents represents the sub events
type SubEvents struct {
	ID         string `json:"id"`
	IsFeatured bool   `json:"isFeatured"`
	Track      string `json:"track"`
}
