package laundry

import (
	model "apigateway/core/model/laundry"
)

type machineDetail struct {
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

//ServiceRequest represents the access and data provided by the laundry service provider to facilitate submitting a service request
type ServiceRequest struct {
	SubscriptionKey string
	APIKey          string
	Token           string
	Details         machineDetail
	ProblemCodes    []string
	CookieValue     string
}

//NewServiceRequest initializes a new instance of a ServiceRequest struct
func NewServiceRequest(subscriptionKey string) *ServiceRequest {
	sr := ServiceRequest{SubscriptionKey: subscriptionKey}
	return &sr
}

//InitiateRequest returns the data necessary for reporting a problem
func (sr *ServiceRequest) InitiateRequest(machineid string) (*model.MachineRequestDetail, error) {
	sr.getAPIKey()
	sr.getToken()
	sr.getMachineDetails(machineid)
	sr.getProblemCodes(sr.Details.MachineType)
	mrd := model.MachineRequestDetail{}
	//map current ServiceRequest properties to MachineRequestDetail Properties
	return &mrd, nil
}

func (sr *ServiceRequest) getAPIKey() error {
	//make calls to get api key - sets value of sr.ApiKey
	return nil
}

func (sr *ServiceRequest) getToken() error {
	//makes call to get token and session cookie
	return nil
}

func (sr *ServiceRequest) getMachineDetails(machineID string) error {
	return nil
}

func (sr *ServiceRequest) getProblemCodes(machinetype string) error {
	return nil
}
