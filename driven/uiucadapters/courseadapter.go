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

package uiucadapters

import (
	model "application/core/model"
	uiuc "application/core/model/uiuc"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

// StudentCourseAdapter is a vendor specific structure that implements the UIUC and GiesCourse lookup interface
type StudentCourseAdapter struct {
}

// NewCourseAdapter returns a vendor specific implementation of the Course lookup interface
func NewCourseAdapter() StudentCourseAdapter {
	return StudentCourseAdapter{}

}

// GetStudentCourses returns a list of courses for the given tudent
func (lv StudentCourseAdapter) GetStudentCourses(uin string, termid string, accessToken string, conf *model.EnvConfigData) (*[]model.Course, int, error) {

	courseAPIEP := conf.CentralCampusURL
	courseAuth := conf.CentralCampusKey
	if termid == "" {
		crntDate := time.Now()
		crntYear := strconv.Itoa(crntDate.Year())
		if crntDate.Month() >= 6 && crntDate.Month() <= 12 {
			termid = "1" + crntYear + "8"
		} else {
			termid = "1" + crntYear + "1"
		}
	}
	finalURL := courseAPIEP + "/student-registration/student-enrollment-query/v2_0/" + uin + "/" + termid

	retValue := make([]model.Course, 0)

	campusData, statusCode, err := lv.getData(finalURL, accessToken, courseAuth)
	if err != nil {
		return nil, statusCode, err
	}

	if len(campusData.List) == 0 {
		return nil, 404, errors.New("no course data found")
	}

	if len(campusData.List[0].CourseRegistration) == 0 {
		return nil, 404, errors.New("no course data found")
	}

	for i := 0; i < len(campusData.List[0].CourseRegistration); i++ {
		course := campusData.List[0].CourseRegistration[i]
		if course.ValidRegistrationStatusType.Code == "R" {
			for i := 0; i < len(course.CourseSection.CourseSectionSession); i++ {
				retValue = append(retValue, *uiuc.NewCourse(course, i))
			}
		}
	}

	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	return &retValue, statusCode, nil
}

// GetTermSessions returns a list of term sessions to the client
func (lv StudentCourseAdapter) GetTermSessions() (*[4]model.TermSession, error) {
	var termSessions [4]model.TermSession
	//if beginning of June make fall term default, spring semester has ended
	crntDate := time.Now()
	crntYear := strconv.Itoa(crntDate.Year())
	if crntDate.Month() >= 6 && crntDate.Month() <= 12 {
		nextYear := strconv.Itoa(crntDate.Year() + 1)
		termSessions[2] = model.TermSession{Term: "Fall - " + crntYear, TermID: "1" + crntYear + "8", CurrentTerm: true}
		termSessions[1] = model.TermSession{Term: "Summer - " + crntYear, TermID: "1" + crntYear + "5", CurrentTerm: false}
		termSessions[0] = model.TermSession{Term: "Spring - " + crntYear, TermID: "1" + crntYear + "1", CurrentTerm: false}
		termSessions[3] = model.TermSession{Term: "Spring - " + nextYear, TermID: "1" + nextYear + "1", CurrentTerm: false}
	} else {
		pastYear := strconv.Itoa(crntDate.Year() - 1)
		termSessions[1] = model.TermSession{Term: "Spring - " + crntYear, TermID: "1" + crntYear + "1", CurrentTerm: true}
		termSessions[2] = model.TermSession{Term: "Summer - " + crntYear, TermID: "1" + crntYear + "5", CurrentTerm: false}
		termSessions[0] = model.TermSession{Term: "Fall - " + pastYear, TermID: "1" + pastYear + "8", CurrentTerm: false}
		termSessions[3] = model.TermSession{Term: "Fall - " + crntYear, TermID: "1" + crntYear + "8", CurrentTerm: false}
	}
	return &termSessions, nil
}

// GetGiesCourses returns a list of courses for the given GIES student
func (lv StudentCourseAdapter) GetGiesCourses(uin string, accessToken string, conf *model.EnvConfigData) (*[]model.GiesCourse, int, error) {

	GiesURL := conf.GiesCourseURL

	finalURL := GiesURL + "/" + uin

	retValue := make([]model.GiesCourse, 0)

	campusData, statusCode, err := lv.getGiesData(finalURL, accessToken)
	if err != nil {
		return nil, statusCode, err
	}

	if len(campusData) == 0 {
		return nil, 404, errors.New("no course data found")
	}

	for i := 0; i < len(campusData); i++ {
		course := campusData[i]
		retValue = append(retValue, *uiuc.NewGiesCourse(course))
	}

	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	return &retValue, statusCode, nil
}

func (lv StudentCourseAdapter) getData(targetURL string, accessToken string, authKey string) (*uiuc.CampusData, int, error) {
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, targetURL, nil)

	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	req.Header.Add("Authorization", "Bearer "+accessToken)
	req.Header.Set("Ocp-Apim-Subscription-Key", authKey)

	res, err := client.Do(req)
	if err != nil {
		return nil, res.StatusCode, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
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
		return nil, res.StatusCode, errors.New("bad request to api end point")
	}

	if res.StatusCode == 406 {
		return nil, res.StatusCode, errors.New("server returned 406: possible uin claim mismatch")
	}
	//campus api returns a 502 when there is no course data
	if res.StatusCode == 502 {
		return nil, 404, errors.New(res.Status)
	}

	if res.StatusCode == 200 || res.StatusCode == 203 {
		data := uiuc.CampusData{}

		err = json.Unmarshal(body, &data)

		if err != nil {
			return nil, res.StatusCode, err
		}
		return &data, res.StatusCode, nil
	}

	return nil, res.StatusCode, errors.New("Error making request: " + fmt.Sprint(res.StatusCode) + ": " + string(body))

}

func (lv StudentCourseAdapter) getGiesData(targetURL string, accessToken string) ([]uiuc.GIESCourse, int, error) {
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, targetURL, nil)

	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	req.Header.Add("access_token", accessToken)
	res, err := client.Do(req)
	if err != nil {
		return nil, res.StatusCode, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
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
		return nil, res.StatusCode, errors.New("bad request to api end point")
	}

	if res.StatusCode == 406 {
		return nil, res.StatusCode, errors.New("server returned 406: possible uin claim mismatch")
	}

	if res.StatusCode == 200 {
		data := make([]uiuc.GIESCourse, 0)

		err = json.Unmarshal(body, &data)

		if err != nil {
			return nil, res.StatusCode, err
		}
		return data, res.StatusCode, nil
	}

	return nil, res.StatusCode, errors.New("Error making request: " + fmt.Sprint(res.StatusCode) + ": " + string(body))

}
