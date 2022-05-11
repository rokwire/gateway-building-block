package model

//LaundryAssets represents the laundry elements of assets.json
type LaundryAssets struct {
	Assets []LaundryAsset `json:"locations"`
}

//LaundryAsset represents a single laundry room asset
type LaundryAsset struct {
	LocationID string         `json:"laundry_location"`
	Details    LaundryDetails `json:"location_details"`
}

//LaundryDetails represents the location details of a single laundry room asset
type LaundryDetails struct {
	Latitude  float32 `json:"latitude"`
	Longitude float32 `jsong:"longitude"`
	Floor     int     `json:"floor"`
}
