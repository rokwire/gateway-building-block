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

import model "application/core/model"

// PrimaryCareProvider represents the full data returned by the campus primary care provider end point
type PrimaryCareProvider struct {
	StudentUIN      string `json:"StudentUIN"`
	ProviderFName   string `json:"ProviderFName"`
	ProviderLName   string `json:"ProviderLName"`
	ProviderUIN     string `json:"ProviderUIN"`
	Image           string `json:"Image"`
	AppointmentLink string `json:"AppointmentLink"`
	LinkText        string `json:"LinkText"`
	Department      string `json:"Department"`
	Title           string `json:"Title"`
}

// NewSuccessTeamMember constructs an app formatted SuccessTeamMember from a PrimaryCareProvider
func NewSuccessTeamMember(pcp *PrimaryCareProvider) (*model.SuccessTeamMember, error) {

	ret := model.SuccessTeamMember{
		FirstName:        pcp.ProviderFName,
		LastName:         pcp.ProviderLName,
		Email:            "",
		Image:            pcp.Image,
		Department:       pcp.Department,
		Title:            pcp.Title,
		ExternalLink:     pcp.AppointmentLink,
		ExternalLinkText: pcp.LinkText,
	}
	return &ret, nil
}
