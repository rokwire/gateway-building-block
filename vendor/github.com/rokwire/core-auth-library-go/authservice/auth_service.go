// Copyright 2021 Board of Trustees of the University of Illinois.
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

package authservice

import (
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/rokwire/core-auth-library-go/authutils"
	"github.com/rokwire/logging-library-go/logs"
	"golang.org/x/sync/syncmap"
	"gopkg.in/go-playground/validator.v9"
)

// -------------------- AuthService --------------------

// AuthService contains the configurations to interface with the auth service
type AuthService struct {
	dataLoader AuthDataLoader

	// ID of implementing service
	serviceID string

	services        *syncmap.Map
	servicesUpdated *time.Time
	servicesLock    *sync.RWMutex

	minRefreshCacheFreq int
	maxRefreshCacheFreq int
}

// GetServiceID returns the ID of the implementing service
func (a *AuthService) GetServiceID() string {
	return a.serviceID
}

// GetServiceReg returns the service registration record for the given ID if found
func (a *AuthService) GetServiceReg(id string) (*ServiceReg, error) {
	a.servicesLock.RLock()
	servicesUpdated := a.servicesUpdated
	maxRefreshFreq := a.maxRefreshCacheFreq
	a.servicesLock.RUnlock()

	var loadServicesError error
	now := time.Now()
	if servicesUpdated == nil || now.Sub(*servicesUpdated).Minutes() > float64(maxRefreshFreq) {
		loadServicesError = a.LoadServices()
	}

	var service ServiceReg

	if a.services == nil {
		return nil, fmt.Errorf("services could not be loaded: %v", loadServicesError)
	}
	itemValue, ok := a.services.Load(id)
	if !ok {
		return nil, fmt.Errorf("service could not be found for id: %s - %v", id, loadServicesError)
	}

	service, ok = itemValue.(ServiceReg)
	if !ok {
		return nil, fmt.Errorf("service could not be parsed for id: %s - %v", id, loadServicesError)
	}

	return &service, loadServicesError
}

// GetServiceRegWithPubKey returns the service registration record for the given ID if found and validates the PubKey
func (a *AuthService) GetServiceRegWithPubKey(id string) (*ServiceReg, error) {
	serviceReg, err := a.GetServiceReg(id)
	if err != nil || serviceReg == nil {
		return nil, fmt.Errorf("failed to retrieve service reg: %v", err)
	}

	if serviceReg.PubKey == nil {
		return nil, fmt.Errorf("service pub key is nil for id %s", id)
	}

	if serviceReg.PubKey.Key == nil {
		err = serviceReg.PubKey.LoadKeyFromPem()
		if err != nil || serviceReg.PubKey.Key == nil {
			return nil, fmt.Errorf("service pub key is invalid for id %s: %v", id, err)
		}
	}

	return serviceReg, nil
}

// LoadServices loads the subscribed service registration records and caches them
// 	This function will be called periodically after refreshCacheFreq, but can be called directly to force a cache refresh
func (a *AuthService) LoadServices() error {
	services, loadServicesError := a.dataLoader.LoadServices()
	if services != nil {
		a.setServices(services)
	}
	return loadServicesError
}

// SubscribeServices subscribes to the provided services
//	If reload is true and one of the services is not already subscribed, the service registrations will be reloaded immediately
func (a *AuthService) SubscribeServices(serviceIDs []string, reload bool) error {
	newSub := false

	for _, serviceID := range serviceIDs {
		subscribed := a.dataLoader.SubscribeService(serviceID)
		if subscribed {
			newSub = true
		}
	}

	if reload && newSub {
		err := a.LoadServices()
		if err != nil {
			return fmt.Errorf("error loading service registrations: %v", err)
		}
	}

	return nil
}

// UnsubscribeServices unsubscribes from the provided service
func (a *AuthService) UnsubscribeServices(serviceIDs []string) {
	for _, serviceID := range serviceIDs {
		a.dataLoader.UnsubscribeService(serviceID)
	}
}

// ValidateServiceRegistration validates that the implementing service has a valid registration for the provided service ID and hostname
func (a *AuthService) ValidateServiceRegistration(serviceHost string) error {
	service, err := a.GetServiceReg(a.serviceID)
	if err != nil || service == nil {
		return fmt.Errorf("no service registration found with id %s: %v", a.serviceID, err)
	}

	if serviceHost != service.Host {
		return fmt.Errorf("service host (%s) does not match expected value (%s) for id %s", service.Host, serviceHost, a.serviceID)
	}

	return nil
}

// ValidateServiceRegistrationKey validates that the implementing service has a valid registration for the provided keypair
func (a *AuthService) ValidateServiceRegistrationKey(privKey *rsa.PrivateKey) error {
	if privKey == nil {
		return errors.New("provided priv key is nil")
	}

	service, err := a.GetServiceRegWithPubKey(a.serviceID)
	if err != nil {
		return fmt.Errorf("failed to retrieve service pub key: %v", err)
	}

	if service.PubKey.Key.Equal(privKey.PublicKey) {
		return fmt.Errorf("service pub key does not match for id %s", a.serviceID)
	}

	return nil
}

// SetMinRefreshCacheFreq sets the minimum frequency at which cached service registration records are refreshed in minutes
// 	The default value is 1
func (a *AuthService) SetMinRefreshCacheFreq(freq int) {
	a.servicesLock.Lock()
	a.minRefreshCacheFreq = freq
	a.servicesLock.Unlock()
}

// SetMaxRefreshCacheFreq sets the minimum frequency at which cached service registration records are refreshed in minutes
// 	The default value is 60
func (a *AuthService) SetMaxRefreshCacheFreq(freq int) {
	a.servicesLock.Lock()
	a.maxRefreshCacheFreq = freq
	a.servicesLock.Unlock()
}

func (a *AuthService) setServices(services []ServiceReg) {
	a.servicesLock.Lock()

	a.services = &syncmap.Map{}
	if len(services) > 0 {
		for _, service := range services {
			a.services.Store(service.ServiceID, service)
			a.services.Store(service.ServiceAccountID, service)
		}
	}

	time := time.Now()
	a.servicesUpdated = &time

	a.servicesLock.Unlock()
}

// NewAuthService creates and configures a new AuthService instance
func NewAuthService(serviceID string, serviceHost string, dataLoader AuthDataLoader) (*AuthService, error) {
	// Subscribe to the implementing service to validate registration
	dataLoader.SubscribeService(serviceID)

	lock := &sync.RWMutex{}
	services := &syncmap.Map{}

	auth := &AuthService{dataLoader: dataLoader, serviceID: serviceID, services: services, servicesLock: lock,
		minRefreshCacheFreq: 1, maxRefreshCacheFreq: 60}

	err := auth.LoadServices()
	if err != nil {
		return nil, fmt.Errorf("error loading services: %v", err)
	}

	err = auth.ValidateServiceRegistration(serviceHost)
	if err != nil {
		return nil, fmt.Errorf("unable to validate service registration: please contact the auth service system admin to register your service - %v", err)
	}

	return auth, nil
}

func (a *AuthService) CheckForRefresh() (bool, error) {
	a.servicesLock.RLock()
	servicesUpdated := a.servicesUpdated
	minRefreshFreq := a.minRefreshCacheFreq
	a.servicesLock.RUnlock()

	var loadServicesError error
	now := time.Now()
	if servicesUpdated == nil || now.Sub(*servicesUpdated).Minutes() > float64(minRefreshFreq) {
		loadServicesError = a.LoadServices()
		return true, loadServicesError
	}
	return false, loadServicesError
}

// NewTestAuthService creates and configures a new AuthService instance for testing purposes
func NewTestAuthService(serviceID string, serviceHost string, dataLoader AuthDataLoader) (*AuthService, error) {
	// Subscribe to the implementing service to validate registration
	dataLoader.SubscribeService(serviceID)

	lock := &sync.RWMutex{}
	services := &syncmap.Map{}

	auth := &AuthService{dataLoader: dataLoader, serviceID: serviceID, services: services, servicesLock: lock,
		minRefreshCacheFreq: 1, maxRefreshCacheFreq: 60}
	err := auth.LoadServices()
	if err != nil {
		return nil, fmt.Errorf("error loading services: %v", err)
	}

	return auth, nil
}

// -------------------- DataLoader --------------------

// AuthDataLoader declares an interface to load data from an auth service
type AuthDataLoader interface {
	// GetAccessToken gets an access token
	GetAccessToken() error
	// GetDeletedAccounts loads deleted account IDs
	GetDeletedAccounts() ([]string, error)
	ServiceRegLoader
}

//RemoteAuthDataLoaderImpl provides a AuthDataLoader implementation for a remote auth service
type RemoteAuthDataLoaderImpl struct {
	config RemoteAuthDataLoaderConfig

	accessToken AccessToken

	timerDone               chan bool
	getDeletedAccountsTimer *time.Timer

	logger *logs.Logger

	*RemoteServiceRegLoaderImpl
}

//RemoteAuthDataLoaderConfig represents a configuration for a remote data loader
type RemoteAuthDataLoaderConfig struct {
	AuthServicesHost string // URL of auth services host
	ServiceToken     string // Static token issued by the auth service, used to get access tokens from the auth service

	AccessTokenPath     string // Path to auth service access token endpoint
	DeletedAccountsPath string // Path to auth service deleted accounts endpoint
	ServiceRegPath      string // Path to auth service service registration endpoint

	AccessTokenRequest       *http.Request
	DeletedAccountsCallback  func([]string) error // Function to call once the deleted accounts list is received from the auth service
	GetDeletedAccountsPeriod int64                // How often to request deleted account list from the auth service (in hours)
}

// GetAccessToken implements AuthDataLoader interface
func (r *RemoteAuthDataLoaderImpl) GetAccessToken() error {
	client := &http.Client{}
	resp, err := client.Do(r.config.AccessTokenRequest)
	if err != nil {
		return fmt.Errorf("error requesting access token: %v", err)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading body of access token response: %v", err)
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("error getting access token: %d - %s", resp.StatusCode, string(body))
	}

	err = json.Unmarshal(body, &r.accessToken)
	if err != nil {
		return fmt.Errorf("error on unmarshal access token response: %v", err)
	}

	return nil
}

// GetDeletedAccounts implements AuthDataLoader interface
func (r *RemoteAuthDataLoaderImpl) GetDeletedAccounts() ([]string, error) {
	accountIDs, err := r.requestDeletedAccounts()
	if err != nil {
		if strings.HasPrefix(err.Error(), "error getting deleted accounts: 401") {
			// access token may have expired, so get a new one and try once more
			tokenErr := r.GetAccessToken()
			if tokenErr != nil {
				return nil, fmt.Errorf("error getting new access token - %v - after %v", tokenErr, err)
			}

			accountIDs, err = r.requestDeletedAccounts()
			if err != nil {
				return nil, err
			}

			return accountIDs, nil
		}

		return nil, err
	}

	return accountIDs, nil
}

func (r *RemoteAuthDataLoaderImpl) requestDeletedAccounts() ([]string, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", r.config.AuthServicesHost+r.config.DeletedAccountsPath, nil)
	if err != nil {
		return nil, fmt.Errorf("error formatting request to get deleted accounts: %v", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("%s %s", r.accessToken.TokenType, r.accessToken.Token))

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error requesting deleted accounts: %v", err)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading body of deleted accounts response: %v", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("error getting deleted accounts: %d - %s", resp.StatusCode, string(body))
	}

	var deletedAccounts []string
	err = json.Unmarshal(body, &deletedAccounts)
	if err != nil {
		return nil, fmt.Errorf("error on unmarshal deleted accounts response: %v", err)
	}

	return deletedAccounts, nil
}

func (r *RemoteAuthDataLoaderImpl) setupGetDeletedAccountsTimer() {
	//cancel if active
	if r.getDeletedAccountsTimer != nil {
		r.timerDone <- true
		r.getDeletedAccountsTimer.Stop()
	}

	r.getDeletedAccounts(r.config.DeletedAccountsCallback)
}

func (r *RemoteAuthDataLoaderImpl) getDeletedAccounts(callback func([]string) error) {
	accountIDs, err := r.GetDeletedAccounts()
	if err != nil && r.logger != nil {
		r.logger.Error(err.Error())
	}

	err = callback(accountIDs)
	if err != nil && r.logger != nil {
		r.logger.Errorf("Received error from callback function: %v", err)
	}

	duration := time.Hour * time.Duration(r.config.GetDeletedAccountsPeriod)
	r.getDeletedAccountsTimer = time.NewTimer(duration)
	select {
	case <-r.getDeletedAccountsTimer.C:
		// timer expired
		r.getDeletedAccountsTimer = nil

		r.getDeletedAccounts(callback)
	case <-r.timerDone:
		// timer aborted
		r.getDeletedAccountsTimer = nil
	}
}

// NewRemoteAuthDataLoader creates and configures a new NewRemoteAuthDataLoaderImpl instance for the provided auth services url
func NewRemoteAuthDataLoader(config RemoteAuthDataLoaderConfig, subscribedServices []string, logger *logs.Logger) (*RemoteAuthDataLoaderImpl, error) {
	if config.AuthServicesHost == "" {
		return nil, errors.New("auth services host is missing")
	}
	if config.ServiceToken == "" && config.DeletedAccountsCallback != nil {
		return nil, errors.New("service token is missing")
	}

	err := constructDataLoaderConfig(&config)
	if err != nil {
		return nil, fmt.Errorf("error constructing data loader config: %v", err)
	}

	serviceRegLoader := NewRemoteServiceRegLoader(subscribedServices)

	timerDone := make(chan bool)

	dataLoader := RemoteAuthDataLoaderImpl{config: config, timerDone: timerDone, logger: logger, RemoteServiceRegLoaderImpl: serviceRegLoader}
	serviceRegLoader.dataLoader = &dataLoader

	if config.DeletedAccountsCallback != nil {
		dataLoader.setupGetDeletedAccountsTimer()
	}

	return &dataLoader, nil
}

func constructDataLoaderConfig(config *RemoteAuthDataLoaderConfig) error {
	if config.AccessTokenPath == "" {
		config.AccessTokenPath = "/bbs/access-token"
	}
	if config.DeletedAccountsPath == "" {
		config.DeletedAccountsPath = "/bbs/deleted-accounts"
	}
	if config.ServiceRegPath == "" {
		config.ServiceRegPath = "/bbs/service-regs"
	}

	if config.AccessTokenRequest == nil {
		r, err := authutils.GetDefaultAccessTokenRequest(config.AuthServicesHost, config.AccessTokenPath, config.ServiceToken)
		if err != nil {
			return err
		}

		config.AccessTokenRequest = r
	}
	if config.GetDeletedAccountsPeriod <= 0 {
		config.GetDeletedAccountsPeriod = 2
	}

	return nil
}

// AccessToken represents an access token granted by a remote auth service
type AccessToken struct {
	Token     string `json:"access_token"`
	TokenType string `json:"token_type"`
}

// -------------------- ServiceRegLoader --------------------

// ServiceRegLoader declares an interface to load the service registrations for specified services
type ServiceRegLoader interface {
	// LoadServices loads the service registration records for all subscribed services
	LoadServices() ([]ServiceReg, error)
	//GetSubscribedServices returns the list of currently subscribed services
	GetSubscribedServices() []string
	// SubscribeService subscribes the loader to the given service
	// 	Returns true if the specified service was added or false if it was already found
	SubscribeService(serviceID string) bool
	// UnsubscribeService unsubscribes the loader from the given service
	// 	Returns true if the specified service was removed or false if it was not found
	UnsubscribeService(serviceID string) bool
}

//RemoteServiceRegLoaderImpl provides a ServiceRegLoader implementation for a remote auth service
type RemoteServiceRegLoaderImpl struct {
	dataLoader *RemoteAuthDataLoaderImpl

	*ServiceRegSubscriptions
}

// LoadServices implements ServiceRegLoader interface
func (r *RemoteServiceRegLoaderImpl) LoadServices() ([]ServiceReg, error) {
	if len(r.GetSubscribedServices()) == 0 {
		return nil, nil
	}

	client := &http.Client{}
	req, err := http.NewRequest("GET", r.dataLoader.config.AuthServicesHost+r.dataLoader.config.ServiceRegPath, nil)
	if err != nil {
		return nil, fmt.Errorf("error formatting request to load services: %v", err)
	}

	servicesQuery := strings.Join(r.GetSubscribedServices(), ",")

	q := req.URL.Query()
	q.Add("ids", servicesQuery)
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error requesting services: %v", err)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading body of service response: %v", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("error loading services: %d - %s", resp.StatusCode, string(body))
	}

	var services []ServiceReg
	err = json.Unmarshal(body, &services)
	if err != nil {
		return nil, fmt.Errorf("error on unmarshal service response: %v", err)
	}

	validate := validator.New()
	for _, service := range services {
		err = validate.Struct(service)
		if err != nil {
			return nil, fmt.Errorf("error validating service data: %v", err)
		}
		service.PubKey.LoadKeyFromPem()
	}

	return services, nil
}

// NewRemoteServiceRegLoader creates and configures a new RemoteServiceRegLoaderImpl instance for the provided auth services url
func NewRemoteServiceRegLoader(subscribedServices []string) *RemoteServiceRegLoaderImpl {
	subscriptions := NewServiceRegSubscriptions(subscribedServices)
	return &RemoteServiceRegLoaderImpl{ServiceRegSubscriptions: subscriptions}
}

// -------------------- ServiceRegSubscriptions --------------------

// ServiceRegSubscriptions defined a struct to hold service registration subscriptions
// 	This struct implements the subcription part of the ServiceRegLoader interface
//	If you subscribe to the reserved "all" service ID, all registered services
//	will be loaded
type ServiceRegSubscriptions struct {
	subscribedServices []string // Service registrations to load
	servicesLock       *sync.RWMutex
}

// GetSubscribedServices returns the list of subscribed services
func (r *ServiceRegSubscriptions) GetSubscribedServices() []string {
	r.servicesLock.RLock()
	defer r.servicesLock.RUnlock()

	return r.subscribedServices
}

// SubscribeService adds the given service ID to the list of subscribed services if not already present
// 	Returns true if the specified service was added or false if it was already found
func (r *ServiceRegSubscriptions) SubscribeService(serviceID string) bool {
	r.servicesLock.Lock()
	defer r.servicesLock.Unlock()

	if !authutils.ContainsString(r.subscribedServices, serviceID) {
		r.subscribedServices = append(r.subscribedServices, serviceID)
		return true
	}

	return false
}

// UnsubscribeService removed the given service ID from the list of subscribed services if presents
// 	Returns true if the specified service was removed or false if it was not found
func (r *ServiceRegSubscriptions) UnsubscribeService(serviceID string) bool {
	r.servicesLock.Lock()
	defer r.servicesLock.Unlock()

	services, removed := authutils.RemoveString(r.subscribedServices, serviceID)
	r.subscribedServices = services

	return removed
}

// NewServiceRegSubscriptions creates and configures a new ServiceRegSubscriptions instance
func NewServiceRegSubscriptions(subscribedServices []string) *ServiceRegSubscriptions {
	lock := &sync.RWMutex{}
	return &ServiceRegSubscriptions{subscribedServices: subscribedServices, servicesLock: lock}
}

// -------------------- ServiceReg --------------------

// ServiceReg represents a service registration record
type ServiceReg struct {
	ServiceID        string  `json:"service_id" bson:"service_id" validate:"required"`
	ServiceAccountID string  `json:"service_account_id" bson:"service_account_id"`
	Host             string  `json:"host" bson:"host" validate:"required"`
	PubKey           *PubKey `json:"pub_key" bson:"pub_key"`
}

// -------------------- PubKey --------------------

// PubKey represents a public key object including the key and related metadata
type PubKey struct {
	Key    *rsa.PublicKey `json:"-" bson:"-"`
	KeyPem string         `json:"key_pem" bson:"key_pem" validate:"required"`
	Alg    string         `json:"alg" bson:"alg" validate:"required"`
	Kid    string         `json:"-" bson:"-"`
}

// LoadKeyFromPem parses "KeyPem" and sets the "Key" and "Kid"
func (p *PubKey) LoadKeyFromPem() error {
	if p == nil {
		return fmt.Errorf("pubkey is nil")
	}

	key, err := jwt.ParseRSAPublicKeyFromPEM([]byte(p.KeyPem))
	if err != nil {
		p.Key = nil
		p.Kid = ""
		return fmt.Errorf("error parsing key string: %v", err)
	}

	kid, err := authutils.GetKeyFingerprint(key)
	if err != nil {
		p.Key = nil
		p.Kid = ""
		return fmt.Errorf("error getting key fingerprint: %v", err)
	}

	p.Key = key
	p.Kid = kid

	return nil
}
