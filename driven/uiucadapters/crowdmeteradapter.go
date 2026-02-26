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
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

// UIUCCrowdMeterAdapter is a vendor specific structure that implements the crowdmeter interface
type UIUCCrowdMeterAdapter struct {
}

// NewUIUCCrowdMeterAdapter returns a vendor specific implementation of the crowdmeter interface
func NewUIUCCrowdMeterAdapter() UIUCCrowdMeterAdapter {
	return UIUCCrowdMeterAdapter{}

}

// GetCrowdData returns all crowd meter data
func (cm UIUCCrowdMeterAdapter) GetCrowdData(conf *model.EnvConfigData) (*[]model.Crowd, error) {
	if conf == nil {
		return nil, errors.New("missing environment config")
	}

	body, statusCode, err := cm.getData(conf.CrowdMeterURL)
	if err != nil {
		return nil, err
	}

	if statusCode != http.StatusOK {
		return nil, errors.New("unexpected status: " + fmt.Sprint(statusCode))
	}

	var data []model.Crowd
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err
	}
	return &data, nil
}

func (cm UIUCCrowdMeterAdapter) getData(targetURL string) ([]byte, int, error) {
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, targetURL, nil)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, res.StatusCode, err
	}

	if res.StatusCode == http.StatusUnauthorized {
		return nil, res.StatusCode, errors.New(res.Status)
	}
	if res.StatusCode == http.StatusForbidden {
		return nil, res.StatusCode, errors.New(res.Status)
	}
	if res.StatusCode == http.StatusBadRequest {
		return nil, res.StatusCode, errors.New("bad request to api end point")
	}

	if res.StatusCode == http.StatusOK {
		return body, res.StatusCode, nil
	}

	return nil, res.StatusCode, errors.New("error making request: " + fmt.Sprint(res.StatusCode) + ": " + string(body))
}
