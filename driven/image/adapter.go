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
	"bytes"
	"fmt"
	"image"
	"image/png"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"

	"github.com/rokwire/core-auth-library-go/v3/authservice"
)

// Adapter implements the Sports interface
type Adapter struct {
	baseURL        string
	accountManager *authservice.ServiceAccountManager
}

// ProcessImages downloads from webtools and uploads in content
func (im Adapter) ProcessImages(item []model.WebToolsEvent) ([]model.ContentImagesURL, error) {
	var contentimageUrl []model.ContentImagesURL
	for _, w := range item {
		if w.LargeImageUploaded != "false" {
			webtoolsImage, err := im.downloadWebtoolImages(w)
			if err != nil {
				return nil, err
			}
			uploadImageFromContent, _ := im.uploadImageFromContent(webtoolsImage.ImageData,
				webtoolsImage.Height, webtoolsImage.Width, webtoolsImage.Quality,
				webtoolsImage.Path, webtoolsImage.FileName)
			contentImage := model.ContentImagesURL{ID: w.EventID, ImageURL: uploadImageFromContent}
			contentimageUrl = append(contentimageUrl, contentImage)
			fmt.Println(contentimageUrl)
		}
	}
	return contentimageUrl, nil

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

// Function to upload image to another API along with additional data
func (im Adapter) uploadImageFromContent(imageData []byte, height int, width int, quality int, path, fileName string) (string, error) {
	// URL to which the request will be sent
	targetURL := "http://localhost/content/bbs/image"

	// Send the request and get the response
	response, err := sendRequest(targetURL, path, width, height, quality, string(imageData))
	if err != nil {
		fmt.Println("Error:", err)
		return "", err
	}

	fmt.Println("Response:", response)
	return response, nil
}

func sendRequest(targetURL, path string, width, height, quality int, filePath string) (string, error) {
	// Create a new buffer to store the multipart form data
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	// Add the path, width, height, and quality as form fields
	_ = writer.WriteField("path", path)
	_ = writer.WriteField("width", strconv.Itoa(width))
	_ = writer.WriteField("height", strconv.Itoa(height))
	_ = writer.WriteField("quality", strconv.Itoa(quality))

	// Add the file as a form file field
	fileWriter, err := writer.CreateFormFile("fileName", "image.jpg")
	if err != nil {
		return "", fmt.Errorf("error creating form file: %w", err)
	}

	// Copy the file data into the file writer
	_, err = io.Copy(fileWriter, bytes.NewReader([]byte(filePath)))
	if err != nil {
		return "", fmt.Errorf("error copying file data: %w", err)
	}

	// Close the multipart writer
	writer.Close()

	// Create the HTTP request
	request, err := http.NewRequest("POST", targetURL, &requestBody)
	if err != nil {
		return "", fmt.Errorf("error creating HTTP request: %w", err)
	}

	// Set the content type header
	request.Header.Set("Content-Type", writer.FormDataContentType())

	// Send the request
	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return "", fmt.Errorf("error sending HTTP request: %w", err)
	}
	defer response.Body.Close()

	// Read the response body
	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body: %w", err)
	}

	// Convert response body to string
	responseString := string(responseBody)

	return responseString, nil
}

// NewImageAdapter creates a new image adapter instance
func NewImageAdapter(imageHost string, accountManager *authservice.ServiceAccountManager) *Adapter {
	return &Adapter{baseURL: imageHost, accountManager: accountManager}
}
