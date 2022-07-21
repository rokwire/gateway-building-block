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

//Person represents the basic structure returned to the caller when contanct information is requested
type Person struct {
	UIN               string             `json:"uin"`
	FirstName         string             `json:"firstName"`
	LastName          string             `json:"lastName"`
	PreferredName     string             `json:"preferred"`
	MailingAddress    Address            `json:"mailingAddress"`
	PermAddress       Address            `json:"permanentAddress"`
	EmergencyContacts []EmergencyContact `json:"emergencycontacts"`
}

//AddressType is used as an enumeration for address types
type AddressType string

//PhoneType is an enumeration representing phone number types
type PhoneType string

//constants for address types
const (
	Mailing   AddressType = "MA"
	Permanent AddressType = "PR"
)

//constants for phone types
const (
	MailingAddressPhone PhoneType = "MA"
	PermAddressPhone    PhoneType = "PR"
	CellPhone           PhoneType = "CELL"
	ECPhone             PhoneType = "EC"
)

//CodeDescType is a generic struct representing simple code/value objects
type CodeDescType struct {
	Code string
	Name string
}

//Address represents an address returned as part of a Person object
type Address struct {
	Type      AddressType
	Street1   string
	City      string
	StateAbbr string
	StateName string
	ZipCode   string
	County    string
	Phone     PhoneNumber
}

//PhoneNumber represents the parts of a phone number returned as part of a person object
type PhoneNumber struct {
	AreaCode string
	Number   string
}

//EmergencyContact represents the data needed to display emergency contact information for a person
type EmergencyContact struct {
	Priority     string
	RelationShip CodeDescType
	FirstName    string
	LastName     string
	Address      Address
}
