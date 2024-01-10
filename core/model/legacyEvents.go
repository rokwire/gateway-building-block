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

type ResponseWS struct {
	XMLName       xml.Name `xml:"responseWS"`
	Text          string   `xml:",chardata"`
	Deprecated    string   `xml:"deprecated"`
	MaxPageSize   string   `xml:"maxPageSize"`
	PublicEventWS []struct {
		Text                               string `xml:",chardata"`
		CalendarId                         string `xml:"calendarId"`
		CalendarName                       string `xml:"calendarName"`
		EventId                            string `xml:"eventId"`
		Recurrence                         string `xml:"recurrence"`
		RecurrenceId                       string `xml:"recurrenceId"`
		OriginatingCalendarId              string `xml:"originatingCalendarId"`
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
	ID      string `json:"id"`
	EventID string `json:"eventId"`
}
