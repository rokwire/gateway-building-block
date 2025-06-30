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

import (
	"github.com/rokwire/rokwire-building-block-sdk-go/utils/logging/logutils"
)

const (
	//TypeGiesCourse type
	TypeGiesCourse logutils.MessageDataType = "giescourse"
)

const (
	//TypeCourseData type
	TypeCourseData logutils.MessageDataType = "coursedata"
)

// GiesCourse represents the elements of a course returned for Gies students
type GiesCourse struct {
	Term       string `json:"term" bson:"term"`
	Subject    string `json:"subject" bson:"subject"`
	Number     string `json:"number" bson:"number"`
	Section    string `json:"section" bson:"section"`
	Title      string `json:"title" bson:"title"`
	Instructor string `json:"instructor" bson:"instructor"`
}

// CourseSection represents the elements of a course section
type CourseSection struct {
	Days                  string   `json:"days" bson:"days"`
	MeetingDateOrRange    string   `json:"meeting_dates_or_range" bson:"meeting_dates_or_range"`
	Room                  string   `json:"room" bson:"room"`
	BuildingName          string   `json:"buildingname" bson:"buildingname"`
	BuildingID            string   `json:"buildingid" bson:"buildingid"`
	InstructionType       string   `json:"instructiontype" bson:"instructiontype"`
	Instructor            string   `json:"instructor" bson:"instructor"`
	StartTime             string   `json:"start_time" bson:"start_time"`
	EndTime               string   `json:"endtime" bson:"endtime"`
	Location              Building `json:"building" bson:"building"`
	CourseReferenceNumber string   `json:"courseReferenceNumber" bson:"courseReferenceNumber"`
	SectionNumber         string   `json:"course_section" bson:"course_section"`
}

// Course represents the full elements of a course
type Course struct {
	Title             string        `json:"coursetitle" bson:"coursetitle"`
	ShortName         string        `json:"courseshortname" bson:"courseshortname"`
	Number            string        `json:"coursenumber" bson:"coursenumber"`
	InstructionMethod string        `json:"instructionmethod" bson:"instructionmethod"`
	Section           CourseSection `json:"coursesection" bson:"coursesection"`
}
