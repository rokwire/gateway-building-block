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

package web

import (
	"net/http"

	"github.com/rokwire/core-auth-library-go/v3/authorization"
	"github.com/rokwire/logging-library-go/v2/errors"
	"github.com/rokwire/logging-library-go/v2/logutils"

	"github.com/rokwire/core-auth-library-go/v3/authservice"
	"github.com/rokwire/core-auth-library-go/v3/tokenauth"
)

// Auth handler
type Auth struct {
	client tokenauth.Handlers
	admin  tokenauth.Handlers
	bbs    tokenauth.Handlers
	tps    tokenauth.Handlers
	system tokenauth.Handlers
}

// NewAuth creates new auth handler
func NewAuth(serviceRegManager *authservice.ServiceRegManager) (*Auth, error) {
	client, err := newClientAuth(serviceRegManager)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionCreate, "client auth", nil, err)
	}
	clientHandlers := tokenauth.NewHandlers(client)

	admin, err := newAdminAuth(serviceRegManager)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionCreate, "admin auth", nil, err)
	}
	adminHandlers := tokenauth.NewHandlers(admin)

	bbs, err := newBBsAuth(serviceRegManager)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionCreate, "bbs auth", nil, err)
	}
	bbsHandlers := tokenauth.NewHandlers(bbs)

	tps, err := newTPSAuth(serviceRegManager)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionCreate, "tps auth", nil, err)
	}
	tpsHandlers := tokenauth.NewHandlers(tps)

	system, err := newSystemAuth(serviceRegManager)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionCreate, "system auth", nil, err)
	}
	systemHandlers := tokenauth.NewHandlers(system)

	auth := Auth{
		client: clientHandlers,
		admin:  adminHandlers,
		bbs:    bbsHandlers,
		tps:    tpsHandlers,
		system: systemHandlers,
	}
	return &auth, nil
}

///////

func newClientAuth(serviceRegManager *authservice.ServiceRegManager) (*tokenauth.StandardHandler, error) {
	clientPermissionAuth := authorization.NewCasbinStringAuthorization("driver/web/client_permission_policy.csv")
	clientScopeAuth := authorization.NewCasbinScopeAuthorization("driver/web/client_scope_policy.csv", serviceRegManager.AuthService.ServiceID)
	clientTokenAuth, err := tokenauth.NewTokenAuth(true, serviceRegManager, clientPermissionAuth, clientScopeAuth)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionCreate, "client token auth", nil, err)
	}

	check := func(claims *tokenauth.Claims, req *http.Request) (int, error) {
		if claims.Admin {
			return http.StatusUnauthorized, errors.ErrorData(logutils.StatusInvalid, "admin claim", nil)
		}
		if claims.System {
			return http.StatusUnauthorized, errors.ErrorData(logutils.StatusInvalid, "system claim", nil)
		}

		return http.StatusOK, nil
	}

	auth := tokenauth.NewScopeHandler(clientTokenAuth, check)
	return auth, nil
}

func newAdminAuth(serviceRegManager *authservice.ServiceRegManager) (*tokenauth.StandardHandler, error) {
	adminPermissionAuth := authorization.NewCasbinStringAuthorization("driver/web/admin_permission_policy.csv")
	adminTokenAuth, err := tokenauth.NewTokenAuth(true, serviceRegManager, adminPermissionAuth, nil)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionCreate, "admin token auth", nil, err)
	}

	check := func(claims *tokenauth.Claims, req *http.Request) (int, error) {
		if !claims.Admin {
			return http.StatusUnauthorized, errors.ErrorData(logutils.StatusInvalid, "admin claim", nil)
		}

		return http.StatusOK, nil
	}

	auth := tokenauth.NewStandardHandler(adminTokenAuth, check)
	return auth, nil
}

func newBBsAuth(serviceRegManager *authservice.ServiceRegManager) (*tokenauth.StandardHandler, error) {
	bbsPermissionAuth := authorization.NewCasbinStringAuthorization("driver/web/bbs_permission_policy.csv")
	bbsTokenAuth, err := tokenauth.NewTokenAuth(true, serviceRegManager, bbsPermissionAuth, nil)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionStart, "bbs token auth", nil, err)
	}

	check := func(claims *tokenauth.Claims, req *http.Request) (int, error) {
		if !claims.Service {
			return http.StatusUnauthorized, errors.ErrorData(logutils.StatusInvalid, "service claim", nil)
		}

		if !claims.FirstParty {
			return http.StatusUnauthorized, errors.ErrorData(logutils.StatusInvalid, "first party claim", nil)
		}

		return http.StatusOK, nil
	}

	auth := tokenauth.NewStandardHandler(bbsTokenAuth, check)
	return auth, nil
}

func newTPSAuth(serviceRegManager *authservice.ServiceRegManager) (*tokenauth.StandardHandler, error) {
	tpsPermissionAuth := authorization.NewCasbinStringAuthorization("driver/web/tps_permission_policy.csv")
	tpsTokenAuth, err := tokenauth.NewTokenAuth(true, serviceRegManager, tpsPermissionAuth, nil)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionStart, "tps token auth", nil, err)
	}

	check := func(claims *tokenauth.Claims, req *http.Request) (int, error) {
		if !claims.Service {
			return http.StatusUnauthorized, errors.ErrorData(logutils.StatusInvalid, "service claim", nil)
		}

		if claims.FirstParty {
			return http.StatusUnauthorized, errors.ErrorData(logutils.StatusInvalid, "first party claim", nil)
		}

		return http.StatusOK, nil
	}

	auth := tokenauth.NewStandardHandler(tpsTokenAuth, check)
	return auth, nil
}

func newSystemAuth(serviceRegManager *authservice.ServiceRegManager) (*tokenauth.StandardHandler, error) {
	systemPermissionAuth := authorization.NewCasbinStringAuthorization("driver/web/system_permission_policy.csv")
	systemTokenAuth, err := tokenauth.NewTokenAuth(true, serviceRegManager, systemPermissionAuth, nil)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionCreate, "system token auth", nil, err)
	}

	check := func(claims *tokenauth.Claims, req *http.Request) (int, error) {
		if !claims.System {
			return http.StatusUnauthorized, errors.ErrorData(logutils.StatusInvalid, "system claim", nil)
		}

		return http.StatusOK, nil
	}

	auth := tokenauth.NewStandardHandler(systemTokenAuth, check)
	return auth, nil
}
