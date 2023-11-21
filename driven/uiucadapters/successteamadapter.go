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
	"application/core/interfaces"
	model "application/core/model"
	uiuc "application/core/model/uiuc"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
)

// SuccessTeamAdapter is a vendor specific structure that implements the UIUC success team
type SuccessTeamAdapter struct {
	StorageAdapter interfaces.Storage
}

// NewSuccessTeamAdapter returns a vendor specific implementation of the success team adapter
func NewSuccessTeamAdapter(adapter interfaces.Storage) *SuccessTeamAdapter {
	successteamadapter := SuccessTeamAdapter{StorageAdapter: adapter}
	return &successteamadapter
}

// GetSuccessTeam returns a list of
func (sta SuccessTeamAdapter) GetSuccessTeam(uin string, unitid string, accessToken string, conf *model.EnvConfigData) (*[]model.SuccessTeamMember, int, error) {

	retValue := make([]model.SuccessTeamMember, 0)
	pcpdata, status, err := sta.GetPrimaryCareProvider(uin, accessToken, conf)

	if err != nil {
		return nil, status, err
	}

	retValue = append(retValue, *pcpdata...)

	advisordata, advstatus, adverr := sta.GetAcademicAdvisors(uin, unitid, accessToken, conf)
	if adverr != nil {
		return nil, advstatus, adverr
	}
	retValue = append(retValue, *advisordata...)
	return &retValue, 200, nil
}

// GetPrimaryCareProvider returns a list of
func (sta SuccessTeamAdapter) GetPrimaryCareProvider(uin string, accessToken string, conf *model.EnvConfigData) (*[]model.SuccessTeamMember, int, error) {

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

// GetAcademicAdvisors returns a list of
func (sta SuccessTeamAdapter) GetAcademicAdvisors(uin string, unitid string, accessToken string, conf *model.EnvConfigData) (*[]model.SuccessTeamMember, int, error) {

	baseURL := conf.EngAppointmentBaseURL
	finalURL := ""
	var headers = make(map[string]string)
	headers["Authorization"] = "Bearer " + accessToken

	calendars, err := sta.StorageAdapter.FindCalendars(unitid)

	if err != nil {
		return nil, 500, err
	}

	s := make([]model.SuccessTeamMember, 0)

	for j := 0; j < len(*calendars); j++ {
		unitCal := (*calendars)[j]
		finalURL = baseURL + "users/" + uin + "/calendars/" + strconv.Itoa(unitCal.CalendarID) + "/advisors"
		vendorData, err := sta.getAdvisorData(finalURL, "GET", headers, nil)
		if err != nil {
			return nil, 500, err
		}

		var advisors []uiuc.EngineeringAdvisor
		err = json.Unmarshal(vendorData, &advisors)
		if err != nil {
			return nil, 500, err
		}

		for i := 0; i < len(advisors); i++ {
			advisor := advisors[i]
			firstName := strings.TrimSpace(strings.Split(advisor.Name, ",")[0])
			lastName := strings.TrimSpace(strings.Split(advisor.Name, ",")[1])
			stm := model.SuccessTeamMember{FirstName: firstName, LastName: lastName, Email: "", ExternalLink: "", ExternalLinkText: "", Department: advisor.CalendarName, Title: "Academic Advisor", TeamMemberID: advisor.ID}
			s = append(s, stm)
		}

	}

	return &s, 200, nil
}

func (sta SuccessTeamAdapter) getPCPData(targetURL string, accessToken string) (*uiuc.PrimaryCareProvider, int, error) {
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

func (sta SuccessTeamAdapter) getAdvisorData(targetURL string, method string, headers map[string]string, postdata *strings.Reader) ([]byte, error) {

	client := &http.Client{}

	var postbody = io.Reader(nil)
	if postdata != nil {
		postbody = postdata
	}

	req, err := http.NewRequest(method, targetURL, postbody)

	if err != nil {
		return nil, err
	}

	for key, element := range headers {
		req.Header.Add(key, element)
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode == 400 {
		return nil, errors.New(res.Status + " : " + string(body))
	}
	if res.StatusCode == 401 {
		return nil, errors.New(res.Status + " : " + string(body))
	}

	if res.StatusCode == 403 {
		return nil, errors.New(res.Status + ": " + string(body))
	}

	if res.StatusCode == 406 {
		return nil, errors.New("server returned 406: possible uin claim mismatch")
	}

	if res.StatusCode == 409 {
		return nil, errors.New(res.Status + " : " + string(body))
	}

	if res.StatusCode == 200 || res.StatusCode == 201 || res.StatusCode == 204 {

		return body, nil
	}

	return nil, errors.New("error making request: " + fmt.Sprint(res.StatusCode) + ": " + string(body))
}
