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

package tokenauth

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/rokwire/core-auth-library-go/authorization"
	"github.com/rokwire/core-auth-library-go/authservice"
	"github.com/rokwire/core-auth-library-go/authutils"
)

const (
	AudRokwire string = "rokwire"
)

// Claims represents the standard claims included in access tokens
type Claims struct {
	// Required Standard Claims: sub, aud, exp, iat
	jwt.StandardClaims
	OrgID         string `json:"org_id" validate:"required"`    // Organization ID
	AppID         string `json:"app_id"`                        // Application ID
	Purpose       string `json:"purpose" validate:"required"`   // Token purpose (eg. access, csrf...)
	AuthType      string `json:"auth_type" validate:"required"` // Authentication method (eg. email, phone...)
	Permissions   string `json:"permissions"`                   // Granted permissions
	Scope         string `json:"scope"`                         // Granted scope
	Anonymous     bool   `json:"anonymous"`                     // Is the user anonymous?
	Authenticated bool   `json:"authenticated"`                 // Did the user authenticate? (false on refresh)
	Service       bool   `json:"service"`                       // Is this token for a service account?
	Admin         bool   `json:"admin"`                         // Is this token for an admin?

	// User Data: DO NOT USE AS IDENTIFIER OR SHARE WITH THIRD-PARTY SERVICES
	Name  string `json:"name,omitempty"`  // User full name
	Email string `json:"email,omitempty"` // User email address
	Phone string `json:"phone,omitempty"` // User phone number

	//TODO: Once the new user ID scheme has been adopted across all services these claims should be removed
	UID string `json:"uid,omitempty"` // Unique user identifier for specified "auth_type"
}

// TokenAuth contains configurations and helper functions required to validate tokens
type TokenAuth struct {
	authService         *authservice.AuthService
	acceptRokwireTokens bool

	permissionAuth authorization.Authorization
	scopeAuth      authorization.Authorization

	blacklist     []string
	blacklistLock *sync.RWMutex
	blacklistSize int
}

// CheckToken validates the provided token and returns the token claims
func (t *TokenAuth) CheckToken(token string, purpose string) (*Claims, error) {
	for i := len(t.blacklist) - 1; i >= 0; i-- {
		if token == t.blacklist[i] {
			return nil, fmt.Errorf("known invalid token")
		}
	}
	authServiceReg, err := t.authService.GetServiceRegWithPubKey("auth")
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve auth service pub key: %v", err)
	}

	parsedToken, tokenErr := jwt.ParseWithClaims(token, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return authServiceReg.PubKey.Key, nil
	})
	if parsedToken == nil {
		return nil, errors.New("failed to parse token")
	}

	claims, ok := parsedToken.Claims.(*Claims)
	if !ok {
		return nil, errors.New("failed to parse token claims")
	}

	// Check token claims
	if claims.Subject == "" {
		return nil, errors.New("token sub missing")
	}
	if claims.ExpiresAt == 0 {
		return nil, errors.New("token exp missing")
	}
	if claims.IssuedAt == 0 {
		return nil, errors.New("token iat missing")
	}
	if claims.OrgID == "" {
		return nil, errors.New("token org_id missing")
	}
	if claims.AuthType == "" {
		return nil, errors.New("token auth_type missing")
	}
	if claims.Issuer != authServiceReg.Host {
		return nil, fmt.Errorf("token iss (%s) does not match %s", claims.Issuer, authServiceReg.Host)
	}
	if claims.Purpose != purpose {
		return nil, fmt.Errorf("token purpose (%s) does not match %s", claims.Purpose, purpose)
	}

	aud := strings.Split(claims.Audience, ",")
	if !(authutils.ContainsString(aud, t.authService.GetServiceID()) || (t.acceptRokwireTokens && authutils.ContainsString(aud, AudRokwire))) {
		acceptAuds := t.authService.GetServiceID()
		if t.acceptRokwireTokens {
			acceptAuds += " or " + AudRokwire
		}

		return nil, fmt.Errorf("token aud (%s) does not match %s", claims.Audience, acceptAuds)
	}

	// Check token headers
	alg, _ := parsedToken.Header["alg"].(string)
	if alg != authServiceReg.PubKey.Alg {
		return nil, fmt.Errorf("token alg (%s) does not match %s", alg, authServiceReg.PubKey.Alg)
	}
	typ, _ := parsedToken.Header["typ"].(string)
	if typ != "JWT" {
		return nil, fmt.Errorf("token typ (%s) does not match JWT", typ)
	}
	kid, _ := parsedToken.Header["kid"].(string)
	if kid != authServiceReg.PubKey.Kid {
		if !parsedToken.Valid {
			if claims.ExpiresAt > time.Now().Unix() {
				refreshed, refreshErr := t.authService.CheckForRefresh()
				if refreshErr != nil {
					return nil, fmt.Errorf("initial token check returned invalid, error on retry: %v", refreshErr)
				}
				if refreshed {
					return t.retryCheckToken(token, purpose)
				} else {
					return nil, fmt.Errorf("token invalid: %v", tokenErr)
				}
			}
			return nil, fmt.Errorf("token is expired %d", claims.ExpiresAt)
		}
		return nil, fmt.Errorf("token has valid signature but invalid kid %s", kid)
	}

	if !parsedToken.Valid {
		return nil, fmt.Errorf("token invalid: %v", tokenErr)
	}

	return claims, nil
}

func (t *TokenAuth) retryCheckToken(token string, purpose string) (*Claims, error) {
	retryClaims, retryErr := t.CheckToken(token, purpose)
	if retryErr != nil {
		t.blacklistLock.Lock()
		if len(t.blacklist) >= t.blacklistSize {
			t.blacklist = t.blacklist[1:]
		}
		t.blacklist = append(t.blacklist, token)
		t.blacklistLock.Unlock()
	}
	return retryClaims, retryErr
}

// CheckRequestTokens is a convenience function which retrieves and checks any tokens included in a request
// and returns the access token claims
// Mobile Clients/Secure Servers: Access tokens must be provided as a Bearer token
//								  in the "Authorization" header
// Web Clients: Access tokens must be provided in the "rokwire-access-token" cookie
//				and CSRF tokens must be provided in the "CSRF" header
func (t *TokenAuth) CheckRequestTokens(r *http.Request) (*Claims, error) {
	accessToken, csrfToken, err := GetRequestTokens(r)
	if err != nil {
		return nil, fmt.Errorf("error getting request tokens: %v", err)
	}

	accessClaims, err := t.CheckToken(accessToken, "access")
	if err != nil {
		return nil, fmt.Errorf("error validating access token: %v", err)
	}

	if csrfToken != "" {
		csrfClaims, err := t.CheckToken(csrfToken, "csrf")
		if err != nil {
			return nil, fmt.Errorf("error validating csrf token: %v", err)
		}

		err = t.ValidateCsrfTokenClaims(accessClaims, csrfClaims)
		if err != nil {
			return nil, fmt.Errorf("error validating csrf token claims: %v", err)
		}
	}

	return accessClaims, nil
}

// ValidateCsrfTokenClaims will validate that the CSRF token claims appropriately match the access token claims
//	Returns nil on success and error on failure.
func (t *TokenAuth) ValidateCsrfTokenClaims(accessClaims *Claims, csrfClaims *Claims) error {
	if csrfClaims.Subject != accessClaims.Subject {
		return fmt.Errorf("csrf sub (%s) does not match access sub (%s)", csrfClaims.Subject, accessClaims.Subject)
	}

	if csrfClaims.OrgID != accessClaims.OrgID {
		return fmt.Errorf("csrf org_id (%s) does not match access org_id (%s)", csrfClaims.OrgID, accessClaims.OrgID)
	}

	return nil
}

// ValidatePermissionsClaim will validate that the provided token claims contain one or more of the required permissions
//	Returns nil on success and error on failure.
func (t *TokenAuth) ValidatePermissionsClaim(claims *Claims, requiredPermissions []string) error {
	if len(requiredPermissions) == 0 {
		return nil
	}

	if claims.Permissions == "" {
		return errors.New("permissions claim empty")
	}

	// Grant access if claims contain any of the required permissions
	permissions := strings.Split(claims.Permissions, ",")
	for _, v := range requiredPermissions {
		if authutils.ContainsString(permissions, v) {
			return nil
		}
	}

	return fmt.Errorf("required permissions not found: required %v, found %s", requiredPermissions, claims.Permissions)
}

// AuthorizeRequestPermissions will authorize the request if the permissions claim passes the permissionsAuth
//	Returns nil on success and error on failure.
func (t *TokenAuth) AuthorizeRequestPermissions(claims *Claims, request *http.Request) error {
	if claims == nil || claims.Permissions == "" {
		return errors.New("permissions claim empty")
	}

	permissions := strings.Split(claims.Permissions, ",")
	object := request.URL.Path
	action := request.Method

	return t.permissionAuth.Any(permissions, object, action)
}

// ValidateScopeClaim will validate that the provided token claims contain the required scope
// 	If an empty required scope is provided, the claims must contain a valid global scope such as 'all' or '{service}:all'
//	Returns nil on success and error on failure.
func (t *TokenAuth) ValidateScopeClaim(claims *Claims, requiredScope string) error {
	if claims == nil || claims.Scope == "" {
		return errors.New("scope claim empty")
	}

	scopes := strings.Split(claims.Scope, " ")
	if authorization.CheckScopesGlobals(scopes, t.authService.GetServiceID()) {
		return nil
	}

	required, err := authorization.ScopeFromString(requiredScope)
	if err != nil {
		return fmt.Errorf("invalid required scope: %v", err)
	}

	for _, scopeString := range scopes {
		scope, err := authorization.ScopeFromString(scopeString)
		if err != nil {
			continue
		}

		if scope.Match(required) {
			return nil
		}
	}

	return fmt.Errorf("required scope not found: required %s, found %s", requiredScope, claims.Scope)
}

// AuthorizeRequestScope will authorize the request if the scope claim passes the scopeAuth
//	Returns nil on success and error on failure.
func (t *TokenAuth) AuthorizeRequestScope(claims *Claims, request *http.Request) error {
	if claims == nil || claims.Scope == "" {
		return errors.New("scope claim empty")
	}

	scopes := strings.Split(claims.Scope, " ")
	object := request.URL.Path
	action := request.Method

	return t.scopeAuth.Any(scopes, object, action)
}

// SetBlacklistSize sets the maximum size of the token blacklist queue
// 	The default value is 1024
func (t *TokenAuth) SetBlacklistSize(size int) {
	t.blacklistLock.Lock()
	t.blacklistSize = size
	t.blacklistLock.Unlock()
}

// NewTokenAuth creates and configures a new TokenAuth instance
// authorization maybe nil if performing manual authorization
func NewTokenAuth(acceptRokwireTokens bool, authService *authservice.AuthService, permissionAuth authorization.Authorization, scopeAuth authorization.Authorization) (*TokenAuth, error) {
	authService.SubscribeServices([]string{"auth"}, true)

	blLock := &sync.RWMutex{}
	bl := []string{}

	return &TokenAuth{acceptRokwireTokens: acceptRokwireTokens, authService: authService, permissionAuth: permissionAuth, scopeAuth: scopeAuth, blacklistLock: blLock, blacklist: bl, blacklistSize: 1024}, nil
}

// -------------------------- Helper Functions --------------------------

// GetRequestTokens retrieves tokens from the request headers and/or cookies
// Mobile Clients/Secure Servers: Access tokens must be provided as a Bearer token
//								  in the "Authorization" header
// Web Clients: Access tokens must be provided in the "rokwire-access-token" cookie
//				and CSRF tokens must be provided in the "CSRF" header
func GetRequestTokens(r *http.Request) (string, string, error) {
	authorizationHeader := r.Header.Get("Authorization")
	if authorizationHeader != "" {
		splitAuthorization := strings.Fields(authorizationHeader)
		if len(splitAuthorization) != 2 {
			return "", "", errors.New("invalid authorization header format")
		}
		if strings.ToLower(splitAuthorization[0]) != "bearer" {
			return "", "", errors.New("authorization header missing bearer token")
		}
		idToken := splitAuthorization[1]

		return idToken, "", nil
	}

	csrfToken := r.Header.Get("CSRF")
	if csrfToken == "" {
		return "", "", errors.New("missing authorization and csrf header")
	}

	accessCookie, err := r.Cookie("rokwire-access-token")
	if err != nil || accessCookie == nil || accessCookie.Value == "" {
		return "", "", errors.New("missing access token")
	}

	return accessCookie.Value, csrfToken, nil
}
