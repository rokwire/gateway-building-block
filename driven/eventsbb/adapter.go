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

package eventsbb

import (
	"application/core/model"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/rokwire/logging-library-go/v2/logs"
)

// Adapter implements the Storage interface
type Adapter struct {
	baseURL string
	apiKey  string
}

// NewEventsBBAdapter creates new instance
func NewEventsBBAdapter(legacyEventsBaseURL, legacyEventsAPIKey string) *Adapter {
	return &Adapter{
		baseURL: legacyEventsBaseURL,
		apiKey:  legacyEventsAPIKey, // pragma: allowlist secret
	}
}

// LoadAllLegacyEvents sends notification to a user
func (na *Adapter) LoadAllLegacyEvents(log *logs.Log) ([]model.LegacyEvent, error) {

	url := fmt.Sprintf("%s/events", na.baseURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Errorf("legacy_events.LoadAllLegacyEvents: error creating load legacy events request - %s", err)
		return nil, err
	}
	req.Header.Set("ROKWIRE-API-KEY", na.apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Errorf("legacy_events.LoadAllLegacyEvents: error creating load legacy events request - %s", err)
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		errorResponse, _ := io.ReadAll(resp.Body)
		if errorResponse != nil {
			log.Errorf("legacy_events.LoadAllLegacyEvents: error with response code - %s", errorResponse)
		}
		log.Errorf("legacy_events.LoadAllLegacyEvents: error with response code - %d", resp.StatusCode)
		return nil, fmt.Errorf("SendNotification:error with response code != 200")
	}
	var list []model.LegacyEvent
	err = json.NewDecoder(resp.Body).Decode(&list)
	if err != nil {
		log.Errorf("legacy_events.LoadAllLegacyEvents: error with response code - %d", resp.StatusCode)
		return nil, fmt.Errorf("SendNotification: %s", err)
	}
	return list, nil

}
