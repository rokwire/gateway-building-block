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
	//TypeBuilding type
	TypeBuilding logutils.MessageDataType = "building"
)

// Entrance represents the information returned when the closest entrance of a building is requested
type Entrance struct {
	ID           string
	Name         string
	ADACompliant bool
	Available    bool
	ImageURL     string
	Latitude     float64
	Longitude    float64
}

// Building represents the information returned when a requst for a building's details is made
type Building struct {
	ID          string
	Name        string
	Number      string
	FullAddress string
	Address1    string
	Address2    string
	City        string
	State       string
	ZipCode     string
	ImageURL    string
	MailCode    string
	Entrances   []Entrance
	Latitude    float64
	Longitude   float64
	Floors      []string
	Features    []BuildingFeatureLocation
}

// CompactBuilding represents minimal building informaiton needed to display a builgins details on the details panel
type CompactBuilding struct {
	ID          string
	Name        string
	Number      string
	FullAddress string
	ImageURL    string
	Latitude    float64
	Longitude   float64
	Features    []BuildingFeatureLocation
}

// BuildingFeatureLocation represents a list of where each feature belonging to a building can be found
type BuildingFeatureLocation struct {
	Key   string          `json:"key" bson:"key"`
	Value FeatureMapEntry `json:"value" bson:"value"`
}

// FeatureMapEntry represents the floor data associated with a feature key for a building
type FeatureMapEntry struct {
	Name   string   `json:"name" bson:"name"`
	Floors []string `json:"floors" bson:"floors"`
}

// BuildingFeature represents a feature found in buildings
type BuildingFeature struct {
	ID           string  `json:"_id" bson:"_id"`
	BuildingID   string  `json:"building_id" bson:"building_id"`
	EQIndicator  string  `json:"eq_indicator" bson:"eq_indicator"`
	Name         string  `json:"name" bson:"name"`
	FoundOnFloor string  `json:"found_on_floor" bson:"found_on_floor"`
	FoundInRoom  string  `json:"found_in_room" bson:"found_in_room"`
	IsADA        bool    `json:"is_ada" bson:"is_ada"`
	IsExternal   bool    `json:"is_external" bson:"is_external"`
	Comments     string  `json:"comments" bson:"comments"`
	Latitude     float64 `json:"latitude" bson:"latitude"`
	Longitude    float64 `json:"longitude" bson:"longitude"`
}
