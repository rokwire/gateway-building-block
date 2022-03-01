package laundry

import (
	model "apigateway/core/model/laundry"
	"encoding/xml"
	"io/ioutil"
	"log"
	"net/http"
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
	Location   string   `xml:"location"`
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

	log.Printf("now in ListRooms")
	log.Printf(lv.APIKey + ":" + lv.APIUrl)
	url := lv.APIUrl + "/school/?api_key=" + lv.APIKey + "&method=getRoomData&type=json"
	log.Printf(url)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.Status == "200 OK" {
		log.Printf("requests succeeded")
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
	/*
		org := model.Organization{SchoolName: "hello World"}
		org.LaundryRooms = make([]*model.LaundryRoom, 0)
		org.LaundryRooms = append(org.LaundryRooms, newLaundryRoom(1, "clint", "open"))
		return &org, nil
	*/
}

func newLaundryRoom(id int, name string, status string) *model.LaundryRoom {
	lr := model.LaundryRoom{Name: name, ID: id, Status: status}
	return &lr
}

//GetLaundryRoom returns the room details along with the list of machines in that room
func (lv *CSCLaundryView) GetLaundryRoom(roomid int) (*model.RoomDetail, error) {
	rd := model.RoomDetail{}
	//code here to make the web call and return the xml as a room detail object
	return &rd, nil
}
