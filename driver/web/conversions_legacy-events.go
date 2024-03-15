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

package web

import (
	"application/core/model"
	Def "application/driver/web/docs/gen"
)

// ContactsLegacy
func contactToDef(item Def.TpsReqCreateEventContact) model.ContactLegacy {

	return model.ContactLegacy{ContactName: *item.ContactName, ContactEmail: *item.ContactEmail, ContactPhone: *item.ContactPhone}
}

func contactsToDef(item []Def.TpsReqCreateEventContact) []model.ContactLegacy {
	result := make([]model.ContactLegacy, len(item))
	for i, item := range item {
		result[i] = contactToDef(item)
	}
	return result
}

// LocationLegacy
func locationToDef(item Def.TpsReqCreateEventLocation) model.LocationLegacy {

	return model.LocationLegacy{Latitude: float64(*item.Latitude), Longitude: float64(*item.Longitude), Description: *item.Description}
}

func locationsToDef(item []Def.TpsReqCreateEventLocation) []model.LocationLegacy {
	result := make([]model.LocationLegacy, len(item))
	for i, item := range item {
		result[i] = locationToDef(item)
	}
	return result
}
