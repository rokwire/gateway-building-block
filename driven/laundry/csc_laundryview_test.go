package laundry

import (
	"log"
	"os"
	"strings"
	"testing"
)

func TestSchoolsCall(t *testing.T) {
	laundryKey := getEnvKey("LAUNDRY_APIKEY", true)
	laundryAPI := getEnvKey("LAUNDRY_APIURL", true)
	luandryServiceKey := getEnvKey("LAUNDRYSERVICE_APIKEY", true)
	laundryServiceAPI := getEnvKey("LAUNDRYSERVICE_API", true)

	laundryAdapter := NewCSCLaundryAdapter(laundryKey, laundryAPI, luandryServiceKey, laundryServiceAPI)
	_, err := laundryAdapter.ListRooms()
	if err != nil {
		t.Fatalf(`test failed`)
	}
}

func TestSchoolsCallInvalidKey(t *testing.T) {
	laundryKey := getEnvKey("LAUNDRY_APIKEY", true)
	laundryAPI := getEnvKey("LAUNDRY_APIURL", true)
	luandryServiceKey := getEnvKey("LAUNDRYSERVICE_APIKEY", true)
	laundryServiceAPI := getEnvKey("LAUNDRYSERVICE_API", true)

	laundryAdapter := NewCSCLaundryAdapter(laundryKey, laundryAPI, luandryServiceKey, laundryServiceAPI)
	_, err := laundryAdapter.ListRooms()
	if err != nil {
		t.Fatalf(`test failed`)
	}
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
	return value
}
