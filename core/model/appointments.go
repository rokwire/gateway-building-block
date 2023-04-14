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
	"github.com/rokwire/logging-library-go/v2/logutils"
)

const (
	//TypeAppointments type
	TypeAppointments logutils.MessageDataType = "appointments"
)

// Question represents a question asked as part of an appointmentr request
type Question struct {
	ID           string   `json:"id" bson:"id"`
	ProviderID   int      `json:"provider_Id" bson:"provider_Id"`
	Required     bool     `json:"required" bson:"required"`
	Type         string   `json:"type" bson:"type"`
	SelectValues []string `json:"select_values" bson:"select_values"`
}

// TimeSlot represents an avaialable appontment timeslot
type TimeSlot struct {
	ID         int                    `json:"id" bson:"id"`
	ProviderID int                    `json:"provider_Id" bson:"provider_Id"`
	UnitID     int                    `json:"unit_id" bson:"unit_id"`
	PersonID   int                    `json:"person_id" bson:"person_id"`
	StartTime  string                 `json:"start_time" bson:"start_time"`
	EndTime    string                 `json:"end_time" bson:"end_time"`
	Capacity   int                    `json:"capacity" bson:"capacity"`
	Filled     bool                   `json:"filled" bson:"filled"`
	Details    map[string]interface{} `json:"details" bson:"details"`
}

// AppointmentOptions represents the available timeslots and questions for a unitid/advisorid calendar
type AppointmentOptions struct {
	TimeSlots []TimeSlot `json:"time_slots" bson:"time_slots"`
	Questions []Question `json:"questions" bson:"questions"`
}

// AppointmentUnit represents units with availalbe appointment integrations
type AppointmentUnit struct {
	ID               int    `json:"id" bson:"id"`
	ProviderID       int    `json:"provider_id" bson:"provider_id"`
	Name             string `json:"name" bson:"name"`
	Location         string `json:"location" bson:"location"`
	HoursOfOperation string `json:"hours_of_operation" bson:"hours_of_operation"`
	Details          string `json:"details" bason:"details"`
}

// AppointmentPerson represents a person who is accepting appointments
type AppointmentPerson struct {
	ID            string `json:"id" bson:"id"`
	ProviderID    int    `json:"provider_id" bson:"provider_id"`
	UnitID        int    `json:"unit_id" bson:"unit_id"`
	NextAvailable string `json:"next_available" bson:"next_available"`
	Name          string `json:"name" bson:"name"`
	Notes         string `json:"notes" bson:"notes"`
}
