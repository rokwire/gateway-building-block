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

package uiuc

import (
	model "application/core/model"
)

// GIESCourse represents a light weight course definition used for GIES specific student operations
type GIESCourse struct {
	Term       string `json:"Term"`
	Subject    string `json:"Subject"`
	Number     string `json:"Number"`
	Section    string `json:"Section"`
	Title      string `json:"Title"`
	Instructor string `json:"Instructor"`
}

// NewGiesCourse returns an app formatted GiesCourse object from the campus definition
func NewGiesCourse(cr GIESCourse) *model.GiesCourse {
	ret := model.GiesCourse{}
	ret.Instructor = cr.Instructor
	ret.Number = cr.Number
	ret.Section = cr.Section
	ret.Subject = cr.Subject
	ret.Term = cr.Term
	ret.Title = cr.Title
	return &ret
}
