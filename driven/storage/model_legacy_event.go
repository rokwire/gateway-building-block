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

package storage

// legacyEvent wrapper
type legacyEvent struct {
	AllDay            bool    `bson:"allDay"`
	CalendarID        string  `bson:"calendarId"`
	Category          string  `bson:"category"`
	Subcategory       string  `bson:"subcategory"`
	CreatedBy         string  `bson:"createdBy"`
	LongDescription   string  `bson:"longDescription"`
	DataModified      string  `bson:"dataModified"`
	DataSourceEventID string  `bson:"dataSourceEventId"`
	DateCreated       string  `bson:"dateCreated"`
	EndDate           string  `bson:"endDate"`
	EventID           string  `bson:"eventId"`
	IcalURL           string  `bson:"icalUrl"`
	ImageURL          *string `bson:"imageURL"`
	IsEventFree       bool    `bson:"isEventFree"`
	IsVirtial         bool    `bson:"isVirtual"`
	Location          *struct {
		Description string  `bson:"description"`
		Latitude    float64 `bson:"latitude"`
		Longitude   float64 `bson:"longitude"`
	} `bson:"location"`
	OriginatingCalendarID string  `bson:"originatingCalendarId"`
	OutlookURL            string  `bson:"outlookUrl"`
	RecurrenceID          *int    `bson:"recurrenceId"`
	IsSuperEvent          bool    `bson:"isSuperEvent"`
	RecurringFlag         bool    `bson:"recurringFlag"`
	SourceID              string  `bson:"sourceId"`
	Sponsor               string  `bson:"sponsor"`
	StartDate             string  `bson:"startDate"`
	Title                 string  `bson:"title"`
	TitleURL              *string `bson:"titleURL"`
	RegistrationURL       *string `bson:"registrationURL"`
	SubEvents             []struct {
		ID         string `bson:"id"`
		IsFeatured bool   `bson:"isFeatured"`
		Track      string `bson:"track"`
	} `bson:"subEvents"`
}
