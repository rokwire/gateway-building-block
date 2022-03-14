package laundry

import (
	model "apigateway/core/model/laundry"
	"encoding/json"
	"encoding/xml"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type appliance struct {
	XMLName       xml.Name `xml:"appliance"`
	ApplianceKey  string   `xml:"appliance_desc_key"`
	LrmStatus     string   `xml:"lrm_status"`
	ApplianceType string   `xml:"appliance_type"`
	Status        string   `xml:"status"`
	OutOfService  string   `xml:"out_of_service"`
	Name          string   `xml:"label"`
	AvgCycleTime  string   `xml:"avg_cycle_time"`
	TimeRemaining string   `xml:"time_remaining"`
}

type laundryroom struct {
	XMLName    xml.Name     `xml:"laundry_room"`
	Name       string       `xml:"laundry_room_name"`
	CampusName string       `xml:"campus_name"`
	Appliances []*appliance `xml:"appliances>appliance"`
}

type laundrylocation struct {
	Location        int      `xml:"location"`
	XMLName         xml.Name `xml:"laundryroom"`
	Campusname      string   `xml:"campus_name"`
	Laundryroomname string   `xml:"laundry_room_name"`
	Status          string   `xml:"status"`
}

type school struct {
	XMLName      xml.Name           `xml:"school"`
	SchoolName   string             `xml:"school_name"`
	LaundryRooms []*laundrylocation `xml:"laundry_rooms>laundryroom"`
}

type capacity struct {
	XMLName    xml.Name `xml:"laundry_room"`
	NumWashers string   `xml:"washer"`
	NumDryers  string   `xml:"dryer"`
}

type machinedetail struct {
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

//CSCLaundryView is a vendor specific structure that implements the Laundry interface
type CSCLaundryView struct {
	//configuration information (url, api keys...gets passed into here)
	APIKey string
	APIUrl string

	ServiceOCPSubscriptionKey string
	ServiceAPIUrl             string
	serviceToken              string
	serviceSubscriptionKey    string
	serviceCookie             string
}

//NewCSCLaundryAdapter returns a vendor specific implementation of the Laundry interface
func NewCSCLaundryAdapter(apikey string, url string, subscriptionkey string, serviceapiurl string) *CSCLaundryView {
	return &CSCLaundryView{APIKey: apikey, APIUrl: url, ServiceOCPSubscriptionKey: subscriptionkey, ServiceAPIUrl: serviceapiurl}

}

//ListRooms lists the laundry rooms
func (lv *CSCLaundryView) ListRooms() (*model.Organization, error) {

	url := lv.APIUrl + "/school/?api_key=" + lv.APIKey + "&method=getRoomData&type=json"
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.Status == "200 OK" {
		body, bodyerr := ioutil.ReadAll(resp.Body)
		if bodyerr != nil {
			return nil, err
		}

		var nS school
		out := []byte(body)
		if err := xml.Unmarshal(out, &nS); err != nil {
			log.Fatal("could not unmarshal xml data")
			return nil, err
		}
		org := model.Organization{SchoolName: nS.SchoolName}
		org.LaundryRooms = make([]*model.LaundryRoom, 0)

		for _, lr := range nS.LaundryRooms {
			org.LaundryRooms = append(org.LaundryRooms, newLaundryRoom(lr.Location, lr.Laundryroomname, lr.Status))
		}
		return &org, nil
	}
	return nil, err
}

//GetLaundryRoom returns the room details along with the list of machines in that room
func (lv *CSCLaundryView) GetLaundryRoom(roomid string) (*model.RoomDetail, error) {

	url := lv.APIUrl + "/room/?api_key=" + lv.APIKey + "&method=getAppliances&location=" + roomid + "&type=json"
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.Status == "200 OK" {
		body, bodyerr := ioutil.ReadAll(resp.Body)
		if bodyerr != nil {
			return nil, err
		}

		var lr laundryroom
		out := []byte(body)
		if err := xml.Unmarshal(out, &lr); err != nil {
			log.Fatal("could not unmarshal xml data")
			return nil, err
		}

		rd := model.RoomDetail{CampusName: lr.CampusName, RoomName: lr.Name}
		rd.Appliances = make([]*model.Appliance, len(lr.Appliances))
		roomCapacity, _ := lv.getNumAvailable(roomid)
		rd.NumDryers = evalNumAvailable(roomCapacity.NumDryers)
		rd.NumWashers = evalNumAvailable(roomCapacity.NumWashers)

		for i, appl := range lr.Appliances {
			avgCycle, _ := strconv.Atoi(appl.AvgCycleTime)
			outOfService, _ := strconv.ParseBool(appl.OutOfService)
			rd.Appliances[i] = newAppliance(appl.ApplianceKey, appl.ApplianceType, avgCycle, appl.Status, appl.TimeRemaining, outOfService)
		}
		return &rd, nil
	}
	return nil, err
}

//InitServiceRequest gets machine request details needed to initialize a laundry service request
func (lv *CSCLaundryView) InitServiceRequest(machineID string) (*model.MachineRequestDetail, error) {

	err := lv.getServiceSubscriptionKey()
	if err != nil {
		return nil, err
	}

	err = lv.getServiceToken()

	if err != nil {
		return nil, err
	}

	md, err := lv.getMachineDetails(machineID)
	if err != nil {
		return nil, err
	}

	mrd := newMachineRequestDetail(md.MachineID, md.Message, md.RecentServiceStatus)
	log.Printf(mrd.MachineID + " " + mrd.Message)
	mrd.ProblemCodes, err = lv.getProblemCodes(md.MachineType)
	if err != nil {
		return nil, err
	}
	return mrd, nil
}

//SubmitServiceRequest submits a request for a machine
func (lv *CSCLaundryView) SubmitServiceRequest(machineid string, problemCode string, comments string, firstName string, lastName string, phone string) (*model.ServiceRequestResult, error) {

	err := lv.getServiceSubscriptionKey()
	if err != nil {
		return nil, err
	}

	err = lv.getServiceToken()

	if err != nil {
		return nil, err
	}

	srr, err := lv.submitTicket(machineid, problemCode, comments, firstName, lastName, phone)
	if err != nil {
		return nil, err
	}
	return srr, nil
}

func newMachineRequestDetail(machineid string, message string, serviceStatus string) *model.MachineRequestDetail {
	var openTicket = serviceStatus == "Open"
	mrd := model.MachineRequestDetail{MachineID: machineid, Message: message, OpenIssue: openTicket}
	return &mrd
}

func newLaundryRoom(id int, name string, status string) *model.LaundryRoom {
	lr := model.LaundryRoom{Name: name, ID: id, Status: status}
	return &lr
}

func (lv *CSCLaundryView) getNumAvailable(roomid string) (*capacity, error) {

	url := lv.APIUrl + "/room/?api_key=" + lv.APIKey + "&method=getNumAvailable&location=" + roomid + "&type=json"
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.Status == "200 OK" {
		body, bodyerr := ioutil.ReadAll(resp.Body)
		if bodyerr != nil {
			return nil, err
		}

		var cap capacity
		out := []byte(body)
		if err := xml.Unmarshal(out, &cap); err != nil {
			log.Fatal("could not unmarshal xml data")
			return nil, err
		}

		return &cap, nil
	}
	return nil, err
}

func newAppliance(id string, appliancetype string, cycletime int, status string, timeremaining string, outofservice bool) *model.Appliance {
	appl := model.Appliance{ID: id, ApplianceType: appliancetype, AverageCycleTime: cycletime, Status: status, TimeRemaining: timeremaining, OutofService: outofservice}
	return &appl
}

func evalNumAvailable(inputstr string) int {
	if i, err := strconv.Atoi(inputstr); err == nil {
		return i
	}
	return 0
}

func (lv *CSCLaundryView) getServiceSubscriptionKey() error {
	url := lv.ServiceAPIUrl + "/sr-key/getSubscriptionKey"
	method := "POST"

	payload := `{"subscription-id": "univofchicago", "key-type": "primaryKey" }`

	headers := make(map[string]string)
	headers["Ocp-Apim-Subscription-Key"] = lv.ServiceOCPSubscriptionKey
	headers["Content-Type"] = "application/json"

	body, err := lv.makeLaundryServiceWebRequest(url, method, headers, payload)
	if err != nil {
		return err
	}
	lv.serviceSubscriptionKey = string(body)
	return nil
}

func (lv *CSCLaundryView) getServiceToken() error {
	url := lv.ServiceAPIUrl + "/sr/v1/generateToken?subscription-key=" + lv.serviceSubscriptionKey
	method := "GET"

	headers := make(map[string]string)
	headers["Content-Type"] = "application/json"

	body, err := lv.makeLaundryServiceWebRequest(url, method, headers, "")

	if err != nil {
		return err
	}

	var dat map[string]interface{}
	if err := json.Unmarshal(body, &dat); err != nil {
		return err
	}

	lv.serviceToken = dat["token"].(string)
	return nil
}

func (lv *CSCLaundryView) makeLaundryServiceWebRequest(url string, method string, headers map[string]string, postParams string) ([]byte, error) {
	log.Print("Making laundry service web call")
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

	log.Printf("Getting Cookie Value")
	for _, cookie := range res.Cookies() {
		if cookie.Name == "session" {
			lv.serviceCookie = cookie.Value
		}
	}

	log.Printf("setting body value")
	if res.StatusCode != 200 {
		_, err := ioutil.ReadAll(res.Body)
		return nil, err
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func (lv *CSCLaundryView) getMachineDetails(machineid string) (*machinedetail, error) {
	md := machinedetail{}

	url := lv.ServiceAPIUrl + "/sr/v1/machineDetails?subscription-key=" + lv.serviceSubscriptionKey
	method := "POST"

	payload := `{"machineId":"` + machineid + `"}`

	headers := make(map[string]string)
	headers["X-CSRFToken"] = lv.serviceToken
	headers["Cookie"] = "session=" + lv.serviceCookie
	headers["Content-Type"] = "application/json"

	body, err := lv.makeLaundryServiceWebRequest(url, method, headers, payload)

	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(body, &md); err != nil {
		return nil, err
	}
	return &md, nil
}

func (lv *CSCLaundryView) getProblemCodes(machinetype string) ([]string, error) {
	url := lv.ServiceAPIUrl + "/sr/v1/problemCodes?subscription-key=" + lv.serviceSubscriptionKey
	method := "POST"

	payload := `{"machineType": "` + machinetype + `"}`

	headers := make(map[string]string)
	headers["X-CSRFToken"] = lv.serviceToken
	headers["Cookie"] = "session=" + lv.serviceCookie
	headers["Content-Type"] = "application/json"

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

func (lv *CSCLaundryView) submitTicket(machineid string, problemCode string, comments string, firstName string, lastName string, phone string) (*model.ServiceRequestResult, error) {
	url := lv.ServiceAPIUrl + "/sr/v1/submitServiceRequest?subscription-key=" + lv.serviceSubscriptionKey
	method := "POST"
	headers := make(map[string]string)
	headers["X-CSRFToken"] = lv.serviceToken
	headers["Cookie"] = "session=" + lv.serviceCookie
	headers["Content-Type"] = "application/json"

	payload := `{"machineId": "` + machineid + `", "problemCode": "` + problemCode + `", "comments": "` + comments + `", "firstName": "` + firstName + `", "lastName": "` + lastName + `", "phone": "` + phone + `"}`

	body, err := lv.makeLaundryServiceWebRequest(url, method, headers, payload)

	if err != nil {
		return nil, err
	}

	var obj interface{}

	if err := json.Unmarshal(body, &obj); err != nil {
		return nil, err
	}

	m := obj.(map[string]interface{})
	//already a request for this machine, so got bsack a machine details object
	if m["machineId"] != nil {
		result := model.ServiceRequestResult{Message: "A ticket already exists for this machien", RequestNumber: 0, Status: "Success"}
		return &result, nil
	}
	result := model.ServiceRequestResult{Message: m["message"].(string), RequestNumber: m["serviceRequestNumber"].(int), Status: m["status"].(string)}
	return &result, nil

}
