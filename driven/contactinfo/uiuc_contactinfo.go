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
func (lv *ContactAdapter) GetContactInformation(uin string, accessToken string) (*model.Person, error) {
	finalURL := lv.APIEndpoint + "/" + uin
	campusData, err := lv.getData(finalURL, accessToken)
	if err != nil {
		return nil, err
	}

	retValue, err := newPerson(&campusData.People[0])
	if err != nil {
		return nil, err
	}
	return retValue, nil
}

func (lv *ContactAdapter) getData(targetURL string, accessToken string) (*campusUserData, error) {
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, targetURL, nil)

	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", "Bearer "+accessToken)
	req.Header.Set("Ocp-Apim-Subscription-Key", lv.APIKey)
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode == 401 {
		fmt.Println(string(body))
		return nil, errors.New(string(body))
	}

	if res.StatusCode == 400 {
		fmt.Println(string(body))
		return nil, errors.New("Bad request to api end point")
	}

	data := campusUserData{}
	err = json.Unmarshal(body, &data)

	if err != nil {
		return nil, err
	}
	return &data, nil
}