package model

type Entrance struct {
	ID           string
	Name         string
	ADACompliant bool
	Available    bool
	ImageURL     string
	Latitude     float32
	Longitude    float32
}

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
}
