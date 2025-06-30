package geo

import (
	"application/core/model"
	"context"
	"log"

	"github.com/google/uuid"
	"github.com/rokwire/rokwire-building-block-sdk-go/utils/logging/logs"
	"googlemaps.github.io/maps"
)

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

/*
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
} */

// Adapter implements the GeoAdapter interface
type Adapter struct {
	googleMapsClient maps.Client

	log logs.Log
}

// FindLocation finds the location the location
func (l Adapter) FindLocation(location string) (*model.LegacyLocation, error) {
	return l.findLocationFromGoogle(location)
}

func (l Adapter) findLocationFromGoogle(location string) (*model.LegacyLocation, error) {
	req := &maps.GeocodingRequest{
		Address: location + ", Urbana",
		Components: map[maps.Component]string{
			maps.ComponentAdministrativeArea: "Urbana",
			maps.ComponentCountry:            "US",
		},
	}

	resp, err := l.googleMapsClient.Geocode(context.Background(), req)
	if err != nil {
		log.Printf("API Key Error: %v", err)
		return nil, nil //not found on error
	}

	if len(resp) != 0 {
		lat := resp[0].Geometry.Location.Lat
		lng := resp[0].Geometry.Location.Lng
		legacyLocation := model.LegacyLocation{ID: uuid.NewString(), Name: location,
			Description: location, Lat: &lat, Long: &lng}

		return &legacyLocation, nil
	}
	log.Printf("not found: %s", location)

	return nil, nil
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
