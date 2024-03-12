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

package storage

import (
	"application/core/model"

	"github.com/google/uuid"
)

func consLegacyEvent(g model.LegacyEventItem) model.LegacyEventItem {
	var id string
	if g.Item.ID != "" {
		id = g.Item.ID
	} else {
		id = uuid.NewString()

	}
	return model.LegacyEventItem{SyncProcessSource: g.SyncProcessSource, SyncDate: g.SyncDate,
		Item: model.LegacyEvent{ID: id, AllDay: g.Item.AllDay, CalendarID: g.Item.CalendarID, Category: g.Item.Category,
			Subcategory: g.Item.Subcategory, CreatedBy: g.Item.CreatedBy, LongDescription: g.Item.LongDescription,
			DataModified: g.Item.DataModified, DataSourceEventID: g.Item.DataSourceEventID, DateCreated: g.Item.DateCreated,
			EndDate: g.Item.EndDate, EventID: g.Item.EventID, IcalURL: g.Item.IcalURL, IsEventFree: g.Item.IsEventFree,
			IsVirtial: g.Item.IsVirtial, Location: g.Item.Location, OriginatingCalendarID: g.Item.OriginatingCalendarID,
			OutlookURL: g.Item.OutlookURL, RecurrenceID: g.Item.RecurrenceID, IsSuperEvent: g.Item.IsSuperEvent,
			RecurringFlag: g.Item.RecurringFlag, SourceID: g.Item.SourceID, Sponsor: g.Item.Sponsor, StartDate: g.Item.StartDate,
			Title: g.Item.Title, TitleURL: g.Item.TitleURL, RegistrationURL: g.Item.RegistrationURL, Contacts: g.Item.Contacts, SubEvents: g.Item.SubEvents}}
}

func legacyEventFromStorage(item legacyEvent) model.LegacyEvent {
	return model.LegacyEvent{AllDay: item.AllDay, CalendarID: item.CalendarID, Category: item.Category,
		Subcategory: item.Subcategory, CreatedBy: item.CreatedBy, LongDescription: item.LongDescription, DataModified: item.DataModified,
		DataSourceEventID: item.DataSourceEventID, DateCreated: item.DateCreated, EndDate: item.EndDate, EventID: item.EventID,
		IcalURL: item.IcalURL, ImageURL: item.ImageURL, IsEventFree: item.IsEventFree, IsVirtial: item.IsVirtial, /*Location*/
		OriginatingCalendarID: item.OriginatingCalendarID, OutlookURL: item.OutlookURL, RecurrenceID: item.RecurrenceID,
		IsSuperEvent: item.IsSuperEvent, RecurringFlag: item.RecurringFlag, SourceID: item.SourceID, Sponsor: item.Sponsor,
		StartDate: item.StartDate, Title: item.Title, //TitleURL: item.TitleURL, RegistrationURL: item.RegistrationURL,
		/*SubEvents*/}
}

func legacyEventsFromStorage(itemsList []legacyEvent) []model.LegacyEvent {
	result := make([]model.LegacyEvent, len(itemsList))
	for index, item := range itemsList {
		result[index] = legacyEventFromStorage(item)
	}
	return result
}
