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

	"github.com/rokwire/core-auth-library-go/v3/tokenauth"
)

// Default exposes client APIs for the driver adapters
type Default interface {
	GetVersion() string
}

// Client exposes client APIs for the driver adapters
type Client interface {
	GetExample(orgID string, appID string, id string) (*model.Example, error)
	GetUnitCalendars(orgID string, appID string, id string) (*[]model.UnitCalendar, error)
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
	GetSuccessTeam(uin string, accessToken string) (*[]model.SuccessTeamMember, int, error)
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
}

// TPS exposes third-party service APIs for the driver adapters
type TPS interface {
	GetExample(orgID string, appID string, id string) (*model.Example, error)
}

// System exposes system administrative APIs for the driver adapters
type System interface {
	GetExample(orgID string, appID string, id string) (*model.Example, error)
}
