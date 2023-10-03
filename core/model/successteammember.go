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

package model

import (
	"github.com/rokwire/logging-library-go/v2/logutils"
)

const (
	//TypeSuccessTeamMember type
	TypeSuccessTeamMember logutils.MessageDataType = "success team"
)

// SuccessTeamMember represents a primary care provider, academic advisor, any staff the student might interact with
type SuccessTeamMember struct {
	FirstName        string `json:"first_name" bson:"first_name"`
	LastName         string `json:"last_name" bson:"last_name"`
	Email            string `json:"email" bson:"email"`
	Image            string `json:"image" bson:"image"`
	Department       string `json:"department" bson:"department"`
	Title            string `json:"title" bson:"title"`
	ExternalLink     string `json:"external_link" bson:"external_link"`
	ExternalLinkText string `json:"external_link_text" bson:"external_link_text"`
}
