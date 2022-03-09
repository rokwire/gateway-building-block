package laundry

import (
	model "apigateway/core/model/laundry"
	"encoding/xml"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
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

//CSCLaundryView is a vendor specific structure that implements the Laundry interface
type CSCLaundryView struct {
	//configuration information (url, api keys...gets passed into here)
	APIKey string
	APIUrl string
}

//NewCSCLaundryAdapter returns a vendor specific implementation of the Laundry interface
func NewCSCLaundryAdapter(apikey string, url string) *CSCLaundryView {
	return &CSCLaundryView{APIKey: apikey, APIUrl: url}

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
