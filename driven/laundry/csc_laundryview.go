package laundry

import (
	model "apigateway/core/model/laundry"
	"encoding/xml"
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

type LaundryRoom struct {
	XMLName    xml.Name     `xml:"laundry_room"`
	Name       string       `xml:"laundry_room_name"`
	CampusName string       `xml:"campus_name"`
	Appliances []*appliance `xml:"appliances>appliance"`
}

type LaundryLocation struct {
	Location          int      `xml:"location"`
	XMLName           xml.Name `xml:"laundryroom"`
	Campus_name       string   `xml:"campus_name"`
	Luandry_room_name string   `xml:"laundry_room_name"`
	Status            string   `xml:"status"`
}

type School struct {
	XMLName      xml.Name           `xml:"school"`
	SchoolName   string             `xml:"school_name"`
	LaundryRooms []*LaundryLocation `xml:"laundry_rooms>laundryroom"`
}

type Capacity struct {
	Location   string   `xml:"location"`
	XMLName    xml.Name `xml:"laundryroom"`
	NumWashers string   `xml:"available_washers"`
	NumDryers  string   `xml:"available_dryers"`
}

type Capacities struct {
	XMLName        xml.Name    `xml:"laundry_rooms"`
	RoomCapacities []*Capacity `xml:"laundryroom"`
}

type CSCLaundryView struct {
	//configuration information (url, api keys...gets passed into here)
}

func (lv *CSCLaundryView) ListRooms() (*model.Organization, error) {
	org := model.Organization{}
	//code here to make the web call and transform the xml into an organization object
	return &org, nil
}

func (lv *CSCLaundryView) GetLaundryRoom(roomid int) (*model.RoomDetail, error) {
	rd := model.RoomDetail{}
	//code here to make the web call and return the xml as a room detail object
	return &rd, nil
}
