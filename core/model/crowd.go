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
	"github.com/rokwire/rokwire-building-block-sdk-go/utils/logging/logutils"
)

const (
	//TypeCrowd type
	TypeCrowd logutils.MessageDataType = "crowd"
)

// Crowd represents the busy-level schedule for a location.
type Crowd struct {
	CrowdType    string `json:"CrowdType"`
	LocationID   int    `json:"LocationID"`
	LocationName string `json:"LocationName"`
	Days         []Day  `json:"days"`
}

// Day holds busy levels for each hour for a named day.
type Day struct {
	BusyLevels []int  `json:"BusyLevels" bson:"busy_levels"`
	Day        string `json:"Day" bson:"day"`
}
