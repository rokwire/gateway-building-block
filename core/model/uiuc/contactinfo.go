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
)

// SimpleType contains a common data structure used in some properties of the campus data
type SimpleType struct {
	Code        string `json:"code"`
	Description string `json:"description"`
}

// CampusUserData represents the full data coming back from campus
type CampusUserData struct {
	Object  string         `json:"object"`
	Version string         `json:"version"`
	People  []CampusPerson `json:"list"`
}

// Name represents the fields campus uses to represent a person's name
type Name struct {
	Pidm      int    `json:"pidm"`
	Uin       string `json:"uin"`
	LastName  string `json:"lastName"`
	FirstName string `json:"firstName"`
	NameType  string `json:"type"`
}

// Address represents the campus definition of a person's address
type Address struct {
	GUID            int        `json:"guid"`
	Pidm            int        `json:"pidm"`
	FromDate        string     `json:"fromDate"`
	ActivityDate    string     `json:"activityDate"`
	Type            SimpleType `json:"type"`
	SequenceNum     int        `json:"sequenceNum"`
	StreetLine1     string     `json:"streetLine1"`
	City            string     `json:"city"`
	State           SimpleType `json:"state"`
	ZipCode         string     `json:"zipCode"`
	County          SimpleType `json:"county"`
	EffectiveStatus string     `json:"effectiveStatus"`
}

// Phone represents the canpus definitiono of a person's phone number
type Phone struct {
	GUID                  int        `json:"guid"`
	Pidm                  int        `json:"pidm"`
	SequenceNum           int        `json:"sequenceNum"`
	Type                  SimpleType `json:"type"`
	ActivityDate          string     `json:"activityDate"`
	LinkedAddressType     SimpleType `json:"linkedAddressType"`
	LinkedAddressSequence int        `json:"linkedAddressSequence"`
	AreaCode              string     `json:"areaCode"`
	PhoneNumber           string     `json:"phoneNumber"`
	PrimaryInd            string     `json:"primaryInd"`
	EffectiveStatus       string     `json:"effectiveStatus"`
}

// EmergencyContactName represents the campus definition of an EmergencyContact name
type EmergencyContactName struct {
	LastName  string `json:"lastName"`
	FirstName string `json:"firstName"`
}

// EmergencyPhone represents the campus definition of an emergency phone number
type EmergencyPhone struct {
	PhoneArea   string `json:"areaCode"`
	PhoneNumber string `json:"phoneNumber"`
}

// EmergencyAddress represents the campus definition of an emergency contact address
type EmergencyAddress struct {
	Type    SimpleType `json:"type"`
	Street1 string     `json:"streetLine1"`
	City    string     `json:"city"`
	State   SimpleType `json:"state"`
	ZipCode string     `json:"zipCode"`
}

// EmergencyContact represetnts the campus definition of a person's emergency contact person
type EmergencyContact struct {
	GUID         int                  `json:"guid"`
	Pidm         int                  `json:"pidm"`
	Priority     string               `json:"priority"`
	Relationship SimpleType           `json:"relationship"`
	Name         EmergencyContactName `json:"name"`
	Phone        EmergencyPhone       `json:"phone"`
	Address      EmergencyAddress     `json:"address"`
}

// CampusPerson represents the campus definitioin of a person's contact information
type CampusPerson struct {
	Names             []Name             `json:"name"`
	Addresses         []Address          `json:"address"`
	Phone             []Phone            `json:"phone"`
	EmergencyContacts []EmergencyContact `json:"emergencyContact"`
}

// NewPerson constructs an app formatted person object from the campus representation
func NewPerson(cr *CampusPerson) (*model.Person, error) {
	ret := model.Person{}

	for i := 0; i < len(cr.Names); i++ {
		crntName := cr.Names[i]
		if crntName.NameType == "LEGAL" {
			ret.FirstName = crntName.FirstName
			ret.LastName = crntName.LastName
			ret.UIN = crntName.Uin
		} else {
			ret.PreferredName = crntName.FirstName
		}
	}

	for i := 0; i < len(cr.Addresses); i++ {
		ca := cr.Addresses[i]
		switch ca.Type.Code {
		case "MA":
			ret.MailingAddress = model.Address{Street1: ca.StreetLine1, City: ca.City,
				StateAbbr: ca.State.Code, StateName: ca.State.Description, ZipCode: ca.ZipCode,
				County: ca.County.Description, Type: "MA"}
		case "PR":
			ret.PermAddress = model.Address{Street1: ca.StreetLine1, City: ca.City,
				StateAbbr: ca.State.Code, StateName: ca.State.Description, ZipCode: ca.ZipCode,
				County: ca.County.Description, Type: "PR"}
		default:
		}
	}

	for i := 0; i < len(cr.Phone); i++ {
		cp := cr.Phone[i]
		switch cp.Type.Code {
		case "MA":
			ret.MailingAddress.Phone = model.PhoneNumber{AreaCode: cp.AreaCode, Number: cp.PhoneNumber}
		case "PR":
			ret.PermAddress.Phone = model.PhoneNumber{AreaCode: cp.AreaCode, Number: cp.PhoneNumber}
		default:
		}
	}

	for i := 0; i < len(cr.EmergencyContacts); i++ {
		ec := cr.EmergencyContacts[i]
		newEC := model.EmergencyContact{FirstName: ec.Name.FirstName, LastName: ec.Name.LastName,
			Priority:     ec.Priority,
			RelationShip: model.CodeDescType{Code: ec.Relationship.Code, Name: ec.Relationship.Description},
			Address: model.Address{Street1: ec.Address.Street1, City: ec.Address.City, StateAbbr: ec.Address.State.Code,
				StateName: ec.Address.State.Description, ZipCode: ec.Address.ZipCode, Type: "ECA",
				Phone: model.PhoneNumber{AreaCode: ec.Phone.PhoneArea, Number: ec.Phone.PhoneNumber}}}
		ret.EmergencyContacts = append(ret.EmergencyContacts, newEC)

	}

	return &ret, nil
}
