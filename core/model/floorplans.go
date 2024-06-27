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
	//TypeFloorPlan type
	TypeFloorPlan logutils.MessageDataType = "bldg floor plan"
)

// FloorPlanMarker respresents a floor plan marker
type FloorPlanMarker struct {
	RenderID    string `json:"_id" bson:"_id"`
	Label       string `json:"label" bson:"label"`
	Description string `json:"description" bson:"description"`
	Display     string `json:"display" bson:"display"`
	Icon        string `json:"icon" bson:"icon"`
}

// FloorPlanHighlite represents a floor plan highlight
type FloorPlanHighlite struct {
	RenderID string `json:"_id" bson:"_id"`
	Label    string `json:"label" bson:"label"`
	Color    string `json:"color" bson:"color"`
	Display  string `json:"display" bson:"display"`
}

// FloorPlan represents a  floor plan object
type FloorPlan struct {
	BuildingNumber string              `json:"building_number" bson:"building_number"`
	BuildingFloor  string              `json:"building_floor" bson:"building_floor"`
	SVGEncoding    string              `json:"svg_encoding" bson:"svg_encoding"`
	SVG            string              `json:"svg" bson:"svg"`
	Markers        []FloorPlanMarker   `json:"markers" bson:"markers"`
	Highlites      []FloorPlanHighlite `json:"highlites" bson:"highlites"`
}
