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

package model

// GiesCourse represents the elements of a course returned for Gies students
type GiesCourse struct {
	Term       string `json:"term"`
	Subject    string `json:"subject"`
	Number     string `json:"number"`
	Section    string `json:"section"`
	Title      string `json:"title"`
	Instructor string `json:"instructor"`
}

// CourseSection represents the elements of a course section
type CourseSection struct {
	Days                  string   `json:"days"`
	MeetingDateOrRange    string   `json:"meeting_dates_or_range"`
	Room                  string   `json:"room"`
	BuildingName          string   `json:"buildingname"`
	BuildingID            string   `json:"buildingid"`
	InstructionType       string   `json:"instructiontype"`
	Instructor            string   `json:"instructor"`
	StartTime             string   `json:"start_time"`
	EndTime               string   `json:"endtime"`
	Location              Building `json:"building"`
	CourseReferenceNumber string   `json:"courseReferenceNumber"`
}

// Course represents the full elements of a course
type Course struct {
	Title             string        `json:"coursetitle"`
	ShortName         string        `json:"courseshortname"`
	Number            string        `json:"coursenumber"`
	InstructionMethod string        `json:"instructionmethod"`
	Section           CourseSection `json:"coursesection"`
}
