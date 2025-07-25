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

	"github.com/rokwire/rokwire-building-block-sdk-go/utils/errors"
	"github.com/rokwire/rokwire-building-block-sdk-go/utils/logging/logutils"
)

const (
	// TypeConfig configs type
	TypeConfig logutils.MessageDataType = "config"
	// TypeConfigData config data type
	TypeConfigData logutils.MessageDataType = "config data"
	// TypeEnvConfigData env configs type
	TypeEnvConfigData logutils.MessageDataType = "env config data"

	// ConfigTypeEnv is the Config Type for EnvConfigData
	ConfigTypeEnv string = "env"
)

// Config contain generic configs
type Config struct {
	ID          string      `json:"id" bson:"_id"`
	Type        string      `json:"type" bson:"type"`
	AppID       string      `json:"app_id" bson:"app_id"`
	OrgID       string      `json:"org_id" bson:"org_id"`
	System      bool        `json:"system" bson:"system"`
	Data        interface{} `json:"data" bson:"data"`
	DateCreated time.Time   `json:"date_created" bson:"date_created"`
	DateUpdated *time.Time  `json:"date_updated" bson:"date_updated"`
}

// EnvConfigData contains environment configs for this service
type EnvConfigData struct {
	ExampleEnv              string `json:"example_env" bson:"example_env"`
	CentralCampusURL        string `json:"GATEWAY_CENTRALCAMPUS_ENDPOINT" bson:"GATEWAY_CENTRALCAMPUS_ENDPOINT"`
	CentralCampusKey        string `json:"GATEWAY_CENTRALCAMPUS_APIKEY" bson:"GATEWAY_CENTRALCAMPUS_APIKEY"`
	GiesCourseURL           string `json:"GATEWAY_GIESCOURSES_ENDPOINT" bson:"GATEWAY_GIESCOURSES_ENDPOINT"`
	WayFindingURL           string `json:"GATEWAY_WAYFINDING_APIURL" bson:"GATEWAY_WAYFINDING_APIURL"`
	WayFindingKey           string `json:"GATEWAY_WAYFINDING_APIKEY" bson:"GATEWAY_WAYFINDING_APIKEY"`
	LaundryViewURL          string `json:"GATEWAY_LAUNDRY_APIURL" bson:"GATEWAY_LAUNDRY_APIURL"`
	LaundryViewKey          string `json:"GATEWAY_LAUNDRY_APIKEY" bson:"GATEWAY_LAUNDRY_APIKEY"`
	LaundryServiceKey       string `json:"GATEWAY_LAUNDRYSERVICE_APIKEY" bson:"GATEWAY_LAUNDRYSERVICE_APIKEY"`
	LaundyrServiceURL       string `json:"GATEWAY_LAUNDRYSERVICE_API" bson:"GATEWAY_LAUNDRYSERVICE_API"`
	LaundryServiceBasicAuth string `json:"GATEWAY_LAUNDRYSERVICE_BASICAUTH" bson:"GATEWAY_LAUNDRYSERVICE_BASICAUTH"`
	EngAppointmentBaseURL   string `json:"GATEWAY_APPOINTMENTS_ENGURL" bson:"GATEWAY_APPOINTMENTS_ENGURL"`
	PCPEndpoint             string `json:"GATEWAY_STUDENTSUCCESS_PCPENDPOINT" bson:"GATEWAY_STUDENTSUCCESS_PCPENDPOINT"`
	ImageEndpoint           string `json:"GATEWAY_STUDENTSUCCESS_IMAGES" bson:"GATEWAY_STUDENTSUCCESS_IMAGES"`
}

// GetConfigData returns a pointer to the given config's Data as the given type T
func GetConfigData[T ConfigData](c Config) (*T, error) {
	if data, ok := c.Data.(T); ok {
		return &data, nil
	}
	return nil, errors.ErrorData(logutils.StatusInvalid, TypeConfigData, &logutils.FieldArgs{"type": c.Type})
}

// ConfigData represents any set of data that may be stored in a config
type ConfigData interface {
	EnvConfigData | map[string]interface{}
}
