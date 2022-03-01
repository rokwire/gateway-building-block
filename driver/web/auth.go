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

package web

import (
	"apigateway/core"
	"apigateway/core/model"
	"context"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/lestrrat-go/jwx/jwk"
	"golang.org/x/sync/syncmap"
	"gopkg.in/ericchiang/go-oidc.v2"
)

type cacheUser struct {
	lastUsage time.Time
}

//Auth handler
type Auth struct {
	apiKeysAuth *APIKeysAuth
	userAuth    *UserAuth
	adminAuth   *AdminAuth
}

//Start starts the auth module
func (auth *Auth) Start() error {
	auth.adminAuth.start()
	auth.userAuth.start()

	return nil
}

func (auth *Auth) clientIDCheck(w http.ResponseWriter, r *http.Request) bool {
	clientID := r.Header.Get("APP")
	if len(clientID) == 0 {
		clientID = "edu.illinois.rokwire"
	}

	log.Println(fmt.Sprintf("400 - Bad Request"))
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte("Bad Request"))
	return false
}

func (auth *Auth) apiKeyCheck(w http.ResponseWriter, r *http.Request) bool {
	return auth.apiKeysAuth.check(w, r)
}

func (auth *Auth) userCheck(w http.ResponseWriter, r *http.Request) (bool, *model.ShibbolethUser, *string) {
	return auth.userAuth.userCheck(w, r)
}

//NewAuth creates new auth handler
func NewAuth(app *core.Application, appKeys []string, oidcProvider string,
	oidcAppClientID string, appClientID string, webAppClientID string, phoneAuthSecret string,
	authKeys string, authIssuer string) *Auth {
	log.Printf("creating api keys")
	apiKeysAuth := newAPIKeysAuth(appKeys)
	//log.Printf("creating user auth")
	//userAuth2 := newUserAuth(app, oidcProvider, oidcAppClientID, phoneAuthSecret, authKeys, authIssuer)
	//log.Printf("creating admin auth")
	//adminAuth := newAdminAuth(app, oidcProvider, appClientID, webAppClientID)

	//auth := Auth{apiKeysAuth: apiKeysAuth, userAuth: userAuth2, adminAuth: adminAuth}
	auth := Auth{apiKeysAuth: apiKeysAuth}
	return &auth
}

/////////////////////////////////////

//APIKeysAuth entity
type APIKeysAuth struct {
	appKeys []string
}

func (auth *APIKeysAuth) check(w http.ResponseWriter, r *http.Request) bool {
	apiKey := r.Header.Get("ROKWIRE-API-KEY")
	if len(apiKey) == 0 {
		log.Println(fmt.Sprintf("401 - Missing API key %s", apiKey))
		return false
	}

	appKeys := auth.appKeys
	exist := false
	for _, element := range appKeys {
		if element == apiKey {
			exist = true
			break
		}
	}
	if !exist {
		log.Println(fmt.Sprintf("401 - Unauthorized for key %s", apiKey))
		return false
	}
	return true
}

func newAPIKeysAuth(appKeys []string) *APIKeysAuth {
	auth := APIKeysAuth{appKeys}
	return &auth
}

////////////////////////////////////

type userData struct {
	UIuceduUIN        *string   `json:"uiucedu_uin"`
	Sub               *string   `json:"sub"`
	Email             *string   `json:"email"`
	UIuceduIsMemberOf *[]string `json:"uiucedu_is_member_of"`
}

//AdminAuth entity
type AdminAuth struct {
	app *core.Application

	appVerifier    *oidc.IDTokenVerifier
	appClientID    string
	webAppVerifier *oidc.IDTokenVerifier
	webAppClientID string

	cachedUsers     *syncmap.Map //cache users while active - 5 minutes timeout
	cachedUsersLock *sync.RWMutex
}

func (auth *AdminAuth) start() {

}

//gets the token from the request - as cookie or as Authorization header.
//returns the id token and its type - mobile or web. If the token is taken by the cookie it is web otherwise it is mobile
func (auth *AdminAuth) getIDToken(r *http.Request) (*string, *string, error) {
	var tokenType string

	//1. Check if there is a cookie
	cookie, err := r.Cookie("rwa-at-data")
	if err == nil && cookie != nil && len(cookie.Value) > 0 {
		//there is a cookie
		tokenType = "web"
		return &cookie.Value, &tokenType, nil
	}

	//2. Check if there is a token in the Authorization header
	authorizationHeader := r.Header.Get("Authorization")
	if len(authorizationHeader) <= 0 {
		return nil, nil, errors.New("error getting Authorization header")
	}
	splitAuthorization := strings.Fields(authorizationHeader)
	if len(splitAuthorization) != 2 {
		return nil, nil, errors.New("error processing the Authorization header")
	}
	// expected - Bearer 1234
	if splitAuthorization[0] != "Bearer" {
		return nil, nil, errors.New("error processing the Authorization header")
	}
	rawIDToken := splitAuthorization[1]
	tokenType = "mobile"
	return &rawIDToken, &tokenType, nil
}

func (auth *AdminAuth) verify(rawIDToken string, tokenType string) (*oidc.IDToken, error) {
	switch tokenType {
	case "mobile":
		log.Println("AdminAuth -> mobile app client token")
		return auth.appVerifier.Verify(context.Background(), rawIDToken)
	case "web":
		log.Println("AdminAuth -> web app client token")
		return auth.webAppVerifier.Verify(context.Background(), rawIDToken)
	default:
		return nil, errors.New("AdminAuth -> there is an issue with the audience")
	}
}

func (auth *AdminAuth) responseBadRequest(w http.ResponseWriter) {
	log.Println("AdminAuth -> 400 - Bad Request")

	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte("Bad Request"))
}

func (auth *AdminAuth) responseUnauthorized(token string, w http.ResponseWriter) {
	log.Printf("AdminAuth -> 401 - Unauthorized for token %s", token)

	w.WriteHeader(http.StatusUnauthorized)
	w.Write([]byte("Unauthorized"))
}

func (auth *AdminAuth) responseForbbiden(info string, w http.ResponseWriter) {
	log.Printf("AdminAuth -> 403 - Forbidden - %s", info)

	w.WriteHeader(http.StatusForbidden)
	w.Write([]byte("Forbidden"))
}

func (auth *AdminAuth) responseInternalServerError(w http.ResponseWriter) {
	log.Println("AdminAuth -> 500 - Internal Server Error")

	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte("Internal Server Error"))
}

func newAdminAuth(app *core.Application, oidcProvider string, appClientID string, webAppClientID string) *AdminAuth {
	provider, err := oidc.NewProvider(context.Background(), oidcProvider)
	if err != nil {
		log.Fatalln(err)
	}

	appVerifier := provider.Verifier(&oidc.Config{ClientID: appClientID})
	webAppVerifier := provider.Verifier(&oidc.Config{ClientID: webAppClientID})

	cacheUsers := &syncmap.Map{}
	lock := &sync.RWMutex{}

	auth := AdminAuth{app: app, appVerifier: appVerifier, appClientID: appClientID,
		webAppVerifier: webAppVerifier, webAppClientID: webAppClientID,
		cachedUsers: cacheUsers, cachedUsersLock: lock}
	return &auth
}

/////////////////////////////////////

type tokenData struct {
	UID      string
	Name     string
	Email    string
	Phone    string
	ClientID string
	Groups   string
	Auth     string
	Type     string
	ISS      string
}

//UserAuth entity
type UserAuth struct {
	app *core.Application

	//shibboleth - keep for back compatability
	appIDTokenVerifier *oidc.IDTokenVerifier

	//phone - keep for back compatability
	phoneAuthSecret string

	//auth service
	Keys   jwk.Set
	Issuer string

	cachedUsers     *syncmap.Map //cache users while active - 5 minutes timeout
	cachedUsersLock *sync.RWMutex

	rosters     []map[string]string //cache rosters
	rostersLock *sync.RWMutex
}

func (auth *UserAuth) start() {

}

func (auth *UserAuth) mainCheck(r *http.Request) (bool, *model.ShibbolethUser, *string) {
	//get the tokens
	token, tokenSourceType, csrfToken, err := auth.getTokens(r)
	if err != nil {
		log.Printf("error gettings tokens - %s", err)
		return false, nil, nil
	}

	//check if all input data is available
	if token == nil || len(*token) == 0 {
		return false, nil, nil
	}
	rawToken := *token //we have token
	if *tokenSourceType == "cookie" && (csrfToken == nil || len(*csrfToken) == 0) {
		//if the token is sent via cookie then we must have csrf token as well
		return false, nil, nil
	}

	// determine the token type: 1 for shibboleth, 2 for phone, 3 for auth access token
	// 1 & 2 are deprecated but we support them for back compatability
	tokenType, err := auth.getTokenType(rawToken)
	if err != nil {
		return false, nil, nil
	}
	if !(*tokenType == 1 || *tokenType == 2 || *tokenType == 3) {
		return false, nil, nil
	}

	// process the token - validate it, extract the user identifier
	var externalID string
	var authType string
	user := &model.ShibbolethUser{}

	switch *tokenType {
	case 1:
		//support this for back compatability
		shibboData, err := auth.processShibbolethToken(rawToken)
		if err != nil {
			return false, nil, nil
		}
		user.Uin = shibboData.UIuceduUIN
		user.Email = shibboData.Email
		user.Membership = shibboData.UIuceduIsMemberOf
		authType = "shibboleth"
	case 2:
		//support this for back compatability
		phone, err := auth.processPhoneToken(rawToken)
		if err != nil {
			return false, nil, nil
		}
		user.Phone = phone
		authType = "phone"
	case 3:
		//mobile app sends just token, the browser sends token + csrf token

		csrfCheck := false
		if *tokenSourceType == "cookie" {
			csrfCheck = true
		}

		tokenData, err := auth.processAccessToken(rawToken, csrfCheck, csrfToken)
		if err != nil {
			return false, nil, nil
		}

		tokenAuth := tokenData.Auth
		if tokenAuth == "oidc" {
			externalID = tokenData.UID
			authType = "shibboleth"
		} else if tokenAuth == "rokwire_phone" {
			externalID = tokenData.UID
			authType = "phone"
		} else {
			return false, nil, nil
		}
	}

	//TODO - refactor!!!
	// if phone token then treat it as shibboleth
	if authType == "phone" {
		foundedUIN := auth.findUINByPhone(externalID)
		if foundedUIN == nil {
			//not found, it means that this phone is not added, so return unauthorized
			return false, nil, nil
		}
		//it was found
		externalID = *foundedUIN
		authType = "shibboleth"
	}

	return true, user, &authType
}

//token source type - cookie and header
func (auth *UserAuth) getTokens(r *http.Request) (*string, *string, *string, error) {
	//1. Check if there is a cookie
	cookie, err := r.Cookie("rokwire-access")
	if err == nil && cookie != nil && len(cookie.Value) > 0 {
		//there is a cookie
		tokenSourceType := "cookie"
		csrfToken := r.Header.Get("CSRF")

		return &cookie.Value, &tokenSourceType, &csrfToken, nil
	}

	//2. Check if there is a token in the Authorization header
	authorizationHeader := r.Header.Get("Authorization")
	if len(authorizationHeader) <= 0 {
		//no authorization
		return nil, nil, nil, nil
	}
	splitAuthorization := strings.Fields(authorizationHeader)
	if len(splitAuthorization) != 2 {
		//bad authorization
		return nil, nil, nil, nil
	}
	// expected - Bearer 1234
	if splitAuthorization[0] != "Bearer" {
		//bad authorization
		return nil, nil, nil, nil
	}
	token := splitAuthorization[1]
	tokenSourceType := "header"
	return &token, &tokenSourceType, nil, nil
}

func (auth *UserAuth) userCheck(w http.ResponseWriter, r *http.Request) (bool, *model.ShibbolethUser, *string) {
	//apply main check
	ok, user, authType := auth.mainCheck(r)
	if !ok {
		return false, nil, nil
	}

	return true, user, authType
}

//mobile app sends just token, the browser sends token + csrf token
func (auth *UserAuth) processAccessToken(token string, csrfCheck bool, csrfToken *string) (*tokenData, error) {

	//1. apply csrf check
	if csrfCheck {

		if csrfToken == nil || len(*csrfToken) == 0 {
			return nil, errors.New("missing csrf token")
		}

		crsfTokenData, err := auth.validateToken(*csrfToken, "csrf")
		if err != nil {
			log.Printf("error trying to validate csrf token - %s", err)
			return nil, err
		}

		if crsfTokenData == nil {
			log.Printf("not valid csrf token - %s", *csrfToken)
			return nil, errors.New("not valid csrf token")
		}
	}

	//2. apply access token check
	accessTokenData, err := auth.validateToken(token, "access")
	if err != nil {
		log.Printf("error trying to validate access token - %s", err)
		return nil, err
	}

	if accessTokenData == nil {
		log.Printf("not valid access token - %s", token)
		return nil, errors.New("not valid access token")
	}

	return accessTokenData, nil
}

//token type - access or csrf
func (auth *UserAuth) validateToken(token string, tokenType string) (*tokenData, error) {
	//extract the data - header and payload
	tokenSegments := strings.Split(token, ".")
	if len(tokenSegments) != 3 {
		return nil, errors.New("token segments count is != 3")
	}
	//header data
	headerData, err := jwt.DecodeSegment(tokenSegments[0])
	if err != nil {
		log.Printf("error decoding the header segment - %s", err)
		return nil, err
	}
	headerMap := make(map[string]string)
	err = json.Unmarshal(headerData, &headerMap)
	if err != nil {
		log.Println("error unmarshaling the header data" + err.Error())
		return nil, err
	}

	//payload
	payloadData, err := jwt.DecodeSegment(tokenSegments[1])
	if err != nil {
		log.Printf("error decoding the payload segment - %s", err)
		return nil, err
	}
	var tokenData *tokenData
	err = json.Unmarshal(payloadData, &tokenData)
	if err != nil {
		log.Println("error unmarshaling the payload data" + err.Error())
		return nil, err
	}

	//check issuer
	if tokenData.ISS != auth.Issuer {
		log.Printf("issuer does not match: - %s", tokenData.ISS)
		return nil, errors.New("issuer does not match:" + tokenData.ISS)
	}

	//check keys
	kid := headerMap["kid"]
	if len(kid) == 0 {
		log.Println("kid header is missing")
		return nil, errors.New("kid header is missing")
	}

	jwkKeyMatch, _ := auth.Keys.LookupKeyID(kid)
	if jwkKeyMatch == nil {
		log.Printf("no matching kid found")
		return nil, errors.New("no matching kid found")
	}
	publicKey := jwkKeyMatch.(jwk.RSAPublicKey)

	//validate
	jwk := rsa.PublicKey{}
	parsedToken, err := jwt.ParseWithClaims(token, jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
		if err := publicKey.Raw(&jwk); err != nil {
			log.Println("failed to create public key:", err)
			return nil, err
		}
		return &jwk, nil
	})
	if err != nil {
		log.Printf("error parse/validate token - %s", err)
		return nil, err
	}
	if !parsedToken.Valid {
		log.Printf("not valid token - %s", token)
		return nil, errors.New("not valid token:" + token)
	}

	//check token type
	if tokenData.Type != tokenType {
		log.Printf("invalid type %s", tokenData.Type)
		return nil, errors.New("invalid type - " + token)
	}

	return tokenData, nil
}

func (auth *UserAuth) processShibbolethToken(token string) (*userData, error) {
	// Validate the token
	idToken, err := auth.appIDTokenVerifier.Verify(context.Background(), token)
	if err != nil {
		log.Printf("error validating token - %s\n", err)
		return nil, err
	}

	// Get the user data from the token
	var userData *userData
	if err := idToken.Claims(&userData); err != nil {
		log.Printf("error getting user data from token - %s\n", err)
		return nil, err
	}
	//we must have UIuceduUIN
	if userData.UIuceduUIN == nil {
		log.Printf("missing uiuceuin data in the token - %s\n", token)
		return nil, errors.New("missing uiuceuin data in the token")
	}
	return userData, nil
}

func (auth *UserAuth) findUINByPhone(phone string) *string {
	rosters := auth.getRosters()
	if len(rosters) == 0 {
		return nil
	}

	for _, item := range rosters {
		cPhone := item["phone"]
		if cPhone == phone {
			uin := item["uin"]
			return &uin
		}
	}
	return nil
}

func (auth *UserAuth) processPhoneToken(token string) (*string, error) {
	claims := jwt.MapClaims{}
	_, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(auth.phoneAuthSecret), nil
	})
	if err != nil {
		return nil, err
	}

	for key, val := range claims {
		if key == "phoneNumber" {
			phoneValue := val.(string)
			return &phoneValue, nil
		}
	}
	return nil, errors.New("there is no phoneNumber claim in the phone token")
}

// type: 1 for shibboleth, 2 for phone, 3 for auth access token
// 1 & 2 are deprecated but we support them for back compatability
func (auth *UserAuth) getTokenType(token string) (*int, error) {
	parser := new(jwt.Parser)
	claims := jwt.MapClaims{}
	_, _, err := parser.ParseUnverified(token, claims)
	if err != nil {
		return nil, err
	}

	for key := range claims {
		if key == "uiucedu_uin" {
			tokenType := 1
			return &tokenType, nil
		}
		if key == "phoneNumber" {
			tokenType := 2
			return &tokenType, nil
		}
		if key == "uid" {
			tokenType := 3
			return &tokenType, nil
		}
	}
	return nil, errors.New("not supported token type")
}

func (auth *UserAuth) getCachedUser(externalID string) *cacheUser {
	auth.cachedUsersLock.RLock()
	defer auth.cachedUsersLock.RUnlock()

	var cachedUser *cacheUser //to return

	item, _ := auth.cachedUsers.Load(externalID)
	if item != nil {
		cachedUser = item.(*cacheUser)
	}

	//keep the last get time
	if cachedUser != nil {
		cachedUser.lastUsage = time.Now()
		auth.cachedUsers.Store(externalID, cachedUser)
	}

	return cachedUser
}

func (auth *UserAuth) cacheUser(externalID string) {
	auth.cachedUsersLock.RLock()

	cacheUser := &cacheUser{lastUsage: time.Now()}
	auth.cachedUsers.Store(externalID, cacheUser)

	auth.cachedUsersLock.RUnlock()
}

func (auth *UserAuth) deleteCacheUser(externalID string) {
	auth.cachedUsersLock.RLock()

	auth.cachedUsers.Delete(externalID)

	auth.cachedUsersLock.RUnlock()
}

func (auth *UserAuth) clearCacheUsers() {
	log.Println("UserAuth -> clearCacheUsers")

	auth.cachedUsersLock.RLock()

	auth.cachedUsers = &syncmap.Map{}

	auth.cachedUsersLock.RUnlock()
}

func (auth *UserAuth) setRosters(rosters []map[string]string) {
	auth.rostersLock.RLock()

	auth.rosters = rosters

	auth.rostersLock.RUnlock()
}

func (auth *UserAuth) getRosters() []map[string]string {
	auth.rostersLock.RLock()
	defer auth.rostersLock.RUnlock()

	return auth.rosters
}

func (auth *UserAuth) responseBadRequest(w http.ResponseWriter) {
	log.Println(fmt.Sprintf("400 - Bad Request"))

	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte("Bad Request"))
}

func (auth *UserAuth) responseUnauthorized(logInfo string, w http.ResponseWriter) {
	log.Println(fmt.Sprintf("401 - Unauthorized - %s", logInfo))

	w.WriteHeader(http.StatusUnauthorized)
	w.Write([]byte("Unauthorized"))
}

func (auth *UserAuth) responseInternalServerError(w http.ResponseWriter) {
	log.Println(fmt.Sprintf("500 - Internal Server Error"))

	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte("Internal Server Error"))
}

func (auth *UserAuth) responseForbbiden(info string, w http.ResponseWriter) {
	log.Printf("403 - Forbidden - %s", info)

	w.WriteHeader(http.StatusForbidden)
	w.Write([]byte("Forbidden"))
}

func newUserAuth(app *core.Application, oidcProvider string, oidcAppClientID string,
	phoneAuthSecret string, keys string, issuer string) *UserAuth {

	provider, err := oidc.NewProvider(context.Background(), oidcProvider)
	if err != nil {
		log.Fatalln(err)
	}
	appIDTokenVerifier := provider.Verifier(&oidc.Config{ClientID: oidcAppClientID})

	keysSet, err := jwk.ParseString(keys)
	if err != nil {
		log.Fatalln(err)
	}

	cacheUsers := &syncmap.Map{}
	lock := &sync.RWMutex{}

	cacheRosters := []map[string]string{}
	rostersLock := &sync.RWMutex{}

	auth := UserAuth{app: app, appIDTokenVerifier: appIDTokenVerifier, phoneAuthSecret: phoneAuthSecret, Keys: keysSet, Issuer: issuer,
		cachedUsers: cacheUsers, cachedUsersLock: lock, rosters: cacheRosters, rostersLock: rostersLock}
	return &auth
}
