package geo

import (
	"application/core/model"
	"context"
	"log"

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

/*
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
}*/

var DefinedLocation = map[string][2]float64{
	"Davenport 109A": {40.107335, -88.226069},
	"Nevada Dance Studio (905 W. Nevada St.)":                       {40.105825, -88.219873},
	"18th Ave Library, 175 W 18th Ave, Room 205, Oklahoma City, OK": {36.102183, -97.111245},
	"Champaign County Fairgrounds":                                  {40.1202191, -88.2178757},
	"Student Union SLC Conference room":                             {39.727282, -89.617477},
	"Armory, room 172 (the Innovation Studio)":                      {40.104749, -88.23195},
	"Student Union Room 235":                                        {39.727282, -89.617477},
	"Uni 206, 210, 211":                                             {40.11314, -88.225259},
	"Uni 205, 206, 210":                                             {40.11314, -88.225259},
	"Southern Historical Association Combs Chandler 30":             {38.258116, -85.756139},
	"St. Louis, MO":                                                 {38.694237, -90.4493},
	"Student Union SLC":                                             {39.727282, -89.617477},
	"Purdue University, West Lafayette, Indiana":                    {40.425012, -86.912645},
	"MP 7":                  {40.100803, -88.23604},
	"116 Roger Adams Lab":   {40.107741, -88.224943},
	"2700 Campus Way 45221": {39.131894, -84.519143},
	"The Orange Room, Main Library - 1408 W. Gregory Drive, Champaign IL": {40.1047044, -88.22901039999999},

	//CalName2Location
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

// ProcessLocation process the location
func (l Adapter) ProcessLocation(eventID, calendarName, sponsor, location string) (*model.LegacyLocation, error) {

	var legacyLocation *model.LegacyLocation

	if location == "" {
		legacyLocation = &model.LegacyLocation{
			ID:          eventID, //arh...
			Name:        calendarName,
			Description: location,
			Lat:         nil,
			Long:        nil,
		}
		return legacyLocation, nil
	}

	for name, cords := range DefinedLocation {
		if location == name {
			legacyLocation = &model.LegacyLocation{
				ID:          eventID, //arh...
				Name:        name,
				Description: location,
				Lat:         &cords[0],
				Long:        &cords[1],
			}
			return legacyLocation, nil
		}
	}

	_, statiLocation := searchStaticLocation(calendarName, sponsor, location)
	if statiLocation != nil {
		legacyLocation = &model.LegacyLocation{
			ID:          eventID, //arh...
			Name:        calendarName,
			Description: statiLocation.Description,
			Lat:         &statiLocation.Latitude,
			Long:        &statiLocation.Longitude,
		}
		return legacyLocation, nil
	}

	locationFromGoogle, _ := l.findLocationFromGoogle(nil, location, eventID, calendarName)
	if locationFromGoogle != nil {
		legacyLocation = &model.LegacyLocation{
			ID:          locationFromGoogle.ID, //arh...
			Name:        locationFromGoogle.Name,
			Description: locationFromGoogle.Description,
			Lat:         locationFromGoogle.Lat,
			Long:        locationFromGoogle.Long,
		}
		return legacyLocation, nil
	}

	// Default return (if none of the conditions are met)
	return legacyLocation, nil
}

// searchStaticLocation looks for a static location based on the calendar name, sponsor, and location description
func searchStaticLocation(calendarName, sponsor, location string) (bool, *GeoInfo) {
	/*for _, tip := range tip4CalALoc {
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
	return false, nil */

	return false, nil

}

func (l Adapter) findLocationFromGoogle(client *maps.Client, location string, eventID string, calendarName string) (*model.LegacyLocation, error) {
	entry := make(map[string]interface{})
	var legacyLocation model.LegacyLocation
	// Подготвяме заявката за геокодиране
	req := &maps.GeocodingRequest{
		Address: location + ", Urbana",
		Components: map[maps.Component]string{
			maps.ComponentAdministrativeArea: "Urbana",
			maps.ComponentCountry:            "US",
		},
	}

	// Извършваме заявката
	resp, err := l.googleMapsClient.Geocode(context.Background(), req)
	if err != nil {
		log.Printf("API Key Error: %v", err)
		entry["location"] = map[string]string{"description": location}
		return nil, nil
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
		legacyLocation = model.LegacyLocation{ID: eventID, Name: calendarName,
			Description: location, Lat: &lat, Long: &lng}
	} else {
		entry["location"] = map[string]string{"description": location}
		log.Printf("calendarId: %s, dataSourceEventId: %s, location: %s geolocation not found",
			entry["calendarId"], entry["dataSourceEventId"], location)
	}
	return &legacyLocation, nil
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
