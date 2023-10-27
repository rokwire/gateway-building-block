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

package uiuc

import (
	model "application/core/model"
)

// CampusEntrance representes a campus specific building entrance
type CampusEntrance struct {
	UUID         string  `json:"uuid"`
	Name         string  `json:"descriptive_name"`
	ADACompliant bool    `json:"is_ada_compliant"`
	Available    bool    `json:"is_available_for_use"`
	ImageURL     string  `json:"image"`
	Latitude     float64 `json:"latitude"`
	Longitude    float64 `json:"longitude"`
}

// CampusBuilding represents a campus specific building
type CampusBuilding struct {
	UUID        string           `json:"uuid"`
	Name        string           `json:"name"`
	Number      string           `json:"number"`
	FullAddress string           `json:"location"`
	Address1    string           `json:"address_1"`
	Address2    string           `json:"address_2"`
	City        string           `json:"city"`
	State       string           `json:"state"`
	ZipCode     string           `json:"zipcode"`
	ImageURL    string           `json:"image"`
	MailCode    string           `json:"mailcode"`
	Entrances   []CampusEntrance `json:"entrances"`
	Latitude    float64          `json:"building_centroid_latitude"`
	Longitude   float64          `json:"building_centroid_longitude"`
}

// ServerResponse represents a UIUC specific server response
type ServerResponse struct {
	Status         string `json:"status"`
	HTTPStatusCode int    `json:"http_return"`
	CollectionType string `json:"collection"`
	Count          int    `json:"count"`
	ErrorList      string `json:"errors"`
	ErrorMessage   string `json:"error_text"`
}

// ServerLocationData respresnts a UIUC specific data structure for building location data
type ServerLocationData struct {
	Response  ServerResponse   `json:"response"`
	Buildings []CampusBuilding `json:"results"`
}

// NewBuilding creates a wayfinding.Building instance from a campusBuilding,
// including all active entrances for the building
func NewBuilding(bldg CampusBuilding) *model.Building {
	newBldg := model.Building{ID: bldg.UUID, Name: bldg.Name, ImageURL: bldg.ImageURL, Address1: bldg.Address1, Address2: bldg.Address2, FullAddress: bldg.FullAddress, City: bldg.City, ZipCode: bldg.ZipCode, State: bldg.State, Latitude: bldg.Latitude, Longitude: bldg.Longitude}
	newBldg.Entrances = make([]model.Entrance, 0)
	for _, n := range bldg.Entrances {
		if n.Available {
			newBldg.Entrances = append(newBldg.Entrances, *NewEntrance(n))
		}
	}
	return &newBldg
}

// NewBuildingList returns a list of wayfinding buildings created frmo a list of campus building objects.
func NewBuildingList(bldgList *[]CampusBuilding) *[]model.Building {
	retList := make([]model.Building, len(*bldgList))
	for i := 0; i < len(*bldgList); i++ {
		cmpsBldg := (*bldgList)[i]
		crntBldng := NewBuilding(cmpsBldg)
		retList[i] = *crntBldng
	}
	return &retList
}

// NewEntrance creates a wayfinding.Entrance instance from a campusEntrance object
func NewEntrance(ent CampusEntrance) *model.Entrance {
	newEnt := model.Entrance{ID: ent.UUID, Name: ent.Name, ADACompliant: ent.ADACompliant, Available: ent.Available, ImageURL: ent.ImageURL, Latitude: ent.Latitude, Longitude: ent.Longitude}
	return &newEnt
}
