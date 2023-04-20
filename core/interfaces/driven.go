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

package interfaces

import (
	"application/core/model"
	"time"
)

// Storage is used by core to storage data - DB storage adapter, file storage adapter etc
type Storage interface {
	RegisterStorageListener(listener StorageListener)
	PerformTransaction(func(storage Storage) error) error

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
}

// Appointments represents the adapter needed to interace with various appoinment data providers
type Appointments interface {
	GetUnits(uin string, accesstoken string, providerid int, conf *model.EnvConfigData) (*[]model.AppointmentUnit, error)
	GetPeople(uin string, unitID int, providerid int, accesstoken string, conf *model.EnvConfigData) (*[]model.AppointmentPerson, error)
	GetTimeSlots(uin string, unitid int, advisorid int, providerid int, startdate time.Time, enddate time.Time, accesstoken string, conf *model.EnvConfigData) (*model.AppointmentOptions, error)
	CreateAppointment(appt *model.AppointmentPost, accesstoken string, conf *model.EnvConfigData) (string, error)
	//DeleteAppointment(uin string, accesstoken string) (string, error)
}
