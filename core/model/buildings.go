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
	"github.com/rokwire/logging-library-go/v2/logutils"
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
	Latitude     float32
	Longitude    float32
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
	Latitude    float32
	Longitude   float32
}
