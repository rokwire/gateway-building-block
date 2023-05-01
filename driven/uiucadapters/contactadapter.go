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
)

// UIUCContactAdapter is a vendor specific structure that implements the contanct information interface
type UIUCContactAdapter struct {
}

// NewUIUCContactAdapter returns a vendor specific implementation of the contanct information interface
func NewUIUCContactAdapter() UIUCContactAdapter {
	return UIUCContactAdapter{}

}

// GetContactInformation returns a contact information object for a student
func (lv UIUCContactAdapter) GetContactInformation(uin string, accessToken string, mode string, conf *model.EnvConfigData) (*model.Person, int, error) {

	campusAPI := conf.CentralCampusURL
	campusKey := conf.CentralCampusKey
	finalURL := campusAPI + "/person/contact-summary-query/" + uin

	if mode != "0" {
		finalURL = campusAPI + "/mock/123456789"
	}

	campusData, statusCode, err := lv.getData(finalURL, accessToken, campusKey)
	if err != nil {
		return nil, statusCode, err
	}

	if len(campusData.People) == 0 {
		return nil, 404, errors.New("no contact data found")
	}

	retValue, err := uiuc.NewPerson(&campusData.People[0])
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	return retValue, statusCode, nil
}

func (lv UIUCContactAdapter) getData(targetURL string, accessToken string, apikey string) (*uiuc.CampusUserData, int, error) {
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, targetURL, nil)

	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	req.Header.Add("Authorization", "Bearer "+accessToken)
	req.Header.Set("Ocp-Apim-Subscription-Key", apikey)
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

	//campus api returns a 502 when there is no banner contact data for the uin
	if res.StatusCode == 502 {
		return nil, 404, errors.New(res.Status)
	}
	if res.StatusCode == 200 {
		data := uiuc.CampusUserData{}
		err = json.Unmarshal(body, &data)

		if err != nil {
			return nil, res.StatusCode, err
		}
		return &data, res.StatusCode, nil
	}

	return nil, res.StatusCode, errors.New("Error making request: " + fmt.Sprint(res.StatusCode) + ": " + string(body))

}
