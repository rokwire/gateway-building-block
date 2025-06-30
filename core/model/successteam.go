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
	"github.com/rokwire/rokwire-building-block-sdk-go/utils/logging/logutils"
)

const (
	//TypeSuccessTeam type
	TypeSuccessTeam logutils.MessageDataType = "success team"
)

// SuccessTeam represents a primary care provider, academic advisor, any staff the student might interact with
type SuccessTeam struct {
	PrimaryCareProviders []SuccessTeamMember `json:"primary_care_providers" bson:"primary_care_providers"`
	AcademicAdvisors     []SuccessTeamMember `json:"academic_advisors" bson:"academic_advisors"`
}
