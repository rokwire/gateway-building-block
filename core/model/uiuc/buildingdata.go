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

// UIUCFloorPlanServerResponse represents a UIUC, floorplan specific server response
type UIUCFloorPlanServerResponse struct {
	Status          string `json:"status"`
	HttpReturn      int    `json:"http_return"`
	Collection      string `json:"collection"`
	CountMarkers    int    `json:"count_markers"`
	CountHighlights int    `json:"count_highights"`
	CountResults    int    `json:"count_results"`
	Errors          string `json:"errors"`
	ErrorText       string `json:"error_text"`
}

// UIUCFloorPlanMarker respresents a UIUC floor plan marker
type UIUCFloorPlanMarker struct {
	RenderID    string `json:"render_id"`
	Label       string `json:"label"`
	Description string `json:"description"`
	Display     string `json:"display"`
	Icon        string `json:"icon"`
}

// UIUCFloorPlanHighlite represents a UIUC specific floor plan highlight
type UIUCFloorPlanHighlite struct {
	RenderID string `json:"render_id"`
	Label    string `json:"label"`
	Color    string `json:"color"`
	Display  string `json:"display"`
}

// UIUCFloorPlan represents a UIUC floor plan object
type UIUCFloorPlan struct {
	BuildingNumber string                  `json:"building_number"`
	BuildingFloor  string                  `json:"building_floor"`
	SVGEncoding    string                  `json:"svg_encoding"`
	SVG            string                  `json:"svg"`
	Markers        []UIUCFloorPlanMarker   `json:"markers"`
	Highlites      []UIUCFloorPlanHighlite `json:"highlites"`
}

// UIUCFloorPlanResult represents the full data returned from UIUC when querying a floorplan
type UIUCFloorPlanResult struct {
	Response UIUCFloorPlanServerResponse `json:"response"`
	Result   UIUCFloorPlan               `json:"result"`
}

// ServerLocationData respresnts a UIUC specific data structure for building location data
type ServerLocationData struct {
	Response  ServerResponse   `json:"response"`
	Buildings []CampusBuilding `json:"results"`
}

// NewFloorPlan creates a wayfinding floorplan instance from a UIUCFloorPlan instance
func NewFloorPlan(fp UIUCFloorPlan) *model.FloorPlan {
	newfp := model.FloorPlan{BuildingNumber: fp.BuildingNumber, BuildingFloor: fp.BuildingFloor, SVGEncoding: fp.SVGEncoding, SVG: fp.SVG}
	for i := 0; i < len(fp.Markers); i++ {
		newfp.Markers = append(newfp.Markers, model.FloorPlanMarker{RenderID: fp.Markers[i].RenderID, Label: fp.Markers[i].Label, Description: fp.Markers[i].Description,
			Display: fp.Markers[i].Display, Icon: fp.Markers[i].Icon})
	}
	for j := 0; j < len(fp.Highlites); j++ {
		newfp.Highlites = append(newfp.Highlites, model.FloorPlanHighlite{RenderID: fp.Highlites[j].RenderID, Label: fp.Highlites[j].Label, Color: fp.Highlites[j].Color,
			Display: fp.Markers[j].Display})
	}
	return &newfp
}

// NewBuilding creates a wayfinding.Building instance from a campusBuilding,
// including all active entrances for the building
func NewBuilding(bldg CampusBuilding) *model.Building {
	newBldg := model.Building{ID: bldg.UUID, Name: bldg.Name, ImageURL: bldg.ImageURL, Address1: bldg.Address1, Address2: bldg.Address2,
		FullAddress: bldg.FullAddress, City: bldg.City, ZipCode: bldg.ZipCode, State: bldg.State, Latitude: bldg.Latitude, Longitude: bldg.Longitude, Number: bldg.Number}
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
