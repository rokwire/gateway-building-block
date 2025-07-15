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
	"time"

	"github.com/rokwire/rokwire-building-block-sdk-go/utils/errors"
	"github.com/rokwire/rokwire-building-block-sdk-go/utils/logging/logutils"
)

// LoadFloorPlanMarkup loads the markup needed to wrap around a floor plan svg
func (a *Adapter) LoadFloorPlanMarkup() (*model.FloorPlanMarkup, error) {
	var data []model.FloorPlanMarkup

	timeout := 15 * time.Second //15 seconds timeout
	err := a.db.floorplanmarkup.FindWithParams(a.context, nil, &data, nil, &timeout)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionFind, model.TypeFloorPlanMarkup, filterArgs(nil), err)
	}
	if len(data) == 0 {
		err = errors.Newf("no floorplan markup found")
		return nil, errors.WrapErrorAction(logutils.ActionFind, model.TypeFloorPlanMarkup, filterArgs(nil), err)
	}
	if len(data) > 1 {
		err = errors.Newf("multiple floorplan markup found")
		return nil, errors.WrapErrorAction(logutils.ActionFind, model.TypeFloorPlanMarkup, filterArgs(nil), err)
	}
	fpm := data[0]

	return &fpm, nil
}
