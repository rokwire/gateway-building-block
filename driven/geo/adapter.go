package geo

import (
	"context"
	"log"
	"strings"

	"github.com/rokwire/logging-library-go/v2/logs"
	"googlemaps.github.io/maps"
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
	"General Events":               {0.0, 0.0},
	"Krannert Center":              {40.1080244, -88.224704},
	"Library Calendar":             {0.0, 0.0},
	"Facility Hours":               {0.0, 0.0},
	"Beckman Main Calendar":        {40.1157707, -88.229393},
	"Lincoln Hall Theater Events":  {40.1066066, -88.2304212},
	"Foellinger Auditorium Events": {40.1059431, -88.2294751},
	"Department of Sociology":      {40.1066528, -88.2305061},
	"NCSA":                         {40.1147743, -88.2252053},
}

// Adapter implements the GeoAdapter interface
type Adapter struct {
	googleMapsClient maps.Client

	log logs.Log
}

// TODO todo
func (na Adapter) TODO(location string) {
	entry := make(map[string]interface{})

	// Подготвяме заявката за геокодиране
	req := &maps.GeocodingRequest{
		Address: location + ", Urbana",
		Components: map[maps.Component]string{
			maps.ComponentAdministrativeArea: "Urbana",
			maps.ComponentCountry:            "US",
		},
	}

	// Извършваме заявката
	resp, err := na.googleMapsClient.Geocode(context.Background(), req)
	if err != nil {
		log.Printf("API Key Error: %v", err)
		entry["location"] = map[string]string{"description": location}
		// Тук трябва да добавите записа към базата данни MongoDB или друга структура
		return
	}

	if len(resp) != 0 {
		lat := resp[0].Geometry.Location.Lat
		lng := resp[0].Geometry.Location.Lng
		geoInfo := map[string]interface{}{
			"latitude":    lat,
			"longitude":   lng,
			"description": location,
		}
		entry["location"] = geoInfo
	} else {
		entry["location"] = map[string]string{"description": location}
		log.Printf("calendarId: %s, dataSourceEventId: %s, location: %s geolocation not found",
			entry["calendarId"], entry["dataSourceEventId"], location)
	}
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

func fetchGeoData(client *maps.Client, location string, entry map[string]interface{}) {

}

/*
func main() {
	found, geoInfo := searchStaticLocation("Krannert Center", "", "studio 5")
	if found {
		println("GeoInfo found:", geoInfo.Description, geoInfo.Latitude, geoInfo.Longitude)
	} else {
		println("No GeoInfo found")
	}

	client, err := maps.NewClient(maps.WithAPIKey("TODO"))
	  if err != nil {
	      log.Fatalf("Error creating client: %v", err)
	  }

	  entry := make(map[string]interface{})
	  fetchGeoData(client, "123 Main St, Urbana", entry)
} */

// NewGeoBBAdapter creates new instance
func NewGeoBBAdapter(googleAPIKey string, logger *logs.Logger) Adapter {
	l := logger.NewLog("geo_bb_adapter", logs.RequestContext{})

	client, err := maps.NewClient(maps.WithAPIKey(googleAPIKey))
	if err != nil {
		log.Fatalf("Error creating google maps client: %v", err)
	}

	return Adapter{
		googleMapsClient: *client,
		log:              *l,
	}
}
