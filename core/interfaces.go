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

package core

import (
	"application/core/model"
	"application/driven/storage"
	"time"

	"github.com/rokwire/core-auth-library-go/v3/tokenauth"
)

// Default exposes client APIs for the driver adapters
type Default interface {
	GetVersion() string
}

// Client exposes client APIs for the driver adapters
type Client interface {
	GetExample(orgID string, appID string, id string) (*model.Example, error)
	GetUnitCalendars(id string) (*[]model.UnitCalendar, error)
	ListLaundryRooms() (*model.Organization, error)
	GetLaundryRoom(roomid string) (*model.RoomDetail, error)
	InitServiceRequest(machineid string) (*model.MachineRequestDetail, error)
	SubmitServiceRequest(machineID string, problemCode string, comments string, firstname string, lastname string, phone string, email string) (*model.ServiceRequestResult, error)
	GetBuilding(bldgID string, adaOnly bool, latitude float64, longitude float64) (*model.Building, error)
	GetEntrance(bldgID string, adaOnly bool, latitude float64, longitude float64) (*model.Entrance, error)
	GetBuildings() (*[]model.Building, error)
	GetContactInfo(uin string, accessToken string, mode string) (*model.Person, int, error)
	GetGiesCourses(uin string, accessToken string) (*[]model.GiesCourse, int, error)
	GetStudentCourses(uin string, termid string, accessToken string) (*[]model.Course, int, error)
	GetTermSessions() (*[4]model.TermSession, error)
	GetSuccessTeam(uin string, unitid string, accessToken string) (*model.SuccessTeam, int, error)
	GetPrimaryCareProvider(uin string, accessToken string) (*[]model.SuccessTeamMember, int, error)
	GetAcademicAdvisors(uin string, unitid string, accessToken string) (*[]model.SuccessTeamMember, int, error)
	GetFloorPlan(buildingnumber string, floornumber string, markers string, highlites string) (*model.FloorPlan, int, error)
	SearchBuildings(bldgName string, returnCompact bool) (*map[string]any, error)
}

// Admin exposes administrative APIs for the driver adapters
type Admin interface {
	GetExample(orgID string, appID string, id string) (*model.Example, error)
	CreateExample(example model.Example) (*model.Example, error)
	UpdateExample(example model.Example) error
	AppendExample(example model.Example) (*model.Example, error)
	DeleteExample(orgID string, appID string, id string) error

	GetConfig(id string, claims *tokenauth.Claims) (*model.Config, error)
	GetConfigs(configType *string, claims *tokenauth.Claims) ([]model.Config, error)
	CreateConfig(config model.Config, claims *tokenauth.Claims) (*model.Config, error)
	UpdateConfig(config model.Config, claims *tokenauth.Claims) error
	DeleteConfig(id string, claims *tokenauth.Claims) error
	AddWebtoolsBlackList(dataSourceIDs []string, dataCalendarIDs []string, dataOriginatingCalendarIDs []string) error
	GetWebtoolsBlackList() ([]model.WebToolsItem, error)
	RemoveWebtoolsBlackList(sourceids []string, calendarids []string, originatingCalendarIdsList []string) error
	GetWebtoolsSummary() (*model.WebToolsSummary, error)
}

// BBs exposes Building Block APIs for the driver adapters
type BBs interface {
	GetExample(orgID string, appID string, id string) (*model.Example, error)
	GetAppointmentUnits(providerid int, uin string, accesstoken string) (*[]model.AppointmentUnit, error)
	GetPeople(uin string, unitID int, providerid int, accesstoken string) (*[]model.AppointmentPerson, error)
	GetAppointmentOptions(uin string, unitid int, peopleid int, providerid int, startdate time.Time, enddate time.Time, accesstoken string) (*model.AppointmentOptions, error)
	CreateAppointment(appt *model.AppointmentPost, accessToken string) (*model.BuildingBlockAppointment, error)
	DeleteAppointment(uin string, providerid int, sourceid string, accesstoken string) (string, error)
	UpdateAppointment(appt *model.AppointmentPost, accessToken string) (*model.BuildingBlockAppointment, error)
	GetLegacyEvents() ([]model.LegacyEvent, error)
}

// TPS exposes third-party service APIs for the driver adapters
type TPS interface {
	GetExample(orgID string, appID string, id string) (*model.Example, error)
	CreateEvents(event []model.LegacyEventItem) ([]model.LegacyEventItem, error)
	DeleteEvents(ids []string, accountID string) error
}

// System exposes system administrative APIs for the driver adapters
type System interface {
	GetExample(orgID string, appID string, id string) (*model.Example, error)
}

// Shared exposes shared APIs for other interface implementations
type Shared interface {
	getExample(orgID string, appID string, id string) (*model.Example, error)
	getBuildingFeatures() ([]model.AppBuildingFeature, error)
}

// EventsBBAdapter is used by core to communicate with the events BB
type EventsBBAdapter interface {
	LoadAllLegacyEvents() ([]model.LegacyEvent, error)
}

// GeoAdapter is used by core to get geo services
type GeoAdapter interface {
	FindLocation(location string) (*model.LegacyLocation, error)
}

// ImageAdapter  is used to precess images
type ImageAdapter interface {
	ProcessImage(item model.WebToolsEvent) (*model.ContentImagesURL, error)
}

// Storage is used by core to storage data - DB storage adapter, file storage adapter etc
type Storage interface {
	RegisterStorageListener(listener storage.Listener)
	PerformTransaction(func(context storage.TransactionContext) error, int64) error

	FindGlobalConfig(context storage.TransactionContext, key string) (*model.GlobalConfigEntry, error)
	SaveGlobalConfig(context storage.TransactionContext, globalConfig model.GlobalConfigEntry) error

	FindConfig(configType string, appID string, orgID string) (*model.Config, error)
	FindConfigByID(id string) (*model.Config, error)
	FindConfigs(configType *string) ([]model.Config, error)
	InsertConfig(config model.Config) error
	UpdateConfig(config model.Config) error
	DeleteConfig(id string) error

	FindExample(orgID string, appID string, id string) (*model.Example, error)
	InsertExample(example model.Example) error
	UpdateExample(example model.Example) error
	DeleteExample(orgID string, appID string, id string) error

	FindCalendars(id string) (*[]model.UnitCalendar, error)

	InitializeLegacyLocations() error
	FindLegacyLocations() (model.LegacyLocationsListType, error)

	FindLegacyEventItems(context storage.TransactionContext) ([]model.LegacyEventItem, error)
	FindLegacyEventItemsBySourceID(context storage.TransactionContext, sourceID string) ([]model.LegacyEventItem, error)
	InsertLegacyEvents(context storage.TransactionContext, items []model.LegacyEventItem) ([]model.LegacyEventItem, error)
	DeleteLegacyEventsByIDs(context storage.TransactionContext, Ids map[string]string) error
	DeleteLegacyEventsBySourceID(context storage.TransactionContext, sourceID string) error
	DeleteLegacyEventsByIDsAndCreator(context storage.TransactionContext, ids []string, accountID string) error
	FindAllLegacyEvents() ([]model.LegacyEvent, error)
	FindAllWebtoolsCalendarIDs() ([]model.WebToolsCalendarID, error)
	FindWebtoolsLegacyEventByID(ids []string) ([]model.LegacyEventItem, error)

	FindWebtoolsBlacklistData(context storage.TransactionContext) ([]model.WebToolsItem, error)
	AddWebtoolsBlacklistData(dataSourceIDs []string, dataCalendarIDs []string, dataOriginatingCalendarIDs []string) error
	RemoveWebtoolsBlacklistData(dataSourceIDs []string, dataCalendarIDs []string, dataOriginatingCalendarIdsList []string) error
	FindWebtoolsOriginatingCalendarIDsBlacklistData() ([]model.WebToolsItem, error)

	FindImageItems() ([]model.ContentImagesURL, error)
	InsertImageItem(items model.ContentImagesURL) error

	FindLegacyLocationItems() ([]model.LegacyLocation, error)
	InsertLegacyLocationItem(items model.LegacyLocation) error

	LoadAppBuildingFeatures() ([]model.AppBuildingFeature, error)
}

// StorageListener represents storage listener
type StorageListener interface {
	OnConfigsUpdated()
	OnExamplesUpdated()
}

// Contact represents the adapter needed to pull campus specific contact information
type Contact interface {
	GetContactInformation(uin string, accessToken string, mode string, conf *model.EnvConfigData) (*model.Person, int, error)
}

// Courses represents the Courses adapter needed to pull campus specific course information
type Courses interface {
	GetStudentCourses(uin string, termid string, accessToken string, conf *model.EnvConfigData) (*[]model.Course, int, error)
	GetTermSessions() (*[4]model.TermSession, error)
	GetGiesCourses(uin string, accessToken string, conf *model.EnvConfigData) (*[]model.GiesCourse, int, error)
}

// LaundryService represents the adapter needed to interact with vendor specific laundry providers
type LaundryService interface {
	ListRooms(conf *model.EnvConfigData) (*model.Organization, error)
	GetLaundryRoom(roomid string, conf *model.EnvConfigData) (*model.RoomDetail, error)
	InitServiceRequest(machineID string, conf *model.EnvConfigData) (*model.MachineRequestDetail, error)
	SubmitServiceRequest(machineid string, problemCode string, comments string, firstName string, lastName string, phone string, email string, conf *model.EnvConfigData) (*model.ServiceRequestResult, error)
}

// WayFinding represents the adapter needed to interact with vendor specific building locations
type WayFinding interface {
	GetEntrance(bldgID string, adaAccessibleOnly bool, latitude float64, longitude float64, conf *model.EnvConfigData) (*model.Entrance, error)
	GetBuildings(conf *model.EnvConfigData) (*[]model.Building, error)
	GetBuilding(bldgID string, adaAccessibleOnly bool, latitude float64, longitude float64, conf *model.EnvConfigData) (*model.Building, error)
	GetFloorPlan(bldgNum string, floornumber string, markers string, highlites string, conf *model.EnvConfigData) (*model.FloorPlan, error)
}

// Appointments represents the adapter needed to interace with various appoinment data providers
type Appointments interface {
	GetUnits(uin string, accesstoken string, providerid int, conf *model.EnvConfigData) (*[]model.AppointmentUnit, error)
	GetPeople(uin string, unitID int, providerid int, accesstoken string, conf *model.EnvConfigData) (*[]model.AppointmentPerson, error)
	GetTimeSlots(uin string, unitid int, advisorid int, providerid int, startdate time.Time, enddate time.Time, accesstoken string, conf *model.EnvConfigData) (*model.AppointmentOptions, error)
	CreateAppointment(appt *model.AppointmentPost, accesstoken string, conf *model.EnvConfigData) (*model.BuildingBlockAppointment, error)
	DeleteAppointment(uin string, sourceid string, accesstoken string, conf *model.EnvConfigData) (string, error)
	UpdateAppointment(appt *model.AppointmentPost, accesstoken string, conf *model.EnvConfigData) (*model.BuildingBlockAppointment, error)
}

// SuccessTeam represents the adapter needed to interface with the various assignedstaff end points to create the user's success team
type SuccessTeam interface {
	GetSuccessTeam(uin string, calendars *[]model.UnitCalendar, accesstoken string, conf *model.EnvConfigData) (*model.SuccessTeam, int, error)
	GetPrimaryCareProvider(uin string, accesstoken string, conf *model.EnvConfigData) (*[]model.SuccessTeamMember, int, error)
	GetAcademicAdvisors(uin string, calendars *[]model.UnitCalendar, accesstoken string, conf *model.EnvConfigData) (*[]model.SuccessTeamMember, int, error)
}
