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

package courses

import (
	model "apigateway/core/model"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

type campusData struct {
	Object  string                  `json:"object"`
	Version string                  `json:"version"`
	List    []studentTermCourseInfo `json:"list"`
}
type studentName struct {
	LastName  string `json:"lastName"`
	FirstName string `json:"firstName"`
}
type studentDemo struct {
	Name            studentName `json:"name"`
	InstitutionalID string      `json:"institutionalId"`
}
type codeDescription struct {
	Description string `json:"description"`
	Code        string `json:"code"`
}

type validPartOfTerm struct {
	Description string `json:"description"`
	StartDate   string `json:"startDate"`
	EndDate     string `json:"endDate"`
	Code        string `json:"code"`
}

type campuscourse struct {
	CourseAbbreviation string          `json:"courseAbbreviation"`
	CourseNumber       string          `json:"courseNumber"`
	CourseTitle        string          `json:"courseTitle"`
	ValidCampus        codeDescription `json:"validCampus"`
	ValidCollege       codeDescription `json:"validCollege"`
	ValidDepartment    codeDescription `json:"validDepartment"`
}
type instructorName struct {
	Type      string `json:"type"`
	LastName  string `json:"lastName"`
	FirstName string `json:"firstName"`
}
type instructorDemo struct {
	Name             instructorName `json:"name"`
	PrimaryIndicator string         `json:"primaryIndicator"`
	InstitutionalID  string         `json:"institutionalId"`
	EmailAddress     string         `json:"emailAddress"`
}
type courseSectionInstructor struct {
	LightweightPerson instructorDemo `json:"lightweightPerson"`
}

type courseSectionSession struct {
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
	CourseSectionInstructor []courseSectionInstructor `json:"courseSectionInstructor"`
	MeetingDateOrRange      string                    `json:"meetingDateOrRange"`
	ValidMeetingType        codeDescription           `json:"validMeetingType"`
	CourseSectionSessionID  string                    `json:"courseSectionSessionID"`
	ValidCourseScheduleType codeDescription           `json:"validCourseScheduleType"`
	ValidBuilding           codeDescription           `json:"validBuilding"`
}
type courseSection struct {
	CourseReferenceNumber string                 `json:"courseReferenceNumber"`
	SectionNumber         string                 `json:"sectionNumber"`
	ValidTerm             codeDescription        `json:"validTerm"`
	Course                campuscourse           `json:"course"`
	StartDate             string                 `json:"startDate"`
	EndDate               string                 `json:"endDate"`
	CreditHours           string                 `json:"creditHours"`
	ValidPartOfTerm       validPartOfTerm        `json:"validPartOfTerm"`
	CourseSectionSession  []courseSectionSession `json:"courseSectionSession"`
}
type courseRegistration struct {
	ValidRegistrationStatusType  codeDescription `json:"validRegistrationStatusType"`
	ValidCourseRegistrationLevel codeDescription `json:"validCourseRegistrationLevel"`
	StudentCourseStartDate       string          `json:"studentCourseStartDate"`
	StudentCourseEndDate         string          `json:"studentCourseEndDate"`
	ValidRegistrationStatus      codeDescription `json:"validRegistrationStatus"`
	ValidGradingMode             codeDescription `json:"validGradingMode"`
	CourseSection                courseSection   `json:"courseSection"`
}
type studentTermCourseInfo struct {
	Student               studentDemo          `json:"lightweightPerson"`
	ValidEnrollmentStatus codeDescription      `json:"validEnrollmentStatus"`
	ValidTerm             codeDescription      `json:"validTerm"`
	CourseRegistration    []courseRegistration `json:"courseRegistration"`
}

//StudentCourseAdapter is a vendor specific structure that implements the GiesCourse lookup interface
type StudentCourseAdapter struct {
	CourseAPIEndpoint string
	CourseAPIKey      string
}

//NewCourseAdapter returns a vendor specific implementation of the Course lookup interface
func NewCourseAdapter(url string, apikey string) *StudentCourseAdapter {
	return &StudentCourseAdapter{CourseAPIEndpoint: url, CourseAPIKey: apikey}

}

//newCourse maps the campus course data to the course data sent back to the app.
func newCourse(cr courseRegistration, courseSectionSessionIndex int) *model.Course {
	ret := model.Course{}
	ret.Number = cr.CourseSection.CourseReferenceNumber
	ret.ShortName = cr.CourseSection.Course.CourseAbbreviation + " " + cr.CourseSection.Course.CourseNumber
	ret.Title = cr.CourseSection.Course.CourseTitle
	ret.InstructionMethod = cr.CourseSection.CourseSectionSession[courseSectionSessionIndex].ValidCourseScheduleType.Code
	crn := cr.CourseSection.CourseReferenceNumber
	css := cr.CourseSection.CourseSectionSession[courseSectionSessionIndex]
	newCS := newCourseSection(css, crn)
	ret.Section = *newCS

	return &ret
}

//newCourseSection maps the coursesectionsession data from campus to the coursesection data sent back to the app
func newCourseSection(cs courseSectionSession, crn string) *model.CourseSection {
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

//GetStudentCourses returns a list of courses for the given tudent
func (lv *StudentCourseAdapter) GetStudentCourses(uin string, termid string, accessToken string) (*[]model.Course, int, error) {

	finalURL := lv.CourseAPIEndpoint + "/student-registration/student-enrollment-query/v2_0/" + uin + "/" + termid

	retValue := make([]model.Course, 0)

	campusData, statusCode, err := lv.getData(finalURL, accessToken)
	if err != nil {
		return nil, statusCode, err
	}

	if len(campusData.List) == 0 {
		return nil, 404, errors.New("No course data found")
	}

	if len(campusData.List[0].CourseRegistration) == 0 {
		return nil, 404, errors.New("No course data found")
	}

	for i := 0; i < len(campusData.List[0].CourseRegistration); i++ {
		course := campusData.List[0].CourseRegistration[i]
		if course.ValidRegistrationStatusType.Code == "R" {
			for i := 0; i < len(course.CourseSection.CourseSectionSession); i++ {
				retValue = append(retValue, *newCourse(course, i))
			}
		}
	}

	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	return &retValue, statusCode, nil
}

func (lv *StudentCourseAdapter) getData(targetURL string, accessToken string) (*campusData, int, error) {
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, targetURL, nil)

	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	req.Header.Add("Authorization", "Bearer "+accessToken)
	req.Header.Set("Ocp-Apim-Subscription-Key", lv.CourseAPIKey)

	res, err := client.Do(req)
	if err != nil {
		return nil, res.StatusCode, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, res.StatusCode, err
	}

	if res.StatusCode == 401 {
		return nil, res.StatusCode, errors.New(res.Status)
	}

	if res.StatusCode == 403 {
		return nil, res.StatusCode, errors.New(res.Status)
	}

	if res.StatusCode == 400 {
		return nil, res.StatusCode, errors.New("Bad request to api end point")
	}

	if res.StatusCode == 406 {
		return nil, res.StatusCode, errors.New("Server returned 406: possible uin claim mismatch")
	}
	//campus api returns a 502 when there is no course data
	if res.StatusCode == 502 {
		return nil, 404, errors.New(res.Status)
	}

	if res.StatusCode == 200 || res.StatusCode == 203 {
		data := campusData{}

		err = json.Unmarshal(body, &data)

		if err != nil {
			return nil, res.StatusCode, err
		}
		return &data, res.StatusCode, nil
	}

	return nil, res.StatusCode, errors.New("Error making request: " + fmt.Sprint(res.StatusCode) + ": " + string(body))

}
