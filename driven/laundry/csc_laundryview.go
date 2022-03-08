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
	AvgCycleType  string   `xml:"avg_cycle_time"`
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
	Location   int      `xml:"location"`
	XMLName    xml.Name `xml:"laundryroom"`
	NumWashers string   `xml:"available_washers"`
	NumDryers  string   `xml:"available_dryers"`
}

type capacities struct {
	XMLName        xml.Name    `xml:"laundry_rooms"`
	RoomCapacities []*capacity `xml:"laundryroom"`
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
		roomCapacities, _ := lv.getNumAvailable()

		for _, lr := range nS.LaundryRooms {
			washers, dryers := findRoomCapacity(lr.Location, *roomCapacities)
			org.LaundryRooms = append(org.LaundryRooms, newLaundryRoom(lr.Location, lr.Laundryroomname, lr.Status, washers, dryers))
		}
		return &org, nil
	}
	return nil, err
	/*
		org := model.Organization{SchoolName: "hello World"}
		org.LaundryRooms = make([]*model.LaundryRoom, 0)
		org.LaundryRooms = append(org.LaundryRooms, newLaundryRoom(1, "clint", "open"))
		return &org, nil
	*/
}

func findRoomCapacity(roomid int, rc capacities) (washers int, dryers int) {
	numWashers := 0
	numDryers := 0
	for _, v := range rc.RoomCapacities {
		if v.Location == roomid {
			if i, err := strconv.Atoi(v.NumWashers); err == nil {
				numWashers = i
			}
			if j, err := strconv.Atoi(v.NumDryers); err == nil {
				numDryers = j
			}
			return numWashers, numDryers
		}
	}
	return numWashers, numDryers
}

func newLaundryRoom(id int, name string, status string, numwashers int, numdryers int) *model.LaundryRoom {
	lr := model.LaundryRoom{Name: name, ID: id, Status: status, AvialableWashers: numwashers, AvailableDryers: numdryers}
	return &lr
}

func (lv *CSCLaundryView) getNumAvailable() (*capacities, error) {

	url := lv.APIUrl + "/school/?api_key=" + lv.APIKey + "&method=getNumAvailable&type=json"
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

		var cap capacities
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
func (lv *CSCLaundryView) GetLaundryRoom(roomid int) (*model.RoomDetail, error) {
	rd := model.RoomDetail{}
	return &rd, nil
}
