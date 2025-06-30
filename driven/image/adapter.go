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
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"strconv"

	"github.com/rokwire/rokwire-building-block-sdk-go/services/core/auth"
	"github.com/rokwire/rokwire-building-block-sdk-go/utils/logging/logs"
)

// Adapter implements the Image interface
type Adapter struct {
	baseURL        string
	accountManager *auth.ServiceAccountManager

	logger logs.Logger
}

// ProcessImage process an image
func (im Adapter) ProcessImage(item model.WebToolsEvent) (*model.ContentImagesURL, error) {
	//downlaod
	webtoolsImage, err := im.downloadWebtoolImages(item)
	if err != nil {
		return nil, err
	}

	if webtoolsImage == nil {
		im.logger.Infof("no webtools image for %s", item.EventID)
		return nil, nil
	}

	//upload
	uploadImageFromContent, err := im.uploadImageFromContent(webtoolsImage.ImageData,
		webtoolsImage.Height, webtoolsImage.Width, webtoolsImage.Quality,
		webtoolsImage.Path, webtoolsImage.FileName)
	if err != nil {
		return nil, err
	}

	res := model.ContentImagesURL{ID: item.EventID, ImageURL: uploadImageFromContent}

	return &res, nil
}

// Why do you call this API two times??
func (im Adapter) downloadWebtoolImages(item model.WebToolsEvent) (*model.ImageData, error) {
	currentAppConfig := "https://calendars.illinois.edu/eventImage"
	currAppConfig := "large.png"
	webtoolImageURL := fmt.Sprintf("%s/%s/%s/%s",
		currentAppConfig,
		item.OriginatingCalendarID,
		item.EventID,
		currAppConfig,
	)

	// Make a GET request to download the image
	response, err := http.Get(webtoolImageURL)
	if err != nil {
		fmt.Println("Error while downloading the image:", err)
		return nil, err
	}
	defer response.Body.Close()

	// Check if the response status code is OK
	if response.StatusCode != http.StatusOK {
		im.logger.Infof("response code %d for %s", response.StatusCode, item.EventID)
		return nil, nil //do not return error when cannot get/find an image
	}

	// Decode the image
	img, _, err := image.Decode(response.Body)
	if err != nil {
		fmt.Println("Error while decoding the image:", err)
		return nil, err
	}

	// Get the image dimensions
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Encode the image as PNG
	var buf bytes.Buffer
	err = png.Encode(&buf, img)
	if err != nil {
		fmt.Println("Error while encoding the image as PNG:", err)
		return nil, err
	}

	// Fetch additional data and return
	webtoolImage := model.ImageData{
		ImageData: buf.Bytes(),
		Height:    height,
		Width:     width,
		Quality:   100,
		Path:      "event/tout",
		FileName:  "image.png",
	}

	return &webtoolImage, nil
}

// Function to upload image to another API along with additional data
func (im Adapter) uploadImageFromContent(imageData []byte, height int, width int, quality int, path, fileName string) (string, error) {
	// URL to which the request will be sent
	targetURL := fmt.Sprintf("%s/content/bbs/image", im.baseURL)

	// Send the request and get the response
	respData, err := im.sendRequest(targetURL, path, width, height, quality, string(imageData))
	if err != nil {
		fmt.Println("Error:", err)
		return "", err
	}

	type response struct {
		URL string `json:"url"`
	}

	var resp response
	err = json.Unmarshal([]byte(respData), &resp)
	if err != nil {
		fmt.Println("Error unmarshalling response:", err)
		return "", err
	}

	return resp.URL, nil
}

func (im Adapter) sendRequest(targetURL, path string, width, height, quality int, filePath string) (string, error) {
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
	response, err := im.accountManager.MakeRequest(request, "all", "all")
	if err != nil {
		log.Printf("error sending request - %s", err)
		return "", err
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		log.Printf("error with response code - %d", response.StatusCode)
		return "", fmt.Errorf("error with response code != 200")
	}

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
func NewImageAdapter(imageHost string, accountManager *auth.ServiceAccountManager, logger logs.Logger) *Adapter {
	return &Adapter{baseURL: imageHost, accountManager: accountManager, logger: logger}
}
