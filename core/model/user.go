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

package model

import (
	"github.com/rokwire/rokwire-building-block-sdk-go/utils/logging/logutils"
)

const (
	//TypeShibbolethUser type
	TypeShibbolethUser logutils.MessageDataType = "shibboleth_user"
)

//////////////////////////

// ShibbolethUser represents shibboleth auth entity
type ShibbolethUser struct {
	Uin        *string   `json:"uiucedu_uin" bson:"uin"`
	Email      *string   `json:"email" bson:"email"`
	Phone      *string   `json:"phone" bson:"phone"`
	Membership *[]string `json:"uiucedu_is_member_of,omitempty" bson:"membership,omitempty"`
} //@name ShibbolethUser
