// Copyright 2021 Board of Trustees of the University of Illinois
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

package logutils

import (
	"fmt"
	"strings"
)

type Fields map[string]interface{}

func (f Fields) ToMap() map[string]interface{} {
	return f
}

type MessageArgs interface {
	String() string
}

type FieldArgs Fields

func (f *FieldArgs) String() string {
	if f == nil {
		return ""
	}

	argMsg := ""
	for k, v := range *f {
		if argMsg != "" {
			argMsg += ", "
		}

		if v != nil {
			argMsg += fmt.Sprintf("%s=%v", k, v)
		} else {
			argMsg += k
		}
	}

	return argMsg
}

type ListArgs []string

func (l *ListArgs) String() string {
	if l == nil {
		return ""
	}

	return strings.Join(*l, ", ")
}

type StringArgs string

func (s StringArgs) String() string {
	return string(s)
}

type MessageDataStatus string
type MessageActionStatus string

type MessageActionType string

type MessageDataType string

const (
	//Errors
	Unimplemented string = "Unimplemented"

	//Types
	StatusValid   MessageDataStatus = "Valid"
	StatusInvalid MessageDataStatus = "Invalid"
	StatusFound   MessageDataStatus = "Found"
	StatusMissing MessageDataStatus = "Missing"

	StatusSuccess MessageActionStatus = "Success"
	StatusError   MessageActionStatus = "Error"

	//Data
	TypeArg         MessageDataType = "arg"
	TypeTransaction MessageDataType = "transaction"
	TypeResult      MessageDataType = "result"

	//Primitives
	TypeInt    MessageDataType = "int"
	TypeUint   MessageDataType = "uint"
	TypeFloat  MessageDataType = "float"
	TypeBool   MessageDataType = "bool"
	TypeString MessageDataType = "string"
	TypeByte   MessageDataType = "byte"
	TypeError  MessageDataType = "error"

	//Requests
	TypeRequest      MessageDataType = "request"
	TypeRequestBody  MessageDataType = "request body"
	TypeResponse     MessageDataType = "response"
	TypeResponseBody MessageDataType = "response body"
	TypeQueryParam   MessageDataType = "query param"

	//Auth
	TypeToken      MessageDataType = "token"
	TypeClaims     MessageDataType = "claims"
	TypeClaim      MessageDataType = "claim"
	TypeScope      MessageDataType = "scope"
	TypePermission MessageDataType = "permission"

	//Actions
	ActionInitialize MessageActionType = "initializing"
	ActionCompute    MessageActionType = "computing"
	ActionRegister   MessageActionType = "registering"
	ActionDeregister MessageActionType = "deregistering"
	ActionStart      MessageActionType = "starting"
	ActionCommit     MessageActionType = "committing"
	ActionRefresh    MessageActionType = "refreshing"

	//Encryption Actions
	ActionEncrypt MessageActionType = "encrypting"
	ActionDecrypt MessageActionType = "decrypting"

	//Request/Response Actions
	ActionSend MessageActionType = "sending"
	ActionRead MessageActionType = "reading"

	//Encode Actions
	ActionParse  MessageActionType = "parsing"
	ActionEncode MessageActionType = "encoding"
	ActionDecode MessageActionType = "decoding"

	//Marshal Actions
	ActionMarshal   MessageActionType = "marshalling"
	ActionUnmarshal MessageActionType = "unmarshalling"
	ActionValidate  MessageActionType = "validating"
	ActionCast      MessageActionType = "casting to"

	//Cache Actions
	ActionCache     MessageActionType = "caching"
	ActionLoadCache MessageActionType = "loading cached"

	//Operation Actions
	ActionGet    MessageActionType = "getting"
	ActionCreate MessageActionType = "creating"
	ActionUpdate MessageActionType = "updating"
	ActionDelete MessageActionType = "deleting"

	//Storage Actions
	ActionLoad    MessageActionType = "loading"
	ActionFind    MessageActionType = "finding"
	ActionInsert  MessageActionType = "inserting"
	ActionReplace MessageActionType = "replacing"
	ActionSave    MessageActionType = "saving"
	ActionCount   MessageActionType = "counting"
)
