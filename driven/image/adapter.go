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

package image

import (
	"application/core/model"
	"fmt"
	"image"
	"image/png"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/rokwire/core-auth-library-go/v3/authservice"
)

// Adapter implements the Sports interface
type Adapter struct {
	baseURL        string
	accountManager *authservice.ServiceAccountManager
}

// ProcessImages downloads from webtools and uploads in content
func (im Adapter) ProcessImages(item []model.WebToolsEvent) ([]model.WebToolsEvent, error) {
	for _, w := range item {
		if w.LargeImageUploaded != "false" {
			webtoolsImage, _ := im.downloadWebtoolImages(w)
			fmt.Println(webtoolsImage)
		}
	}
	return nil, nil

}
func (im Adapter) downloadWebtoolImages(item model.WebToolsEvent) (*model.ImageData, error) {
	var webtoolImage model.ImageData
	currentAppConfig := "https://calendars.illinois.edu/eventImage"
	currAppConfig := "large.png"
	webtoolImageURL := fmt.Sprintf("%s/%s/%s/%s",
		currentAppConfig,
		item.OriginatingCalendarID,
		item.EventID,
		currAppConfig,
	)

	imageResponse, err := http.Get(webtoolImageURL)
	if err != nil {
		fmt.Println("Error:", err)
		return nil, nil
	}
	defer imageResponse.Body.Close()

	if imageResponse.StatusCode == http.StatusNotFound {
		webtoolImageURL = ""
	}

	if imageResponse.StatusCode == http.StatusOK {

		// Make a GET request to the image URL
		response, err := http.Get(webtoolImageURL)
		if err != nil {
			fmt.Println("Error while downloading the image:", err)
			return nil, nil
		}
		defer response.Body.Close()

		// Decode the image
		img, _, err := image.Decode(response.Body)
		if err != nil {
			fmt.Println("Error while decoding the image:", err)
			return nil, nil
		}

		// Get the image dimensions
		bounds := img.Bounds()
		width := bounds.Dx()
		height := bounds.Dy()

		// Set the filename and quality for the JPEG file
		filename := "image.png"

		// Create a new file to save the image as PNG
		file, err := os.Create(filename)
		if err != nil {
			fmt.Println("Error creating file:", err)
			return nil, nil
		}
		defer file.Close()

		// Encode the image as PNG and save it to the file
		err = png.Encode(file, img)
		if err != nil {
			fmt.Println("Error while saving the image as PNG:", err)
			return nil, nil
		}

		// Download the image and fetch additional data
		// Fetch the image from the URL
		resp, err := http.Get(webtoolImageURL)
		if err != nil {
			fmt.Println("Error fetching image:", err)
			return nil, nil
		}
		defer resp.Body.Close()

		// Read the image data into a byte slice
		imageData, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Error reading image data:", err)
			return nil, nil
		}
		webtoolImage = model.ImageData{ImageData: imageData, Height: height, Width: width,
			Quality: 100, Path: "event/tout", FileName: filename}
	}
	return &webtoolImage, nil
}

// NewImageAdapter creates a new image adapter instance
func NewImageAdapter(imageHost string, accountManager *authservice.ServiceAccountManager) *Adapter {
	return &Adapter{baseURL: imageHost, accountManager: accountManager}
}
