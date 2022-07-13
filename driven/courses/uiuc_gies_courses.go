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
)

type course struct {
	Term       string `json:"Term"`
	Subject    string `json:"Subject"`
	Number     string `json:"Number"`
	Section    string `json:"Section"`
	Title      string `json:"Title"`
	Instructor string `json:"Instructor"`
}

//GiesCourseAdapter is a vendor specific structure that implements the GiesCourse lookup interface
type GiesCourseAdapter struct {
	APIEndpoint string
}

//NewGiesCourseAdapter returns a vendor specific implementation of the Course lookup interface
func NewGiesCourseAdapter(url string) *GiesCourseAdapter {
	return &GiesCourseAdapter{APIEndpoint: url}

}

func newCourse(cr course) *model.GiesCourse {
	ret := model.GiesCourse{}
	ret.Instructor = cr.Instructor
	ret.Number = cr.Number
	ret.Section = cr.Section
	ret.Subject = cr.Subject
	ret.Term = cr.Term
	ret.Title = cr.Title
	return &ret
}

//GetStudentCourses returns a list of courses for the given GIES student
func (lv *GiesCourseAdapter) GetStudentCourses(uin string, accessToken string) (*[]model.GiesCourse, int, error) {

	finalURL := lv.APIEndpoint + "/" + uin

	retValue := make([]model.GiesCourse, 0)

	campusData, statusCode, err := lv.getData(finalURL, accessToken)
	if err != nil {
		return nil, statusCode, err
	}

	if len(campusData) == 0 {
		return nil, 404, errors.New("No course data found")
	}

	for i := 0; i < len(campusData); i++ {
		course := campusData[i]
		retValue = append(retValue, *newCourse(course))
	}

	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	return &retValue, statusCode, nil
}

func (lv *GiesCourseAdapter) getData(targetURL string, accessToken string) ([]course, int, error) {
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, targetURL, nil)

	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	req.Header.Add("Authorization", "Bearer "+accessToken)
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

	if res.StatusCode == 200 {
		data := make([]course, 0)

		err = json.Unmarshal(body, &data)

		if err != nil {
			return nil, res.StatusCode, err
		}
		return data, res.StatusCode, nil
	}

	return nil, res.StatusCode, errors.New("Error making request: " + fmt.Sprint(res.StatusCode) + ": " + string(body))

}
