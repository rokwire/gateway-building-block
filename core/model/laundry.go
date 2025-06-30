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
	"github.com/rokwire/rokwire-building-block-sdk-go/utils/logging/logutils"
)

const (
	//TypeLaundryServiceSubmission type
	TypeLaundryServiceSubmission logutils.MessageDataType = "laundryservicerequest"
)

const (
	//TypeLaundryRooms type
	TypeLaundryRooms logutils.MessageDataType = "laundryrooms"
)

// LaundryRoom represents the basic information returned as part of requesting and organization
type LaundryRoom struct {
	ID       int
	Name     string
	Status   string
	Location *LaundryDetails
}

// Organization represents the top most level of information provided by the laundry api
type Organization struct {
	SchoolName   string
	LaundryRooms []*LaundryRoom
}

// RoomDetail represents details about a specific laundry room, including a list of appliances
type RoomDetail struct {
	NumWashers int
	NumDryers  int
	Appliances []*Appliance
	RoomName   string
	CampusName string
	Location   *LaundryDetails
}

// Appliance represents the information specific to an identifiable appliance in a laundry room
type Appliance struct {
	ID               string
	Status           string
	ApplianceType    string
	AverageCycleTime int
	TimeRemaining    *int
	Label            string
}

// MachineRequestDetail represents the basic details needed in order to submit a request about a machine
type MachineRequestDetail struct {
	MachineID    string
	Message      string
	OpenIssue    bool
	ProblemCodes []string
	MachineType  string
}

// ServiceRequestResult represents the information returned upon submission of a machine service request
type ServiceRequestResult struct {
	Message       string
	RequestNumber string
	Status        string
}

// ServiceSubmission represents the data required to submit a service request for a laundry machine
type ServiceSubmission struct {
	MachineID   *string `json:"machineid" bson:"machineid"`
	ProblemCode *string `json:"problemcode" bson:"problemcode"`
	Comments    *string `json:"comments" bson:"comments"`
	FirstName   *string `json:"firstname" bson:"firstname"`
	LastName    *string `json:"lastname" bson:"lastname"`
	Phone       *string `json:"phone" bson:"phone"`
	Email       *string `json:"email" bson:"email"`
}
