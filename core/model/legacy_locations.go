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

import "github.com/google/uuid"

// LegacyLocation wrapper
type LegacyLocation struct {
	ID          string   `json:"id" bson:"_id"`
	Name        string   `json:"name" bson:"name"`
	Description string   `json:"description" bson:"description"`
	Lat         *float64 `json:"lat" json:"lat"`
	Long        *float64 `json:"long" json:"long"`
}

func floatToPointer(val float64) *float64 {
	return &val
}

// LegacyLocationsListType default locations type def
type LegacyLocationsListType []LegacyLocation

// ToBsonRecords To bson objects
func (p LegacyLocationsListType) ToBsonRecords() []interface{} {
	list := make([]interface{}, len(p))
	for index := range p {
		list[index] = p[index]
	}
	return list
}

// ToNameMapping To name obejct mapping
func (p LegacyLocationsListType) ToNameMapping() map[string]LegacyLocation {
	mapping := map[string]LegacyLocation{}
	for index := range p {
		mapping[p[index].Name] = p[index]
	}
	return mapping
}

// DefaultLegacyLocations default locations collection for initialization step
var DefaultLegacyLocations = LegacyLocationsListType{
	{
		ID:          uuid.NewString(),
		Name:        "2700 Campus Way 45221",
		Lat:         floatToPointer(39.131894),
		Long:        floatToPointer(-84.519143),
		Description: "2700 Campus Way 45221",
	},
	{
		ID:          uuid.NewString(),
		Name:        "Davenport 109A",
		Lat:         floatToPointer(40.107335),
		Long:        floatToPointer(-88.226069),
		Description: "Davenport Hall Room 109A",
	},
	{
		ID:          uuid.NewString(),
		Name:        "Nevada Dance Studio (905 W. Nevada St.)",
		Lat:         floatToPointer(40.105825),
		Long:        floatToPointer(-88.219873),
		Description: "Nevada Dance Studio, 905 W. Nevada St.",
	},
	{
		ID:          uuid.NewString(),
		Name:        "18th Ave Library, 175 W 18th Ave, Room 205, Oklahoma City, OK",
		Lat:         floatToPointer(36.102183),
		Long:        floatToPointer(-97.111245),
		Description: "18th Ave Library, 175 W 18th Ave, Room 205, Oklahoma City, OK",
	},
	{
		ID:          uuid.NewString(),
		Name:        "Champaign County Fairgrounds",
		Lat:         floatToPointer(40.1202191),
		Long:        floatToPointer(-88.2178757),
		Description: "Champaign County Fairgrounds",
	},
	{
		ID:          uuid.NewString(),
		Name:        "Student Union SLC Conference room",
		Lat:         floatToPointer(39.727282),
		Long:        floatToPointer(-89.617477),
		Description: "Student Union SLC Conference room",
	},
	{
		ID:          uuid.NewString(),
		Name:        "Student Union SLC Conference Room",
		Lat:         floatToPointer(39.727282),
		Long:        floatToPointer(-89.617477),
		Description: "Student Union SLC Conference Room",
	},
	{
		ID:          uuid.NewString(),
		Name:        "Armory, room 172 (the Innovation Studio)",
		Lat:         floatToPointer(40.104749),
		Long:        floatToPointer(-88.23195),
		Description: "Armory, room 172 (the Innovation Studio)",
	},
	{
		ID:          uuid.NewString(),
		Name:        "Student Union Room 235",
		Lat:         floatToPointer(39.727282),
		Long:        floatToPointer(-89.617477),
		Description: "Student Union Room 235",
	},
	{
		ID:          uuid.NewString(),
		Name:        "Uni 206, 210, 211",
		Lat:         floatToPointer(40.11314),
		Long:        floatToPointer(-88.225259),
		Description: "Uni 206, 210, 211",
	},
	{
		ID:          uuid.NewString(),
		Name:        "Uni 205, 206, 210",
		Lat:         floatToPointer(40.11314),
		Long:        floatToPointer(-88.225259),
		Description: "Uni 205, 206, 210",
	},
	{
		ID:          uuid.NewString(),
		Name:        "Southern Historical Association Combs Chandler 30",
		Lat:         floatToPointer(38.258116),
		Long:        floatToPointer(-85.756139),
		Description: "Southern Historical Association Combs Chandler 30",
	},
	{
		ID:          uuid.NewString(),
		Name:        "St. Louis, MO",
		Lat:         floatToPointer(38.694237),
		Long:        floatToPointer(-90.4493),
		Description: "St. Louis, MO",
	},
	{
		ID:          uuid.NewString(),
		Name:        "Student Union SLC",
		Lat:         floatToPointer(39.727282),
		Long:        floatToPointer(-89.617477),
		Description: "Student Union SLC",
	},
	{
		ID:          uuid.NewString(),
		Name:        "Purdue University, West Lafayette, Indiana",
		Lat:         floatToPointer(40.425012),
		Long:        floatToPointer(-86.912645),
		Description: "Purdue University, West Lafayette, Indiana",
	},
	{
		ID:          uuid.NewString(),
		Name:        "MP 7",
		Lat:         floatToPointer(40.100803),
		Long:        floatToPointer(-88.23604),
		Description: "ARC MP 7",
	},
	{
		ID:          uuid.NewString(),
		Name:        "116 Roger Adams Lab",
		Lat:         floatToPointer(40.107741),
		Long:        floatToPointer(-88.224943),
		Description: "116 Roger Adams Lab",
	},
}
