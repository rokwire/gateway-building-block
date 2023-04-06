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
	"application/core"
	"net/http"

	"github.com/rokwire/core-auth-library-go/v3/tokenauth"
	"github.com/rokwire/logging-library-go/v2/logs"
)

// DefaultAPIsHandler handles the default rest APIs implementation
type DefaultAPIsHandler struct {
	app *core.Application
}

func (h DefaultAPIsHandler) version(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	return l.HTTPResponseSuccessMessage(h.app.Default.GetVersion())
}

// NewDefaultAPIsHandler creates new default API Handler instance
func NewDefaultAPIsHandler(app *core.Application) DefaultAPIsHandler {
	return DefaultAPIsHandler{app: app}
}
