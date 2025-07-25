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

package model

import (
	"github.com/rokwire/rokwire-building-block-sdk-go/utils/logging/logutils"
)

const (
	//TypeTermSession type
	TypeTermSession logutils.MessageDataType = "termsessions"
)

// TermSession represents the elements of a term session
type TermSession struct {
	Term        string `json:"term" bson:"term"`
	TermID      string `json:"termid" bson:"termid"`
	CurrentTerm bool   `json:"is_current" bson:"is_current"`
}
