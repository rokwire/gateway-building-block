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

package uiuc

import (
	model "application/core/model"
	"encoding/xml"
	"regexp"
	"strconv"
)

// Appliance represents a CCS definition of a washer/dryer
type Appliance struct {
	XMLName       xml.Name `xml:"appliance"`
	ApplianceKey  string   `xml:"appliance_desc_key"`
	LrmStatus     string   `xml:"lrm_status"`
	ApplianceType string   `xml:"appliance_type"`
	Status        string   `xml:"status"`
	OutOfService  string   `xml:"out_of_service"`
	Label         string   `xml:"label"`
	AvgCycleTime  string   `xml:"avg_cycle_time"`
	TimeRemaining string   `xml:"time_remaining"`
}

// Laundryroom represents the csc definition of a laundry room
type Laundryroom struct {
	XMLName    xml.Name     `xml:"laundry_room"`
	Name       string       `xml:"laundry_room_name"`
	CampusName string       `xml:"campus_name"`
	Appliances []*Appliance `xml:"appliances>appliance"`
}

// Laundrylocation represents the CSC definition of a laundry location
type Laundrylocation struct {
	Location        int      `xml:"location"`
	XMLName         xml.Name `xml:"laundryroom"`
	Campusname      string   `xml:"campus_name"`
	Laundryroomname string   `xml:"laundry_room_name"`
	Status          string   `xml:"status"`
}

// School represents the csc definition of a customer
type School struct {
	XMLName      xml.Name           `xml:"school"`
	SchoolName   string             `xml:"school_name"`
	LaundryRooms []*Laundrylocation `xml:"laundry_rooms>laundryroom"`
}

// Capacity represents the available washers and dryers in a room
type Capacity struct {
	XMLName    xml.Name `xml:"laundry_room"`
	NumWashers string   `xml:"washer"`
	NumDryers  string   `xml:"dryer"`
}

// Machinedetail represents the CSC machine details needed to submit a service ticket
type Machinedetail struct {
	Address             string `json:"address"`
	LaundryLocation     string `json:"laundryLocaiton"`
	MachineID           string `json:"machineId"`
	MachineType         string `json:"machineType"`
	Message             string `json:"message"`
	Property            string `json:"property"`
	RecentServiceDate   string `json:"recentServiceDate"`
	RecentServiceNotes  string `json:"recentServiceNotes"`
	RecentServiceStatus string `json:"recentServiceStatus"`
	SiteID              string `json:"siteID"`
}

// NewMachineRequestDetail creates an app formatted machinerequestdetail ojbect from campus data
func NewMachineRequestDetail(machineid string, message string, serviceStatus string, machinetype string) *model.MachineRequestDetail {
	var openTicket = serviceStatus == "Open"
	mrd := model.MachineRequestDetail{MachineID: machineid, Message: message, OpenIssue: openTicket, MachineType: machinetype}
	return &mrd
}

// NewLaundryRoom returns an app formatted laundry room object from campus data
func NewLaundryRoom(id int, name string, status string, location *model.LaundryDetails) *model.LaundryRoom {
	lr := model.LaundryRoom{Name: name, ID: id, Status: status, Location: location}
	return &lr
}

// NewAppliance returns an app formatted appliance ojbect from campus data
func NewAppliance(id string, appliancetype string, cycletime int, status string, timeremaining string, label string) *model.Appliance {

	var finalStatus string
	switch status {
	case "Available":
		finalStatus = "available"
	case "In Use":
		finalStatus = "in_use"
	default:
		finalStatus = "out_of_service"
	}

	if finalStatus == "available" || finalStatus == "out_of_service" {
		appl := model.Appliance{ID: id, ApplianceType: appliancetype, AverageCycleTime: cycletime, Status: finalStatus, Label: label}
		return &appl
	}

	re := regexp.MustCompile("[0-9]+")
	intsInString := re.FindAllString(timeremaining, 1)

	if intsInString != nil {
		intConvValue, err := strconv.ParseInt(intsInString[0], 10, 32)
		if err != nil {
			appl := model.Appliance{ID: id, ApplianceType: appliancetype, AverageCycleTime: cycletime, Status: finalStatus, Label: label}
			return &appl
		}

		trValue := int(intConvValue)
		appl := model.Appliance{ID: id, ApplianceType: appliancetype, AverageCycleTime: cycletime, Status: finalStatus, TimeRemaining: &trValue, Label: label}
		return &appl
	}

	appl := model.Appliance{ID: id, ApplianceType: appliancetype, AverageCycleTime: cycletime, Status: finalStatus, Label: label}
	return &appl

}
