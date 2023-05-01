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
	//TypeExample example type
	TypeExample logutils.MessageDataType = "example"
)

// Example is a generic Example data type
type Example struct {
	ID          string     `json:"id" bson:"_id"`
	OrgID       string     `json:"org_id" bson:"org_id"`
	AppID       string     `json:"app_id" bson:"app_id"`
	Data        string     `json:"data" bson:"data"`
	DateCreated time.Time  `json:"date_created" bson:"date_created"`
	DateUpdated *time.Time `json:"date_updated" bson:"date_updated"`
}
