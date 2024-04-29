package geo

import (
	"strings"
)

// Tip is a struct to hold localization tips
type Tip struct {
	CalendarName    string
	SponsorKeyword  string
	LocationKeyword string
	AccessName      string
}

// GeoInfo is a struct to hold the geolocation information
type GeoInfo struct {
	Latitude    float64
	Longitude   float64
	Description string
}

var tip4CalALoc = []Tip{
	{CalendarName: "Krannert Center", SponsorKeyword: "", LocationKeyword: "studio", AccessName: "Krannert Center"},
	{CalendarName: "Krannert Center", SponsorKeyword: "", LocationKeyword: "stage", AccessName: "Krannert Center"},
	{CalendarName: "General Events", SponsorKeyword: "ncsa", LocationKeyword: "ncsa", AccessName: "NCSA"},
	{CalendarName: "National Center for Supercomputing Applications master calendar", SponsorKeyword: "", LocationKeyword: "ncsa", AccessName: "NCSA"},
}

var CalName2Location = map[string][2]float64{
	"General Events":               {},
	"Krannert Center":              {40.1080244, -88.224704},
	"Library Calendar":             {},
	"Facility Hours":               {},
	"Beckman Main Calendar":        {40.1157707, -88.229393},
	"Lincoln Hall Theater Events":  {40.1066066, -88.2304212},
	"Foellinger Auditorium Events": {40.1059431, -88.2294751},
	"Department of Sociology":      {40.1066528, -88.2305061},
	"NCSA":                         {40.1147743, -88.2252053},
}

// searchStaticLocation looks for a static location based on the calendar name, sponsor, and location description
func searchStaticLocation(calendarName, sponsor, location string) (bool, *GeoInfo) {
	for _, tip := range tip4CalALoc {
		if tip.CalendarName == calendarName &&
			strings.Contains(strings.ToLower(sponsor), tip.SponsorKeyword) &&
			strings.Contains(strings.ToLower(location), tip.LocationKeyword) {
			latLong, exists := CalName2Location[tip.AccessName]
			if !exists {
				return false, nil
			}
			geoInfo := GeoInfo{
				Latitude:    latLong[0],
				Longitude:   latLong[1],
				Description: location,
			}
			return true, &geoInfo
		}
	}
	return false, nil
}

func main() {
	found, geoInfo := searchStaticLocation("Krannert Center", "", "studio 5")
	if found {
		println("GeoInfo found:", geoInfo.Description, geoInfo.Latitude, geoInfo.Longitude)
	} else {
		println("No GeoInfo found")
	}
}
