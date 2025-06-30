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

// LoadAppBuildingFeatures loads all of the configured building features
func (a *Adapter) LoadAppBuildingFeatures() ([]model.AppBuildingFeature, error) {

	var data []model.AppBuildingFeature
	//filter := bson.M{}
	timeout := 15 * time.Second //15 seconds timeout
	//err := a.db.legacyEvents.FindWithParams(context, filter, &data, nil, &timeout)
	err := a.db.appbuildingfeatures.FindWithParams(a.context, nil, &data, nil, &timeout)
	//err := a.db.appbuildingfeatures.FindWithContext(a.context, nil, &data, nil)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionFind, model.TypeExample, filterArgs(nil), err)
	}

	return data, nil
}
