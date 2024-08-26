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

package uiucadapters

import (
	model "application/core/model"
	uiuc "application/core/model/uiuc"
	"encoding/json"
	"encoding/xml"
	"errors"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
)

// CSCLaundryView is a vendor specific structure that implements the Laundry interface
type CSCLaundryView struct {
	serviceCookie string
	laundryAssets map[string]model.LaundryDetails
}

// NewCSCLaundryAdapter returns a vendor specific implementation of the Laundry interface
func NewCSCLaundryAdapter(assets map[string]model.LaundryDetails) *CSCLaundryView {
	return &CSCLaundryView{laundryAssets: assets}

}

// ListRooms lists the laundry rooms
func (lv *CSCLaundryView) ListRooms(conf *model.EnvConfigData) (*model.Organization, error) {

	laundryAPI := conf.LaundryViewURL
	laundryKey := conf.LaundryViewKey
	url := laundryAPI + "/school/?api_key=" + laundryKey + "&method=getRoomData&type=json"
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.Status == "200 OK" {
		body, bodyerr := io.ReadAll(resp.Body)
		if bodyerr != nil {
			return nil, err
		}

		var nS uiuc.School
		out := []byte(body)
		if err := xml.Unmarshal(out, &nS); err != nil {
			log.Fatal("could not unmarshal xml data")
			return nil, err
		}
		org := model.Organization{SchoolName: nS.SchoolName}
		org.LaundryRooms = make([]*model.LaundryRoom, 0)

		for _, lr := range nS.LaundryRooms {
			if len(lv.laundryAssets) > 0 {
				org.LaundryRooms = append(org.LaundryRooms, uiuc.NewLaundryRoom(lr.Location, lr.Laundryroomname, lr.Status, lv.getLocationData(strconv.Itoa(lr.Location))))
			} else {
				org.LaundryRooms = append(org.LaundryRooms, uiuc.NewLaundryRoom(lr.Location, lr.Laundryroomname, lr.Status, lv.getLocationData("0")))
			}
		}
		return &org, nil
	}
	return nil, err
}

// GetLaundryRoom returns the room details along with the list of machines in that room
func (lv *CSCLaundryView) GetLaundryRoom(roomid string, conf *model.EnvConfigData) (*model.RoomDetail, error) {

	laundryAPI := conf.LaundryViewURL
	laundryKey := conf.LaundryViewKey
	url := laundryAPI + "/room/?api_key=" + laundryKey + "&method=getAppliances&location=" + roomid + "&type=json"
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.Status == "200 OK" {
		body, bodyerr := io.ReadAll(resp.Body)
		if bodyerr != nil {
			return nil, err
		}

		var lr uiuc.Laundryroom
		out := []byte(body)
		if err := xml.Unmarshal(out, &lr); err != nil {
			log.Fatal("could not unmarshal xml data")
			return nil, err
		}

		rd := model.RoomDetail{CampusName: lr.CampusName, RoomName: lr.Name}
		rd.Appliances = make([]*model.Appliance, len(lr.Appliances))
		roomCapacity, _ := lv.getNumAvailable(laundryAPI, laundryKey, roomid)
		rd.NumDryers = evalNumAvailable(roomCapacity.NumDryers)
		rd.NumWashers = evalNumAvailable(roomCapacity.NumWashers)

		for i, appl := range lr.Appliances {
			avgCycle, _ := strconv.Atoi(appl.AvgCycleTime)
			rd.Appliances[i] = uiuc.NewAppliance(appl.ApplianceKey, appl.ApplianceType, avgCycle, appl.Status, appl.TimeRemaining, appl.Label, appl.OutOfService)
		}

		if len(lv.laundryAssets) > 0 {
			rd.Location = lv.getLocationData(roomid)
		}
		return &rd, nil
	}
	return nil, err
}

func (lv *CSCLaundryView) getLocationData(roomid string) *model.LaundryDetails {
	if asset, ok := lv.laundryAssets[roomid]; ok {
		return &asset
	}
	return &model.LaundryDetails{Latitude: 0, Longitude: 0, Floor: 0}
}

func (lv *CSCLaundryView) getNumAvailable(apiURL string, apikey string, roomid string) (*uiuc.Capacity, error) {

	url := apiURL + "/room/?api_key=" + apikey + "&method=getNumAvailable&location=" + roomid + "&type=json"
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.Status == "200 OK" {
		body, bodyerr := io.ReadAll(resp.Body)
		if bodyerr != nil {
			return nil, err
		}

		var cap uiuc.Capacity
		out := []byte(body)
		if err := xml.Unmarshal(out, &cap); err != nil {
			log.Fatal("could not unmarshal xml data")
			return nil, err
		}

		return &cap, nil
	}
	return nil, err
}

// InitServiceRequest gets machine request details needed to initialize a laundry service request
func (lv *CSCLaundryView) InitServiceRequest(machineID string, conf *model.EnvConfigData) (*model.MachineRequestDetail, error) {

	serviceURL := conf.LaundyrServiceURL
	serviceKey := conf.LaundryServiceKey
	serviceAuthToken := conf.LaundryServiceBasicAuth

	authTokens, err := lv.getAuthTokens(serviceURL, serviceKey, serviceAuthToken)
	if err != nil {
		return nil, err
	}

	subscriptionkey := authTokens["SUBSCRIPTIONKEY"]
	serviceToken := authTokens["SERVICETOKEN"]

	md, err := lv.getMachineDetails(serviceURL, subscriptionkey, serviceAuthToken, serviceToken, machineID)
	if err != nil {
		return nil, err
	}

	mrd := uiuc.NewMachineRequestDetail(md.MachineID, md.Message, md.RecentServiceStatus, md.MachineType)
	mrd.ProblemCodes, err = lv.getProblemCodes(serviceURL, subscriptionkey, serviceToken, serviceAuthToken, md.MachineType)
	if err != nil {
		return nil, err
	}
	return mrd, nil
}

// SubmitServiceRequest submits a request for a machine
func (lv *CSCLaundryView) SubmitServiceRequest(machineid string, problemCode string, comments string, firstName string, lastName string, phone string, email string, conf *model.EnvConfigData) (*model.ServiceRequestResult, error) {

	serviceURL := conf.LaundyrServiceURL
	serviceKey := conf.LaundryServiceKey
	serviceAuthToken := conf.LaundryServiceBasicAuth

	authTokens, err := lv.getAuthTokens(serviceURL, serviceKey, serviceAuthToken)
	if err != nil {
		return nil, err
	}

	subscriptionkey := authTokens["SUBSCRIPTIONKEY"]
	serviceToken := authTokens["SERVICETOKEN"]

	srr, err := lv.submitTicket(serviceURL, subscriptionkey, serviceToken, serviceAuthToken, machineid, problemCode, comments, firstName, lastName, phone, email)
	if err != nil {
		return nil, err
	}
	return srr, nil
}

func (lv *CSCLaundryView) getAuthTokens(serviceURL string, serviceKey string, serviceAuthToken string) (map[string]string, error) {
	subscriptionkey, err := lv.getServiceSubscriptionKey(serviceURL, serviceKey, serviceAuthToken)
	if err != nil {
		return nil, err
	}

	serviceToken, err := lv.getServiceToken(serviceURL, serviceAuthToken, subscriptionkey)
	if err != nil {
		return nil, err
	}

	tokens := make(map[string]string)
	tokens["SUBSCRIPTIONKEY"] = subscriptionkey
	tokens["SERVICETOKEN"] = serviceToken
	return tokens, nil

}

func (lv *CSCLaundryView) getServiceSubscriptionKey(serviceurl string, serviceKey string, authToken string) (string, error) {
	url := serviceurl + "/getSubscriptionKey"
	method := "POST"

	payload := `{"subscription-id": "uiuc", "key-type": "primaryKey" }`

	headers := make(map[string]string)
	headers["Ocp-Apim-Subscription-Key"] = serviceKey
	headers["Content-Type"] = "application/json"
	headers["Authorization"] = "Basic " + authToken

	body, err := lv.makeLaundryServiceWebRequest(url, method, headers, payload)
	if err != nil {
		return "", err
	}

	var dat map[string]interface{}
	if err := json.Unmarshal(body, &dat); err != nil {
		return "", err
	}

	if _, keyExists := dat["subscription-key"]; !keyExists {
		return "", errors.New("subscription key not returned")
	}

	return dat["subscription-key"].(string), nil
}

func (lv *CSCLaundryView) getServiceToken(serviceurl string, authToken string, subscriptionKey string) (string, error) {
	url := serviceurl + "/generateToken?subscription-key=" + subscriptionKey
	method := "GET"

	headers := make(map[string]string)
	headers["Content-Type"] = "application/json"
	headers["Authorization"] = "Basic " + authToken

	body, err := lv.makeLaundryServiceWebRequest(url, method, headers, "")

	if err != nil {
		return "", err
	}

	var dat map[string]interface{}
	if err := json.Unmarshal(body, &dat); err != nil {
		return "", err
	}

	if _, keyExists := dat["token"]; !keyExists {
		return "", errors.New("token not returned")
	}
	return dat["token"].(string), nil
}

func (lv *CSCLaundryView) getMachineDetails(serviceurl string, subscriptionkey string, authtoken string, servicetoken string, machineid string) (*uiuc.Machinedetail, error) {
	md := uiuc.Machinedetail{}

	url := serviceurl + "/machineDetails?subscription-key=" + subscriptionkey
	method := "POST"

	payload := `{"machineId":"` + machineid + `"}`

	headers := make(map[string]string)
	headers["X-CSRFToken"] = servicetoken
	headers["Cookie"] = "session=" + lv.serviceCookie
	headers["Content-Type"] = "application/json"
	headers["Authorization"] = "Basic " + authtoken

	body, err := lv.makeLaundryServiceWebRequest(url, method, headers, payload)

	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(body, &md); err != nil {
		return nil, err
	}
	return &md, nil
}

func (lv *CSCLaundryView) getProblemCodes(serviceurl string, subscriptionkey string, servicetoken string, authtoken string, machinetype string) ([]string, error) {
	url := serviceurl + "/problemCodes?subscription-key=" + subscriptionkey
	method := "POST"

	payload := `{"machineType": "` + machinetype + `"}`

	headers := make(map[string]string)
	headers["X-CSRFToken"] = servicetoken
	headers["Cookie"] = "session=" + lv.serviceCookie
	headers["Content-Type"] = "application/json"
	headers["Authorization"] = "Basic " + authtoken

	body, err := lv.makeLaundryServiceWebRequest(url, method, headers, payload)
	if err != nil {
		return nil, err
	}

	var dat map[string][]string
	if err := json.Unmarshal(body, &dat); err != nil {
		return nil, err
	}

	return dat["problemCodeList"], nil
}

func (lv *CSCLaundryView) submitTicket(serviceurl string, subscriptionkey string, servicetoken string, authtoken string, machineid string, problemCode string, comments string, firstName string, lastName string, phone string, email string) (*model.ServiceRequestResult, error) {
	url := serviceurl + "/submitServiceRequest?subscription-key=" + subscriptionkey
	method := "POST"
	headers := make(map[string]string)
	headers["X-CSRFToken"] = servicetoken
	headers["Cookie"] = "session=" + lv.serviceCookie
	headers["Content-Type"] = "application/json"
	headers["Authorization"] = "Basic " + authtoken

	payload := struct {
		MachineID   string `json:"machineId"`
		ProblemCode string `json:"problemCode"`
		Comments    string `json:"comments"`
		FirstName   string `json:"firstName"`
		LastName    string `json:"lastName"`
		Phone       string `json:"phone"`
		Email       string `json:"email"`
	}{
		MachineID:   machineid,
		ProblemCode: problemCode,
		Comments:    comments,
		FirstName:   firstName,
		LastName:    lastName,
		Phone:       phone,
		Email:       email,
	}

	postData, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	body, err := lv.makeLaundryServiceWebRequest(url, method, headers, string(postData))

	if err != nil {
		return nil, err
	}

	var obj interface{}

	if err := json.Unmarshal(body, &obj); err != nil {
		return nil, err
	}

	m := obj.(map[string]interface{})
	//already a request for this machine, so got back a machine details object
	if m["machineId"] != nil {
		result := model.ServiceRequestResult{Message: "A ticket already exists for this machine", RequestNumber: "0", Status: "Failed"}
		return &result, nil
	}
	result := model.ServiceRequestResult{Message: m["message"].(string), RequestNumber: m["serviceRequestNumber"].(string), Status: "Success"}
	return &result, nil

}

func (lv *CSCLaundryView) makeLaundryServiceWebRequest(url string, method string, headers map[string]string, postParams string) ([]byte, error) {
	payload := strings.NewReader(postParams)
	client := http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		return nil, err
	}

	for headername, headerval := range headers {
		req.Header.Add(headername, headerval)
	}

	res, err := client.Do(req)

	if err != nil {
		log.Printf("%v", err.Error())
		return nil, err
	}

	defer res.Body.Close()
	for _, cookie := range res.Cookies() {
		if cookie.Name == "session" {
			lv.serviceCookie = cookie.Value
		}
	}

	if res.StatusCode != 200 {
		test, err := io.ReadAll(res.Body)
		log.Printf("%v", string(test))
		return nil, err
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}
func evalNumAvailable(inputstr string) int {
	if i, err := strconv.Atoi(inputstr); err == nil {
		return i
	}
	return 0
}
