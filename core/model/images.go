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

// ContentImagesURL is used to keep the imageURL from ContentBB
type ContentImagesURL struct {
	ID       string `json:"id" bson:"_id"`
	ImageURL string `json:"imageURL" bson:"imageURL"`
}

// ImageData is used to keep the the image thata from webtools
type ImageData struct {
	ImageData []byte `json:"image_data"`
	Height    int    `json:"height"`
	Width     int    `json:"width"`
	Quality   int    `json:"quality"`
	Path      string `json:"path"`
	FileName  string `json:"fileName"`
}
