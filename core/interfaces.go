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
	model "apigateway/core/model"
)

// Services exposes APIs for the driver adapters
type Services interface {
	GetVersion() string
	StoreRecord(name string) error
	ListLaundryRooms() (*model.Organization, error)
	GetLaundryRoom(roomid string) (*model.RoomDetail, error)
	InitServiceRequest(machineid string) (*model.MachineRequestDetail, error)
	SubmitServiceRequest(machineID string, problemCode string, comments string, firstname string, lastname string, phone string, email string) (*model.ServiceRequestResult, error)
	GetBuilding(bldgID string, adaOnly bool) (*model.Building, error)
	GetEntrance(bldgID string, adaOnly bool, latitude float64, longitude float64) (*model.Entrance, error)
	GetBuildings() (*[]model.Building, error)
	GetContactInfo(uin string, accessToken string, mode string) (*model.Person, int, error)
}

type servicesImpl struct {
	app *Application
}

func (s *servicesImpl) GetVersion() string {
	return s.app.getVersion()
}

func (s *servicesImpl) StoreRecord(name string) error {
	return s.app.storeRecord(name)
}

func (s *servicesImpl) ListLaundryRooms() (*model.Organization, error) {
	lr, err := s.app.listLaundryRooms()
	return &lr, err
}

func (s *servicesImpl) GetLaundryRoom(roomid string) (*model.RoomDetail, error) {
	ap, err := s.app.listAppliances(roomid)
	return &ap, err
}

func (s *servicesImpl) InitServiceRequest(machineid string) (*model.MachineRequestDetail, error) {
	sr, err := s.app.initServiceRequest(machineid)
	return &sr, err
}

func (s *servicesImpl) SubmitServiceRequest(machineID string, problemCode string, comments string, firstname string, lastname string, phone string, email string) (*model.ServiceRequestResult, error) {
	srr, err := s.app.submitServiceRequest(machineID, problemCode, comments, firstname, lastname, phone, email)
	return &srr, err
}

func (s *servicesImpl) GetBuilding(bldgID string, adaOnly bool) (*model.Building, error) {
	bldg, err := s.app.getBuilding(bldgID, adaOnly)
	return &bldg, err
}

func (s *servicesImpl) GetEntrance(bldgID string, adaOnly bool, latitude float64, longitude float64) (*model.Entrance, error) {
	entrance, err := s.app.getEntrance(bldgID, adaOnly, latitude, longitude)
	return entrance, err
}

func (s *servicesImpl) GetBuildings() (*[]model.Building, error) {
	buildings, err := s.app.getBuildings()
	return buildings, err
}

func (s *servicesImpl) GetContactInfo(uin string, accessToken string, mode string) (*model.Person, int, error) {
	person, statusCode, err := s.app.getContactInfo(uin, accessToken, mode)
	return person, statusCode, err
}

// Storage is used by core to storage data - DB storage adapter, file storage adapter etc
type Storage interface {
	StoreRecord(name string) error
}

//Laundry is used by core to request data from the laundry provider
type Laundry interface {
	ListRooms() (*model.Organization, error)
	GetLaundryRoom(roomid string) (*model.RoomDetail, error)
	InitServiceRequest(machineID string) (*model.MachineRequestDetail, error)
	SubmitServiceRequest(machineID string, problemCode string, comments string, firstname string, lastname string, phone string, email string) (*model.ServiceRequestResult, error)
}

//BuildingLocation is used by core to request data from the building location/entrance provider
type BuildingLocation interface {
	GetBuilding(bldgID string, adaAccessibleOnly bool) (*model.Building, error)
	GetEntrance(bldgID string, adaAccessibleOnly bool, latitude float64, longitude float64) (*model.Entrance, error)
	GetBuildings() (*[]model.Building, error)
}

//ContactInformation is used by core to request data from the contact information provider
type ContactInformation interface {
	GetContactInformation(uin string, accessToken string, mode string) (*model.Person, int, error)
}
