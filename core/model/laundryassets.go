// Copyright 2022 Board of Trustees of the University of Illinois.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
	Longitude float32 `json:"longitude"`
	Floor     int     `json:"floor"`
}
