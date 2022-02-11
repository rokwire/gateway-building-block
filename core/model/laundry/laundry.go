package model

type LaundryRoom struct {
	Id     int
	Name   string
	Status string
}

type RoomDetail struct {
	NumWashers int
	NumDryers  int
	Appliances []*Appliance
}

type Appliance struct {
	Id                 int
	Status             string
	Name               string
	Average_cycle_time int
	Time_remaining     string
	Out_of_service     bool
}

type Organization struct {
	SchoolName   string
	LaundryRooms []*LaundryRoom
}
