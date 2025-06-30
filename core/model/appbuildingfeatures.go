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
	//TypeAppBuildingFeature type
	TypeAppBuildingFeature logutils.MessageDataType = "application building feature"
)

// AppBuildingFeature represents the configured features for campus buildings
type AppBuildingFeature struct {
	CampusName string `json:"campus_name" bson:"campus_name"`
	CampusCode string `json:"campus_code" bson:"campus_code"`
	AppName    string `json:"app_name" bson:"app_name"`
	AppCode    string `json:"app_code" bson:"app_code"`
	ShowInApp  bool   `json:"show_in_app" bson:"show_in_app"`
}
