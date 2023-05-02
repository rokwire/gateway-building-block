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

package uiuc

// EngineeringCalendar represents an entry in the engineering calendar list
type EngineeringCalendar struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// EngineeringAdvisor represents an advisor with calendar entries on a given calendar
type EngineeringAdvisor struct {
	ID                string `json:"advisorId"`
	Name              string `json:"advisorName"`
	Message           string `json:"message"`
	CalendarID        int    `json:"calendarId"`
	CalendarName      string `json:"calendarName"`
	Active            bool   `json:"isActive"`
	AppointmentLength int    `json:"appointmentLength"`
	AvailableSlots    int    `json:"availableSlots"`
	NextAvailableDate string `json:"nextAvailableDate"`
	Announcement      string `json:"announcement"`
	AnnouncementDate  string `json:"announcementDate"`
}

// EngineeringCalendarAdvisors represents a calendar including all advisors
type EngineeringCalendarAdvisors struct {
	Adivsors []EngineeringAdvisor `json:"advisors"`
	ID       int                  `json:"id"`
	Name     string               `json:"name"`
}

// EngineeringTimeSlot represents a time slot on an advisors calendar
type EngineeringTimeSlot struct {
	ID        int    `json:"id"`
	StartDate string `json:"startDate"`
	EndDate   string `json:"endDate"`
}

// EngineeringQuestion represents a question on an advisor's appointment app
type EngineeringQuestion struct {
	ID              string   `json:"id"`
	Order           int      `json:"order"`
	Title           string   `json:"title"`
	Type            string   `json:"type"`
	UploadType      string   `json:"uploadType"`
	SelectionValues []string `json:"selectionValues"`
}

// EngineeringAdvisorWithSchedule represents an advisor plus their schedule information
type EngineeringAdvisorWithSchedule struct {
	TimeSlots         []EngineeringTimeSlot `json:"slotTimes"`
	Questions         []EngineeringQuestion `json:"questions"`
	ID                string                `json:"advisorId"`
	Name              string                `json:"advisorName"`
	Message           string                `json:"message"`
	CalendarID        int                   `json:"calendarId"`
	CalendarName      string                `json:"calendarName"`
	Active            bool                  `json:"isActive"`
	AppointmentLength int                   `json:"appointmentLength"`
	AvailableSlots    int                   `json:"availableSlots"`
	NextAvailableDate string                `json:"nextAvailableDate"`
	Announcement      string                `json:"announcement"`
	AnnouncementDate  string                `json:"announcementDate"`
}

// EngineeringAdvisorAppointments represents an advisors availability
type EngineeringAdvisorAppointments struct {
	TimeSlots []EngineeringTimeSlot `json:"slots"`
	Questions []EngineeringQuestion `json:"questions"`
}

// EngineeringAnswer represnets an answer to an engineering question
type EngineeringAnswer struct {
	QuestionID string `json:"questionId"`
	Value      string `json:"value"`
	UploadID   int    `json:"uploadId"`
}

// EngineeringAppointmentPost represents data needed to create an appointmen
type EngineeringAppointmentPost struct {
	UIN     int                 `json:"uin"`
	SlotID  int                 `json:"slotId"`
	Answers []EngineeringAnswer `json:"answers"`
}
