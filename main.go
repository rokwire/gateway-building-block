/*
 *   Copyright (c) 2020 Board of Trustees of the University of Illinois.
 *   All rights reserved.

 *   Licensed under the Apache License, Version 2.0 (the "License");
 *   you may not use this file except in compliance with the License.
 *   You may obtain a copy of the License at

 *   http://www.apache.org/licenses/LICENSE-2.0

 *   Unless required by applicable law or agreed to in writing, software
 *   distributed under the License is distributed on an "AS IS" BASIS,
 *   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *   See the License for the specific language governing permissions and
 *   limitations under the License.
 */

package main

import (
	"apigateway/core"
	"apigateway/driven/laundry"
	storage "apigateway/driven/storage"
	driver "apigateway/driver/web"
	"fmt"
	"log"
	"os"
	"strings"
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

	port := getEnvKey("PORT", true)

	//mongoDB adapter
	mongoDBAuth := getEnvKey("MONGO_AUTH", true)
	mongoDBName := getEnvKey("MONGO_DATABASE", true)
	mongoTimeout := getEnvKey("MONGO_TIMEOUT", false)
	laundryKey := getEnvKey("LAUNDRY_APIKEY", true)
	laundryAPI := getEnvKey("LAUNDRY_APIURL", true)
	luandryServiceKey := getEnvKey("LAUNDRYSERVICE_APIKEY", true)
	laundryServiceAPI := getEnvKey("LAUNDRYSERVICE_API", true)
	storageAdapter := storage.NewStorageAdapter(mongoDBAuth, mongoDBName, mongoTimeout)
	laundryAdapter := laundry.NewCSCLaundryAdapter(laundryKey, laundryAPI, luandryServiceKey, laundryServiceAPI)

	err := storageAdapter.Start()
	if err != nil {
		log.Fatal("Cannot start the mongoDB adapter - " + err.Error())
	}

	log.Printf("MongoDB Started")
	//application
	application := core.NewApplication(Version, Build, storageAdapter, laundryAdapter)
	application.Start()

	//web adapter
	host := getEnvKey("HOST", true)
	corehost := getEnvKey("CORE_HOST", true)
	log.Printf(corehost)
	log.Printf("Creating web adapter")
	/*
		serviceID := "laundry"
		config := authservice.RemoteAuthDataLoaderConfig{
			AuthServicesHost: corehost,
			ServiceToken:     serviceToken,

			DeletedAccountsCallback: printDeletedAccountIDs,
		}
		logger := logs.NewLogger(serviceID, nil)
		dataLoader, err := authservice.NewRemoteAuthDataLoader(config, nil, logger)
		if err != nil {
			log.Fatalf("Error initializing remote data loader: %v", err)
		}

		authservice, err := authservice.NewAuthService(serviceID, host, dataLoader)
		if err != nil {
			log.Fatalf("Error initializing auth service: %v", err)
		}

		permissionAuth := authorization.NewCasbinStringAuthorization("./permissions_authorization_policy.csv")
		scopeAuth := authorization.NewCasbinScopeAuthorization("./scope_authorization_policy.csv", serviceID)

		tokenAuth, err := tokenauth.NewTokenAuth(true, authservice, permissionAuth, scopeAuth)
		if err != nil {
			log.Fatalf("Error initializing toekan auth: %v", err)
		}

	*/
	tokenAuth := driver.NewTokenAuth(host, corehost)
	fmt.Println("setup complete")

	webAdapter := driver.NewWebAdapter(host, port, application, tokenAuth)

	log.Printf("starting web adapter")
	webAdapter.Start()
}

func getAPIKeys() []string {
	//get from the environment
	rokwireAPIKeys := getEnvKey("ROKWIRE_API_KEYS", true)

	//it is comma separated format
	rokwireAPIKeysList := strings.Split(rokwireAPIKeys, ",")
	if len(rokwireAPIKeysList) <= 0 {
		log.Fatal("For some reasons the apis keys list is empty")
	}

	return rokwireAPIKeysList
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
