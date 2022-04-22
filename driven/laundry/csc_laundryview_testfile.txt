package laundry

import (
	"log"
	"os"
	"testing"
)

func TestSchoolsCall(t *testing.T) {
	laundryKey := getEnvKey("GATEWAY_LAUNDRY_APIKEY", true)
	laundryAPI := getEnvKey("GATEWAY_LAUNDRY_APIURL", true)
	luandryServiceKey := getEnvKey("GATEWAY_LAUNDRYSERVICE_APIKEY", true)
	laundryServiceAPI := getEnvKey("GATEWAY_LAUNDRYSERVICE_API", true)

	laundryAdapter := NewCSCLaundryAdapter(laundryKey, laundryAPI, luandryServiceKey, laundryServiceAPI)
	_, err := laundryAdapter.ListRooms()
	if err != nil {
		t.Fatalf(`test failed`)
	}
}

func TestSchoolsCallInvalidKey(t *testing.T) {
	laundryKey := getEnvKey("GATEWAY_LAUNDRY_APIKEY", true)
	laundryAPI := getEnvKey("GATEWAY_LAUNDRY_APIURL", true)
	luandryServiceKey := getEnvKey("GATEWAY_LAUNDRYSERVICE_APIKEY", true)
	laundryServiceAPI := getEnvKey("GATEWAY_LAUNDRYSERVICE_API", true)

	laundryAdapter := NewCSCLaundryAdapter(laundryKey, laundryAPI, luandryServiceKey, laundryServiceAPI)
	_, err := laundryAdapter.ListRooms()
	if err != nil {
		t.Fatalf(`test failed`)
	}
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
