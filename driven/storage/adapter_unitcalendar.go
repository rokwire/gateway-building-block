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

package storage

import (
	"application/core/model"
	"strconv"

	"github.com/rokwire/logging-library-go/v2/errors"
	"github.com/rokwire/logging-library-go/v2/logutils"
	"go.mongodb.org/mongo-driver/bson"
)

// FindCalendars finds all calendars for a given unit id
func (a *Adapter) FindCalendars(id string) (*[]model.UnitCalendar, error) {
	//filter := bson.M{"org_id": orgID, "app_id": appID, "unit_id": id}
	intid, _ := strconv.Atoi(id)
	filter := bson.M{"unit_id": intid}
	//filter := bson.M{}

	var data []model.UnitCalendar
	err := a.db.unitcalendars.FindWithContext(a.context, filter, &data, nil)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionFind, model.TypeExample, filterArgs(nil), err)
	}

	return &data, nil
}
