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

package model

import (
	"time"

	"github.com/rokwire/logging-library-go/v2/logutils"
)

const (
	//TypeUnitCalendar Orgazational Calendar
	TypeUnitCalendar logutils.MessageDataType = "unit calendar"
)

// UnitCalendar is a container for mapping units to external system calendar ids
type UnitCalendar struct {
	ID          string     `json:"id" bson:"_id"`
	OrgID       string     `json:"org_id" bson:"org_id"`
	AppID       string     `json:"app_id" bson:"app_id"`
	CalendarID  int        `json:"calendar_id" bson:"calendar_id"`
	UnitID      int        `json:"unit_id" bson:"unit_id"`
	UnitName    string     `json:"unit_name" bson:"unit_name"`
	CollegeCode string     `json:"college_code" bson:"college_code"`
	CollegeName string     `json:"college_name" bson:"college_name"`
	DateCreated time.Time  `json:"date_created" bson:"date_created"`
	DateUpdated *time.Time `json:"date_updated" bson:"date_updated"`
}
