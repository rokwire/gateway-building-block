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

package main

import (
	"apigateway/core"
	model "apigateway/core/model"
	contactinfo "apigateway/driven/contactinfo"
	courses "apigateway/driven/courses"
	"apigateway/driven/laundry"
	location "apigateway/driven/location"
	storage "apigateway/driven/storage"
	terms "apigateway/driven/terms"
	driver "apigateway/driver/web"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

var (
	// Version : version of this executable
	Version string
	// Build : build date of this executable
	Build string
)

/*
func printDeletedAccountIDs(accountIDs []string) error {
	log.Printf("Deleted account IDs: %v\n", accountIDs)
	return nil
}
*/

func main() {
	if len(Version) == 0 {
		Version = "dev"
	}

	port := getEnvKey("GATEWAY_PORT", true)

	//mongoDB adapter
	mongoDBAuth := getEnvKey("GATEWAY_MONGO_AUTH", true)
	mongoDBName := getEnvKey("GATEWAY_MONGO_DATABASE", true)
	mongoTimeout := getEnvKey("GATEWAY_MONGO_TIMEOUT", false)
	laundryKey := getEnvKey("GATEWAY_LAUNDRY_APIKEY", true)
	laundryAPI := getEnvKey("GATEWAY_LAUNDRY_APIURL", true)
	luandryServiceKey := getEnvKey("GATEWAY_LAUNDRYSERVICE_APIKEY", true)
	laundryServiceAPI := getEnvKey("GATEWAY_LAUNDRYSERVICE_API", true)
	laundryServiceToken := getEnvKey("GATEWAY_LAUNDRYSERVICE_BASICAUTH", true)
	wayfindingURL := getEnvKey("GATEWAY_WAYFINDING_APIURL", true)
	wayfindingKey := getEnvKey("GATEWAY_WAYFINDING_APIKEY", true)
	// pragma: allowlist nextline secret
	campusInfoAPIKey := getEnvKey("GATEWAY_CENTRALCAMPUS_APIKEY", true)
	campusAITSEndPoint := getEnvKey("GATEWAY_CENTRALCAMPUS_ENDPOINT", true)
	coursesEndPoint := getEnvKey("GATEWAY_GIESCOURSES_ENDPOINT", true)

	//read assets
	file, _ := ioutil.ReadFile("./assets/assets.json")
	assets := model.Asset{}
	_ = json.Unmarshal([]byte(file), &assets)
	laundryAssets := make(map[string]model.LaundryDetails)

	for i := 0; i < len(assets.Laundry.Assets); i++ {
		laundryAsset := assets.Laundry.Assets[i]
		laundryAssets[laundryAsset.LocationID] = laundryAsset.Details
	}

	storageAdapter := storage.NewStorageAdapter(mongoDBAuth, mongoDBName, mongoTimeout)
	laundryAdapter := laundry.NewCSCLaundryAdapter(laundryKey, laundryAPI, luandryServiceKey, laundryServiceAPI, laundryAssets, laundryServiceToken)
	locationAdapter := location.NewUIUCWayFinding(wayfindingKey, wayfindingURL)
	contactAdapter := contactinfo.NewContactAdapter(campusInfoAPIKey, campusAITSEndPoint)
	giescourseAdapter := courses.NewGiesCourseAdapter(coursesEndPoint)
	studentCourseAdapter := courses.NewCourseAdapter(campusAITSEndPoint, campusInfoAPIKey)
	termSessionAdapter := terms.NewTermSessionAdapter()

	err := storageAdapter.Start()
	if err != nil {
		log.Fatal("Cannot start the mongoDB adapter - " + err.Error())
	}

	log.Printf("MongoDB Started")

	//application
	application := core.NewApplication(Version, Build, storageAdapter, laundryAdapter, locationAdapter, contactAdapter, giescourseAdapter, studentCourseAdapter, termSessionAdapter)
	application.Start()

	//web adapter
	host := getEnvKey("GATEWAY_HOST", true)
	corehost := getEnvKey("GATEWAY_CORE_HOST", true)
	log.Printf(corehost)

	tokenAuth := driver.NewTokenAuth(host, corehost)
	fmt.Println("auth setup complete")

	log.Printf("Creating web adapter")
	webAdapter := driver.NewWebAdapter(host, port, application, tokenAuth)

	log.Printf("starting web adapter")
	webAdapter.Start()
}

func getEnvKey(key string, required bool) string {
	//get from the environment
	value, exist := os.LookupEnv(key)
	if !exist {
		if required {
			log.Fatal("No provided environment variable for " + key)
		} else {
			log.Printf("No provided environment variable for " + key)
		}
	}
	printEnvVar(key, value)
	return value
}

func printEnvVar(name string, value string) {
	if Version == "dev" {
		log.Printf("%s=%s", name, value)
	}
}
