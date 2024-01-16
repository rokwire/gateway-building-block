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

package eventsbb

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/rokwire/logging-library-go/v2/logs"
)

// Adapter implements the Storage interface
type Adapter struct {
	baseURL string
	apiKey  string
}

// NewEventsBBAdapter creates new instance
func NewEventsBBAdapter(legacyEventsBaseURL, legacyEventsAPIKey string) *Adapter {
	return &Adapter{
		baseURL: legacyEventsBaseURL,
		apiKey:  legacyEventsAPIKey, // pragma: allowlist secret
	}
}

// LegacyEvent wrapper
type LegacyEvent struct {
	AllDay            bool    `json:"allDay"`
	CalendarID        string  `json:"calendarId"`
	Category          string  `json:"category"`
	Subcategory       string  `json:"subcategory"`
	CreatedBy         string  `json:"createdBy"`
	LongDescription   string  `json:"longDescription"`
	DataModified      string  `json:"dataModified"`
	DataSourceEventID string  `json:"dataSourceEventId"`
	DateCreated       string  `json:"dateCreated"`
	EndDate           *string `json:"endDate"`
	EventID           string  `json:"eventId"`
	IcalURL           string  `json:"icalUrl"`
	ID                string  `json:"id"`
	ImageURL          *string `json:"imageURL"`
	IsEventFree       bool    `json:"isEventFree"`
	IsVirtial         bool    `json:"isVirtual"`
	Location          *struct {
		Description string  `json:"description"`
		Latitude    float64 `json:"latitude"`
		Longitude   float64 `json:"longitude"`
	} `json:"location"`
	OriginatingCalendarID string  `json:"originatingCalendarId"`
	OutlookURL            string  `json:"outlookUrl"`
	RecurrenceID          *int    `json:"recurrenceId"`
	IsSuperEvent          bool    `json:"isSuperEvent"`
	RecurringFlag         bool    `json:"recurringFlag"`
	SourceID              string  `json:"sourceId"`
	Sponsor               string  `json:"sponsor"`
	StartDate             string  `json:"startDate"`
	Title                 string  `json:"title"`
	TitleURL              *string `json:"titleURL"`
	RegistrationURL       *string `json:"registrationURL"`
	SubEvents             []struct {
		ID         string `json:"id"`
		IsFeatured bool   `json:"isFeatured"`
		Track      string `json:"track"`
	} `json:"subEvents"`
}

// LoadAllLegacyEvents sends notification to a user
func (na *Adapter) LoadAllLegacyEvents(log *logs.Log) ([]LegacyEvent, error) {

	url := fmt.Sprintf("%s/events", na.baseURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Errorf("legacy_events.LoadAllLegacyEvents: error creating load legacy events request - %s", err)
		return nil, err
	}
	req.Header.Set("ROKWIRE-API-KEY", na.apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Errorf("legacy_events.LoadAllLegacyEvents: error creating load legacy events request - %s", err)
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		errorResponse, _ := io.ReadAll(resp.Body)
		if errorResponse != nil {
			log.Errorf("legacy_events.LoadAllLegacyEvents: error with response code - %s", errorResponse)
		}
		log.Errorf("legacy_events.LoadAllLegacyEvents: error with response code - %d", resp.StatusCode)
		return nil, fmt.Errorf("SendNotification:error with response code != 200")
	}
	var list []LegacyEvent
	err = json.NewDecoder(resp.Body).Decode(&list)
	if err != nil {
		log.Errorf("legacy_events.LoadAllLegacyEvents: error with response code - %d", resp.StatusCode)
		return nil, fmt.Errorf("SendNotification: %s", err)
	}
	return list, nil

}
