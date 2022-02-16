package model

type LaundryRoom struct {
	Id     int
	Name   string
	Status string
}

type Organization struct {
	SchoolName   string
	LaundryRooms []*LaundryRoom
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

type MachineRequestDetail struct {
	MachineId           string
	Message             string
	RecentServiceStatus bool
	ProblemCodes        []string
}

type ServiceRequestResult struct {
	Message       string
	RequestNumber int
	Status        string
}
