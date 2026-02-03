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
	"context"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/rokwire/rokwire-building-block-sdk-go/services/core/auth"
	"github.com/rokwire/rokwire-building-block-sdk-go/utils/logging/logs"
)

// Adapter implements the Image interface
type Adapter struct {
	baseURL        string
	accountManager *auth.ServiceAccountManager

	logger     logs.Logger
	httpClient *http.Client
}

// ProcessImage process an image
func (im Adapter) ProcessImage(item model.WebToolsEvent) (*model.ContentImagesURL, error) {
	//downlaod
	webtoolsImage, err := im.downloadWebtoolImages(item)
	if err != nil {
		im.logger.Infof("Error with download the webtools image ")
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
		im.logger.Infof("Error with uploading image from content - %s", err)
		return nil, err
	}

	res := model.ContentImagesURL{ID: item.EventID, ImageURL: uploadImageFromContent}

	return &res, nil
}

func (im Adapter) downloadWebtoolImages(item model.WebToolsEvent) (*model.ImageData, error) {

	if item.ImageUploaded != "true" {
		return nil, nil
	}

	// Explicitly skip recurring events
	if item.RecurrenceID != "" && item.RecurrenceID != "0" {
		return nil, nil
	}

	client := im.httpClient
	if client == nil {
		client = &http.Client{Timeout: 6 * time.Second}
	}

	imageURL := fmt.Sprintf(
		"https://calendars.illinois.edu/eventImage/%s/%s/eventImage.png",
		item.OriginatingCalendarID,
		item.EventID,
	)

	log.Printf(
		"[webtools-image] fetch originatingCalendarId=%s eventId=%s url=%s",
		item.OriginatingCalendarID,
		item.EventID,
		imageURL,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, imageURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Printf(
			"[webtools-image] request failed originatingCalendarId=%s eventId=%s err=%v",
			item.OriginatingCalendarID,
			item.EventID,
			err,
		)
		return nil, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf(
			"[webtools-image] image not found originatingCalendarId=%s eventId=%s status=%d",
			item.OriginatingCalendarID,
			item.EventID,
			resp.StatusCode,
		)
		return nil, nil
	}

	if !strings.HasPrefix(resp.Header.Get("Content-Type"), "image/") {
		log.Printf(
			"[webtools-image] non-image response originatingCalendarId=%s eventId=%s content-type=%q",
			item.OriginatingCalendarID,
			item.EventID,
			resp.Header.Get("Content-Type"),
		)
		return nil, nil
	}

	img, _, err := image.Decode(resp.Body)
	if err != nil {
		log.Printf(
			"[webtools-image] decode failed originatingCalendarId=%s eventId=%s err=%v",
			item.OriginatingCalendarID,
			item.EventID,
			err,
		)
		return nil, nil // ✅ skip image, continue batch
	}

	bounds := img.Bounds()

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, err
	}

	log.Printf(
		"[webtools-image] success originatingCalendarId=%s eventId=%s width=%d height=%d",
		item.OriginatingCalendarID,
		item.EventID,
		bounds.Dx(),
		bounds.Dy(),
	)

	return &model.ImageData{
		ImageData: buf.Bytes(),
		Width:     bounds.Dx(),
		Height:    bounds.Dy(),
		Quality:   100,
		Path:      "event/tout",
		FileName:  "image.png",
	}, nil
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
	// Create a buffer to hold the multipart form data
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	// Add form fields: path, width, height and quality
	_ = writer.WriteField("path", path)
	_ = writer.WriteField("width", strconv.Itoa(width))
	_ = writer.WriteField("height", strconv.Itoa(height))
	_ = writer.WriteField("quality", strconv.Itoa(quality))

	// Add the image file to the multipart form
	fileWriter, err := writer.CreateFormFile("fileName", "image.jpg")
	if err != nil {
		return "", fmt.Errorf("error creating form file: %w", err)
	}

	// Copy image data into the multipart file field
	_, err = io.Copy(fileWriter, bytes.NewReader([]byte(filePath)))
	if err != nil {
		return "", fmt.Errorf("error copying file data: %w", err)
	}

	// Close the multipart writer to finalize the request body
	writer.Close()

	// Create a context with timeout to avoid hanging requests
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	// Create the HTTP POST request with the timeout context
	request, err := http.NewRequestWithContext(ctx, "POST", targetURL, &requestBody)
	if err != nil {
		return "", fmt.Errorf("error creating HTTP request: %w", err)
	}

	// Set the correct Content-Type for multipart form data
	request.Header.Set("Content-Type", writer.FormDataContentType())

	// Send the request using the account manager (adds auth, headers, etc.)
	response, err := im.accountManager.MakeRequest(request, "all", "all")
	if err != nil {
		log.Printf("error sending request - %s", err)
		return "", err
	}
	defer response.Body.Close()

	// Validate successful response status
	if response.StatusCode != http.StatusOK {
		log.Printf("error with response code from ContentBB - %d", response.StatusCode)
		return "", fmt.Errorf("error with response code != 200")
	}

	// Read the response body
	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body: %w", err)
	}

	// Return the response as string
	return string(responseBody), nil
}

// NewImageAdapter creates a new image adapter instance
func NewImageAdapter(imageHost string, accountManager *auth.ServiceAccountManager, logger logs.Logger) *Adapter {
	return &Adapter{baseURL: imageHost, accountManager: accountManager, logger: logger, httpClient: &http.Client{
		Timeout: 10 * time.Second,
	}}
}
