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

// SuccessTeamAdapter is a vendor specific structure that implements the UIUC success team
type SuccessTeamAdapter struct {
}

// SuccessTeamAdapter returns a vendor specific implementation of the success team adapter
func NewSuccessTeamAdapter() SuccessTeamAdapter {
	return SuccessTeamAdapter{}

}

// GetSuccessTeam returns a list of
func (sta SuccessTeamAdapter) GetSuccessTeam(uin string, accessToken string, conf *model.EnvConfigData) (*[]model.SuccessTeamMember, int, error) {

	pcpURL := conf.PCPEndpoint + "/" + uin
	pcp, status, err := sta.getPCPData(pcpURL, accessToken)

	if err != nil {
		return nil, status, err
	}

	teamMember, err := uiuc.NewSuccessTeamMember(pcp)

	retValue := make([]model.SuccessTeamMember, 0)
	retValue = append(retValue, *teamMember)

	if err != nil {
		return nil, 500, err
	}
	return &retValue, status, nil
}

func (lv SuccessTeamAdapter) getPCPData(targetURL string, accessToken string) (*uiuc.PrimaryCareProvider, int, error) {
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, targetURL, nil)

	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	req.Header.Add("access_token", accessToken)

	res, err := client.Do(req)
	if err != nil {
		if res == nil {
			return nil, 500, err
		}
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
		data := uiuc.PrimaryCareProvider{}

		err = json.Unmarshal(body, &data)

		if err != nil {
			return nil, res.StatusCode, err
		}
		return &data, res.StatusCode, nil
	}

	return nil, res.StatusCode, errors.New("Error making request: " + fmt.Sprint(res.StatusCode) + ": " + string(body))

}
