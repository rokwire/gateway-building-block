package model

//LaundryRoom represents the basic information returned as part of requesting and organization
type LaundryRoom struct {
	ID       int
	Name     string
	Status   string
	Location *LaundryDetails
}

//Organization represents the top most level of inforomation provided by teh laundry api
type Organization struct {
	SchoolName   string
	LaundryRooms []*LaundryRoom
}

//RoomDetail represents details about a specific laundry room, including a list of appliances
type RoomDetail struct {
	NumWashers int
	NumDryers  int
	Appliances []*Appliance
	RoomName   string
	CampusName string
	Location   *LaundryDetails
}

//Appliance represents the information specific to an identifiiable appliance in a laundry room
type Appliance struct {
	ID               string
	Status           string
	ApplianceType    string
	AverageCycleTime int
	TimeRemaining    *int
	Label            string
}

//MachineRequestDetail represents the basic details needed in order to submit a request about a machine
type MachineRequestDetail struct {
	MachineID    string
	Message      string
	OpenIssue    bool
	ProblemCodes []string
	MachineType  string
}

//ServiceRequestResult represents the information returned upon submission of a machine service request
type ServiceRequestResult struct {
	Message       string
	RequestNumber string
	Status        string
}

//ServiceSubmission represents the data required to submit a service request for a laundry machine
type ServiceSubmission struct {
	MachineID   *string `json:"machineid"`
	ProblemCode *string `json:"problemcode"`
	Comments    *string `json:"comments"`
	FirstName   *string `json:"firstname"`
	LastName    *string `json:"lastname"`
	Phone       *string `json:"phone"`
	Email       *string `json:"email"`
}
