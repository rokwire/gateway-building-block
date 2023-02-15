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

package terms

import (
	model "apigateway/core/model"
	"strconv"
	"time"
)

//TermSessionAdapter is a uiuc specific structure that implements the term session interface
type TermSessionAdapter struct {
}

//NewTermSessionAdapter returns a vendor specific implementation of the contanct information interface
func NewTermSessionAdapter() *TermSessionAdapter {
	return &TermSessionAdapter{}

}

//GetTermSessions returns a list of term sessions to the client
func (lv *TermSessionAdapter) GetTermSessions() (*[4]model.TermSession, error) {
	var termSessions [4]model.TermSession
	//if beginning of June make fall term default, spring semester has ended
	crntDate := time.Now()
	crntYear := strconv.Itoa(crntDate.Year())
	if crntDate.Month() >= 6 && crntDate.Month() <= 12 {
		nextYear := strconv.Itoa(crntDate.Year() + 1)
		termSessions[2] = model.TermSession{Term: "Fall - " + crntYear, TermID: "1" + crntYear + "8", CurrentTerm: true}
		termSessions[1] = model.TermSession{Term: "Summer - " + crntYear, TermID: "1" + crntYear + "5", CurrentTerm: false}
		termSessions[0] = model.TermSession{Term: "Spring - " + crntYear, TermID: "1" + crntYear + "1", CurrentTerm: false}
		termSessions[3] = model.TermSession{Term: "Spring - " + nextYear, TermID: "1" + nextYear + "1", CurrentTerm: false}
	} else {
		pastYear := strconv.Itoa(crntDate.Year() - 1)
		termSessions[1] = model.TermSession{Term: "Spring - " + crntYear, TermID: "1" + crntYear + "1", CurrentTerm: true}
		termSessions[2] = model.TermSession{Term: "Summer - " + crntYear, TermID: "1" + crntYear + "5", CurrentTerm: false}
		termSessions[0] = model.TermSession{Term: "Fall - " + pastYear, TermID: "1" + pastYear + "8", CurrentTerm: false}
		termSessions[3] = model.TermSession{Term: "Fall - " + crntYear, TermID: "1" + crntYear + "8", CurrentTerm: false}
	}
	return &termSessions, nil
}
