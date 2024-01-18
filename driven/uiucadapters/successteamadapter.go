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
	"strings"
)

// SuccessTeamAdapter is a vendor specific structure that implements the UIUC success team
type SuccessTeamAdapter struct {
}

// NewSuccessTeamAdapter returns a vendor specific implementation of the success team adapter
func NewSuccessTeamAdapter() *SuccessTeamAdapter {
	successteamadapter := SuccessTeamAdapter{}
	return &successteamadapter
}

// GetSuccessTeam returns a list of
func (sta SuccessTeamAdapter) GetSuccessTeam(uin string, calendars *[]model.UnitCalendar, accessToken string, conf *model.EnvConfigData) (*model.SuccessTeam, int, error) {

	var retValue model.SuccessTeam

	pcpdata, status, err := sta.GetPrimaryCareProvider(uin, accessToken, conf)

	if err != nil {
		return nil, status, err
	}

	retValue.PrimaryCareProviders = *pcpdata

	advisordata, advstatus, adverr := sta.GetAcademicAdvisors(uin, calendars, accessToken, conf)
	if adverr != nil {
		return nil, advstatus, adverr
	}

	if len(*advisordata) == 0 {
		retValue.AcademicAdvisors = nil
	}
	retValue.AcademicAdvisors = *advisordata
	return &retValue, 200, nil
}

// GetPrimaryCareProvider returns a list of
func (sta SuccessTeamAdapter) GetPrimaryCareProvider(uin string, accessToken string, conf *model.EnvConfigData) (*[]model.SuccessTeamMember, int, error) {

	pcpURL := conf.PCPEndpoint + "/" + uin
	var headers = make(map[string]string)
	headers["access_token"] = accessToken

	pcp, status, err := sta.getPCPData(pcpURL, "GET", headers)

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
func (sta SuccessTeamAdapter) GetAcademicAdvisors(uin string, calendars *[]model.UnitCalendar, accessToken string, conf *model.EnvConfigData) (*[]model.SuccessTeamMember, int, error) {

	baseURL := conf.EngAppointmentBaseURL
	finalURL := ""
	imagebaseURL := conf.ImageEndpoint
	var headers = make(map[string]string)
	headers["Authorization"] = "Bearer " + accessToken

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

			advisorimage, _ := sta.getAdvisorImage(imagebaseURL+"/"+advisor.UIN, accessToken)
			stm := model.SuccessTeamMember{FirstName: firstName, LastName: lastName, Email: "", ExternalLink: "", ExternalLinkText: "", Department: advisor.CalendarName, Title: "Academic Advisor", TeamMemberID: advisor.ID, Image: advisorimage}
			s = append(s, stm)
		}

	}

	return &s, 200, nil
}

func (sta SuccessTeamAdapter) getPCPData(targetURL string, method string, headers map[string]string) (*uiuc.PrimaryCareProvider, int, error) {

	body, status, err := sta.getCampusData(targetURL, method, headers, nil)
	if err != nil {
		return nil, 500, err
	}

	if err != nil {
		return nil, status.StatusCode, err
	}

	if status.StatusCode == 401 {
		return nil, status.StatusCode, errors.New(status.StatusMessage)
	}

	if status.StatusCode == 403 {
		return nil, status.StatusCode, errors.New(status.StatusMessage)
	}

	if status.StatusCode == 400 {
		return nil, status.StatusCode, errors.New("bad request to api endpoint")
	}

	if status.StatusCode == 406 {
		return nil, status.StatusCode, errors.New("server returned 406: possible uin claim mismatch")
	}

	if status.StatusCode == 200 || status.StatusCode == 203 {
		data := uiuc.PrimaryCareProvider{}
		err = json.Unmarshal(body, &data)

		if err != nil {
			return nil, status.StatusCode, err
		}
		return &data, status.StatusCode, nil
	}

	return nil, status.StatusCode, errors.New("Error making request: " + fmt.Sprint(status.StatusCode) + ": " + string(body))

}

func (sta SuccessTeamAdapter) getAdvisorImage(url string, accesstoken string) (string, error) {
	var headers = make(map[string]string)
	headers["access_token"] = accesstoken

	body, status, err := sta.getCampusData(url, "POST", headers, nil)
	if err != nil {
		return "", err
	}

	if status.StatusCode == 400 {
		return "", errors.New(status.StatusMessage + " : " + string(body))
	}
	if status.StatusCode == 401 {
		return "", errors.New(status.StatusMessage + " : " + string(body))
	}

	if status.StatusCode == 403 {
		return "", errors.New(status.StatusMessage + ": " + string(body))
	}

	if status.StatusCode == 404 {
		return "", nil
	}

	if status.StatusCode == 200 || status.StatusCode == 201 || status.StatusCode == 204 {

		return string(body), nil
	}

	return "", errors.New("error making request: " + fmt.Sprint(status.StatusCode) + ": " + string(body))
}

func (sta SuccessTeamAdapter) getAdvisorData(targetURL string, method string, headers map[string]string, postdata *strings.Reader) ([]byte, error) {

	body, status, err := sta.getCampusData(targetURL, method, headers, postdata)
	if err != nil {
		return nil, err
	}

	if status.StatusCode == 400 {
		return nil, errors.New(status.StatusMessage + " : " + string(body))
	}
	if status.StatusCode == 401 {
		return nil, errors.New(status.StatusMessage + " : " + string(body))
	}

	if status.StatusCode == 403 {
		return nil, errors.New(status.StatusMessage + ": " + string(body))
	}

	if status.StatusCode == 406 {
		return nil, errors.New("server returned 406: possible uin claim mismatch")
	}

	if status.StatusCode == 409 {
		return nil, errors.New(status.StatusMessage + " : " + string(body))
	}

	if status.StatusCode == 200 || status.StatusCode == 201 || status.StatusCode == 204 {

		return body, nil
	}

	return nil, errors.New("error making request: " + fmt.Sprint(status.StatusCode) + ": " + string(body))
}

func (sta SuccessTeamAdapter) getCampusData(targetURL string, method string, headers map[string]string, postdata *strings.Reader) ([]byte, returnStatus, error) {
	client := &http.Client{}

	var postbody = io.Reader(nil)
	if postdata != nil {
		postbody = postdata
	}

	req, err := http.NewRequest(method, targetURL, postbody)

	rs := returnStatus{StatusCode: 0, StatusMessage: ""}
	if err != nil {

		rs.StatusCode = 500
		rs.StatusMessage = "Internal Server Error"
		return nil, rs, err
	}

	for key, element := range headers {
		req.Header.Add(key, element)
	}

	res, err := client.Do(req)
	if err != nil {
		rs.StatusCode = 500
		rs.StatusMessage = "Internal Server Error"
		return nil, rs, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		rs.StatusCode = 500
		rs.StatusMessage = "Internal Server Error"
		return nil, rs, err
	}

	rs.StatusCode = res.StatusCode
	rs.StatusMessage = res.Status
	return body, rs, nil

}

type returnStatus struct {
	StatusCode    int
	StatusMessage string
}
