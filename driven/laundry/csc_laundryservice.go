package laundry

import (
	model "apigateway/core/model/laundry"
)

type MachineDetail struct {
	Address             string `json:"address"`
	LaundryLocation     string `json:"laundryLocaiton"`
	MachineId           string `json:"machineId"`
	MachineType         string `json:"machineType"`
	Message             string `json:"message"`
	Property            string `json:"property"`
	RecentServiceDate   string `json:"recentServiceDate"`
	RecentServiceNotes  string `json:"recentServiceNotes"`
	RecentServiceStatus string `json:"recentServiceStatus"`
	SiteID              string `json:"siteID"`
}

type ServiceRequest struct {
	SubscriptionKey string
	ApiKey          string
	Token           string
	Details         MachineDetail
	ProblemCodes    []string
	CookieValue     string
}

func newServiceRequest(subscriptionKey string) *ServiceRequest {
	sr := ServiceRequest{SubscriptionKey: subscriptionKey}
	return &sr
}

func (sr *ServiceRequest) InitiateRequest(machineid string) (*model.MachineRequestDetail, error) {
	sr.GetAPIKey()
	sr.GetToken()
	sr.GetMachineDetails(machineid)
	sr.GetProblemCodes(sr.Details.MachineType)
	mrd := model.MachineRequestDetail{}
	//map current ServiceRequest properties to MachineRequestDetail Properties
	return &mrd, nil
}

func (sr *ServiceRequest) GetAPIKey() error {
	//make calls to get api key - sets value of sr.ApiKey
	return nil
}

func (sr *ServiceRequest) GetToken() error {
	//makes call to get token and session cookie
	return nil
}

func (sr *ServiceRequest) GetMachineDetails(machineId string) error {
	return nil
}

func (sr *ServiceRequest) GetProblemCodes(machinetype string) error {
	return nil
}
