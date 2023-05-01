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

import (
	model "application/core/model"
	"strings"
)

// CampusData represents the full data returned by the campus courses end point
type CampusData struct {
	Object  string                  `json:"object"`
	Version string                  `json:"version"`
	List    []StudentTermCourseInfo `json:"list"`
}

// StudentName represents the Name property of a student
type StudentName struct {
	LastName  string `json:"lastName"`
	FirstName string `json:"firstName"`
}

// StudentDemo represents the demographic information returned from teh campus courses end point
type StudentDemo struct {
	Name            StudentName `json:"name"`
	InstitutionalID string      `json:"institutionalId"`
}

// CodeDescription represents a code description used in the campus course definition
type CodeDescription struct {
	Description string `json:"description"`
	Code        string `json:"code"`
}

// ValidPartOfTerm represents the campus valid term definition
type ValidPartOfTerm struct {
	Description string `json:"description"`
	StartDate   string `json:"startDate"`
	EndDate     string `json:"endDate"`
	Code        string `json:"code"`
}

// Campuscourse represents the course data returned by the campus course endpoint
type Campuscourse struct {
	CourseAbbreviation string          `json:"courseAbbreviation"`
	CourseNumber       string          `json:"courseNumber"`
	CourseTitle        string          `json:"courseTitle"`
	ValidCampus        CodeDescription `json:"validCampus"`
	ValidCollege       CodeDescription `json:"validCollege"`
	ValidDepartment    CodeDescription `json:"validDepartment"`
}

// InstructorName represents the data used by campus to show an instructors name
type InstructorName struct {
	Type      string `json:"type"`
	LastName  string `json:"lastName"`
	FirstName string `json:"firstName"`
}

// InstructorDemo represents the capus definition of an instructor's demographic information
type InstructorDemo struct {
	Name             InstructorName `json:"name"`
	PrimaryIndicator string         `json:"primaryIndicator"`
	InstitutionalID  string         `json:"institutionalId"`
	EmailAddress     string         `json:"emailAddress"`
}

// CourseSectionInstructor represents the campus definition of a course's instructor
type CourseSectionInstructor struct {
	LightweightPerson InstructorDemo `json:"lightweightPerson"`
}

// CourseSectionSession represents a course section meeting details in the campus data
type CourseSectionSession struct {
	MeetsOnMondayFlag       string                    `json:"meetsOnMondayFlag"`
	MeetsOnTuesdayFlag      string                    `json:"meetsOnTuesdayFlag"`
	MeetsOnWednesdayFlag    string                    `json:"meetsOnWednesdayFlag"`
	MeetsOnThursdayFlag     string                    `json:"meetsOnThursdayFlag"`
	MeetsOnFridayFlag       string                    `json:"meetsOnFridayFlag"`
	MeetsOnSaturdayFlag     string                    `json:"meetsOnSaturdayFlag"`
	MeetsOnSundayFlag       string                    `json:"meetsOnSundayFlag"`
	Room                    string                    `json:"room"`
	ArchibusBuildingNumber  string                    `json:"archibusBuildingNumber"`
	StartTime               string                    `json:"startTime"`
	EndTime                 string                    `json:"endTime"`
	CreditHours             string                    `json:"creditHours"`
	CourseSectionInstructor []CourseSectionInstructor `json:"courseSectionInstructor"`
	MeetingDateOrRange      string                    `json:"meetingDateOrRange"`
	ValidMeetingType        CodeDescription           `json:"validMeetingType"`
	CourseSectionSessionID  string                    `json:"courseSectionSessionID"`
	ValidCourseScheduleType CodeDescription           `json:"validCourseScheduleType"`
	ValidBuilding           CodeDescription           `json:"validBuilding"`
}

// CourseSection represents a course section in the campus data
type CourseSection struct {
	CourseReferenceNumber string                 `json:"courseReferenceNumber"`
	SectionNumber         string                 `json:"sectionNumber"`
	ValidTerm             CodeDescription        `json:"validTerm"`
	Course                Campuscourse           `json:"course"`
	StartDate             string                 `json:"startDate"`
	EndDate               string                 `json:"endDate"`
	CreditHours           string                 `json:"creditHours"`
	ValidPartOfTerm       ValidPartOfTerm        `json:"validPartOfTerm"`
	CourseSectionSession  []CourseSectionSession `json:"courseSectionSession"`
}

// CourseRegistration represents a students course registration in the campus data
type CourseRegistration struct {
	ValidRegistrationStatusType  CodeDescription `json:"validRegistrationStatusType"`
	ValidCourseRegistrationLevel CodeDescription `json:"validCourseRegistrationLevel"`
	StudentCourseStartDate       string          `json:"studentCourseStartDate"`
	StudentCourseEndDate         string          `json:"studentCourseEndDate"`
	ValidRegistrationStatus      CodeDescription `json:"validRegistrationStatus"`
	ValidGradingMode             CodeDescription `json:"validGradingMode"`
	CourseSection                CourseSection   `json:"courseSection"`
}

// StudentTermCourseInfo represents the registration data for a student in a given term
type StudentTermCourseInfo struct {
	Student               StudentDemo          `json:"lightweightPerson"`
	ValidEnrollmentStatus CodeDescription      `json:"validEnrollmentStatus"`
	ValidTerm             CodeDescription      `json:"validTerm"`
	CourseRegistration    []CourseRegistration `json:"courseRegistration"`
}

// NewCourse maps the campus course data to the course data sent back to the app.
func NewCourse(cr CourseRegistration, courseSectionSessionIndex int) *model.Course {
	ret := model.Course{}
	ret.Number = cr.CourseSection.CourseReferenceNumber
	ret.ShortName = cr.CourseSection.Course.CourseAbbreviation + " " + cr.CourseSection.Course.CourseNumber
	ret.Title = cr.CourseSection.Course.CourseTitle
	ret.InstructionMethod = cr.CourseSection.CourseSectionSession[courseSectionSessionIndex].ValidCourseScheduleType.Code
	crn := cr.CourseSection.CourseReferenceNumber
	css := cr.CourseSection.CourseSectionSession[courseSectionSessionIndex]
	newCS := NewCourseSection(css, crn)
	ret.Section = *newCS

	return &ret
}

// NewCourseSection maps the coursesectionsession data from campus to the coursesection data sent back to the app
func NewCourseSection(cs CourseSectionSession, crn string) *model.CourseSection {
	ret := model.CourseSection{}
	ret.BuildingName = cs.ValidBuilding.Description
	ret.Room = cs.Room
	ret.BuildingID = cs.ArchibusBuildingNumber
	ret.MeetingDateOrRange = cs.MeetingDateOrRange
	ret.StartTime = cs.StartTime
	ret.EndTime = cs.EndTime
	ret.InstructionType = cs.ValidCourseScheduleType.Code

	//we only want the primary instructor
	if len(cs.CourseSectionInstructor) == 0 {
		ret.Instructor = ""
	} else {

		for i := 0; i < len(cs.CourseSectionInstructor); i++ {
			instructor := cs.CourseSectionInstructor[i]
			if instructor.LightweightPerson.PrimaryIndicator == "Y" {
				ret.Instructor = instructor.LightweightPerson.Name.LastName + ", " + instructor.LightweightPerson.Name.FirstName
				break
			}
		}
	}
	ret.CourseReferenceNumber = crn
	//data coming from campus only contains the fields for each day it is actually taught
	//days the course is not taught will be empty since they won't be int he data
	days := make([]string, 0)
	if cs.MeetsOnMondayFlag != "" {
		days = append(days, "M")
	}
	if cs.MeetsOnTuesdayFlag != "" {
		days = append(days, "Tu")
	}
	if cs.MeetsOnWednesdayFlag != "" {
		days = append(days, "W")
	}
	if cs.MeetsOnThursdayFlag != "" {
		days = append(days, "Th")
	}
	if cs.MeetsOnFridayFlag != "" {
		days = append(days, "F")
	}
	if cs.MeetsOnSaturdayFlag != "" {
		days = append(days, "S")
	}
	if cs.MeetsOnSundayFlag != "" {
		days = append(days, "Su")
	}
	ret.Days = strings.Join(days, ",")

	return &ret
}
