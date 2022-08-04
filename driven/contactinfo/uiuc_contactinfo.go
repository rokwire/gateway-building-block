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

package contactinfo

import (
	model "apigateway/core/model"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

type simpleType struct {
	Code        string `json:"code"`
	Description string `json:"description"`
}

type campusUserData struct {
	Object  string         `json:"object"`
	Version string         `json:"version"`
	People  []campusPerson `json:"list"`
}

type name struct {
	Pidm      int    `json:"pidm"`
	Uin       string `json:"uin"`
	LastName  string `json:"lastName"`
	FirstName string `json:"firstName"`
	NameType  string `json:"type"`
}

type address struct {
	GUID            int        `json:"guid"`
	Pidm            int        `json:"pidm"`
	FromDate        string     `json:"fromDate"`
	ActivityDate    string     `json:"activityDate"`
	Type            simpleType `json:"type"`
	SequenceNum     int        `json:"sequenceNum"`
	StreetLine1     string     `json:"streetLine1"`
	City            string     `json:"city"`
	State           simpleType `json:"state"`
	ZipCode         string     `json:"zipCode"`
	County          simpleType `json:"county"`
	EffectiveStatus string     `json:"effectiveStatus"`
}

type phone struct {
	GUID                  int        `json:"guid"`
	Pidm                  int        `json:"pidm"`
	SequenceNum           int        `json:"sequenceNum"`
	Type                  simpleType `json:"type"`
	ActivityDate          string     `json:"activityDate"`
	LinkedAddressType     simpleType `json:"linkedAddressType"`
	LinkedAddressSequence int        `json:"linkedAddressSequence"`
	AreaCode              string     `json:"areaCode"`
	PhoneNumber           string     `json:"phoneNumber"`
	PrimaryInd            string     `json:"primaryInd"`
	EffectiveStatus       string     `json:"effectiveStatus"`
}

type emergencyContactName struct {
	LastName  string `json:"lastName"`
	FirstName string `json:"firstName"`
}

type emergencyPhone struct {
	PhoneArea   string `json:"areaCode"`
	PhoneNumber string `json:"phoneNumber"`
}

type emergencyAddress struct {
	Type    simpleType `json:"type"`
	Street1 string     `json:"streetLine1"`
	City    string     `json:"city"`
	State   simpleType `json:"state"`
	ZipCode string     `json:"zipCode"`
}

type emergencyContact struct {
	GUID         int                  `json:"guid"`
	Pidm         int                  `json:"pidm"`
	Priority     string               `json:"priority"`
	Relationship simpleType           `json:"relationship"`
	Name         emergencyContactName `json:"name"`
	Phone        emergencyPhone       `json:"phone"`
	Address      emergencyAddress     `json:"address"`
}

type campusPerson struct {
	Names             []name             `json:"name"`
	Addresses         []address          `json:"address"`
	Phone             []phone            `json:"phone"`
	EmergencyContacts []emergencyContact `json:"emergencyContact"`
}

//ContactAdapter is a vendor specific structure that implements the contanct information interface
type ContactAdapter struct {
	APIKey      string
	APIEndpoint string
}

//NewContactAdapter returns a vendor specific implementation of the contanct information interface
func NewContactAdapter(apikey string, url string) *ContactAdapter {
	return &ContactAdapter{APIKey: apikey, APIEndpoint: url}

}

func newPerson(cr *campusPerson) (*model.Person, error) {
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

//GetContactInformation returns a contact information object for a student
func (lv *ContactAdapter) GetContactInformation(uin string, accessToken string, mode string) (*model.Person, int, error) {

	finalURL := lv.APIEndpoint + "/person/contact-summary-query/" + uin

	if mode != "0" {
		finalURL = lv.APIEndpoint + "/mock/123456789"
	}

	campusData, statusCode, err := lv.getData(finalURL, accessToken)
	if err != nil {
		return nil, statusCode, err
	}

	if len(campusData.People) == 0 {
		return nil, 404, errors.New("No contact data found")
	}

	retValue, err := newPerson(&campusData.People[0])
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	return retValue, statusCode, nil
}

func (lv *ContactAdapter) getData(targetURL string, accessToken string) (*campusUserData, int, error) {
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, targetURL, nil)

	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	req.Header.Add("Authorization", "Bearer "+accessToken)
	req.Header.Set("Ocp-Apim-Subscription-Key", lv.APIKey)
	res, err := client.Do(req)
	if err != nil {
		return nil, res.StatusCode, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, res.StatusCode, err
	}

	if res.StatusCode == 401 {
		return nil, res.StatusCode, errors.New(res.Status)
	}

	if res.StatusCode == 403 {
		return nil, res.StatusCode, errors.New(res.Status)
	}

	if res.StatusCode == 400 {
		return nil, res.StatusCode, errors.New("Bad request to api end point")
	}

	if res.StatusCode == 406 {
		return nil, res.StatusCode, errors.New("Server returned 406: possible uin claim mismatch")
	}

	//campus api returns a 502 when there is no banner contact data for the uin
	if res.StatusCode == 502 {
		return nil, 404, errors.New(res.Status)
	}
	if res.StatusCode == 200 {
		data := campusUserData{}
		err = json.Unmarshal(body, &data)

		if err != nil {
			return nil, res.StatusCode, err
		}
		return &data, res.StatusCode, nil
	}

	return nil, res.StatusCode, errors.New("Error making request: " + fmt.Sprint(res.StatusCode) + ": " + string(body))

}
