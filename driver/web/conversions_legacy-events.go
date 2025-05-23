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

// LegacyEventItem

func legacyEventItemToDef(item model.LegacyEventItem) Def.LegacyEventItem {

	status := legacyEventStatusToDef(item.Status)
	legacyEvent := legacyEventToDef(item.Item)
	return Def.LegacyEventItem{Source: item.SyncProcessSource,
		Status: status, LegacyEvent: legacyEvent}
}

func legacyEventsItemsToDef(items []model.LegacyEventItem) []Def.LegacyEventItem {
	result := make([]Def.LegacyEventItem, len(items))
	for i, item := range items {
		result[i] = legacyEventItemToDef(item)
	}
	return result
}

// LegacyEvent

func legacyEventToDef(item model.LegacyEvent) Def.LegacyEvent {
	return Def.LegacyEvent{
		AllDay:                  item.AllDay,
		CalendarId:              item.CalendarID,
		Category:                item.Category,
		Cost:                    item.Cost,
		CreatedBy:               item.CreatedBy,
		DataModified:            item.DataModified,
		DataSourceEventId:       item.DataSourceEventID,
		DateCreated:             item.DateCreated,
		EndDate:                 item.EndDate,
		EventId:                 item.EventID,
		IcalUrl:                 item.IcalURL,
		Id:                      item.ID,
		ImageUrl:                item.ImageURL,
		IsEventFree:             item.IsEventFree,
		IsSuperEvent:            item.IsSuperEvent,
		IsVirtual:               item.IsVirtial,
		LongDescription:         item.LongDescription,
		OriginatingCalendarId:   item.OriginatingCalendarID,
		OriginatingCalendarName: item.OriginatingCalendarName,
		OutlookUrl:              item.OutlookURL,
		RecurrenceId:            item.RecurrenceID,
		RecurringFlag:           item.RecurringFlag,
		RegistrationUrl:         item.RegistrationURL,
		SourceId:                item.SourceID,
		Sponsor:                 item.Sponsor,
		StartDate:               item.StartDate,
		Subcategory:             item.Subcategory,
		Tags:                    item.Tags,
		TargetAudience:          item.TargetAudience,
		Title:                   item.Title,
		TitleUrl:                item.TitleURL,
	}
}

// LegacyEventStatus

func legacyEventStatusToDef(item model.LegacyEventStatus) Def.LegacyEventStatus {
	return Def.LegacyEventStatus{Name: item.Name, ReasonIgnored: item.ReasonIgnored}
}

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

	return model.LocationLegacy{Latitude: float64(*item.Latitude), Longitude: float64(*item.Longitude), Description: *item.Description,
		Address: *item.Address, Building: *item.Building, Floor: *item.Floor, Room: *item.Room}
}

func locationsToDef(item []Def.TpsReqCreateEventLocation) []model.LocationLegacy {
	result := make([]model.LocationLegacy, len(item))
	for i, item := range item {
		result[i] = locationToDef(item)
	}
	return result
}
