package model

//Entrance represents the information returned when the closest entrance of a building is requested
type Entrance struct {
	ID           string
	Name         string
	ADACompliant bool
	Available    bool
	ImageURL     string
	Latitude     float32
	Longitude    float32
}

//Building represents the information returned when a requst for a building's details is made
type Building struct {
	ID          string
	Name        string
	Number      string
	FullAddress string
	Address1    string
	Address2    string
	City        string
	State       string
	ZipCode     string
	ImageURL    string
	MailCode    string
	Entrances   []Entrance
	Latitude    float32
	Longitude   float32
}
