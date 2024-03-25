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

package core

import (
	"application/core/model"
	"time"

	"github.com/google/uuid"
	"github.com/rokwire/core-auth-library-go/v3/authutils"
	"github.com/rokwire/core-auth-library-go/v3/tokenauth"
	"github.com/rokwire/logging-library-go/v2/errors"
	"github.com/rokwire/logging-library-go/v2/logutils"
)

// appAdmin contains admin implementations
type appAdmin struct {
	app *Application
}

// GetExample gets an Example by ID
func (a appAdmin) GetExample(orgID string, appID string, id string) (*model.Example, error) {
	return a.app.shared.getExample(orgID, appID, id)
}

// CreateExample creates a new Example
func (a appAdmin) CreateExample(example model.Example) (*model.Example, error) {
	example.ID = uuid.NewString()
	err := a.app.storage.InsertExample(example)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionCreate, model.TypeExample, nil, err)
	}
	return &example, nil
}

// UpdateExample updates an Example
func (a appAdmin) UpdateExample(example model.Example) error {
	return a.app.storage.UpdateExample(example)
}

// AppendExample appends to the data in an example - Example of transaction usage
func (a appAdmin) AppendExample(example model.Example) (*model.Example, error) {
	/*now := time.Now()
	var newExample *model.Example
	transaction := func(storage interfaces.Storage) error {
		oldExample, err := storage.FindExample(example.OrgID, example.AppID, example.ID)
		if err != nil || oldExample == nil {
			return errors.WrapErrorAction(logutils.ActionFind, model.TypeExample, nil, err)
		}

		oldExample.Data = oldExample.Data + "," + example.Data
		oldExample.DateUpdated = &now

		err = storage.UpdateExample(*oldExample)
		if err != nil {
			return errors.WrapErrorAction(logutils.ActionUpdate, model.TypeExample, nil, err)
		}

		newExample = oldExample
		return nil
	}

	err := a.app.storage.PerformTransaction(transaction)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionCommit, logutils.TypeTransaction, nil, err)
	}

	return newExample, nil */
	return nil, nil
}

// DeleteExample deletes an Example by ID
func (a appAdmin) DeleteExample(orgID string, appID string, id string) error {
	return a.app.storage.DeleteExample(orgID, appID, id)
}

func (a appAdmin) GetConfig(id string, claims *tokenauth.Claims) (*model.Config, error) {
	config, err := a.app.storage.FindConfigByID(id)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionFind, model.TypeConfig, nil, err)
	}
	if config == nil {
		return nil, errors.ErrorData(logutils.StatusMissing, model.TypeConfig, &logutils.FieldArgs{"id": id})
	}

	err = claims.CanAccess(config.AppID, config.OrgID, config.System)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionValidate, "config access", nil, err)
	}

	return config, nil
}

func (a appAdmin) GetConfigs(configType *string, claims *tokenauth.Claims) ([]model.Config, error) {
	configs, err := a.app.storage.FindConfigs(configType)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionFind, model.TypeConfig, nil, err)
	}

	allowedConfigs := make([]model.Config, 0)
	for _, config := range configs {
		if err := claims.CanAccess(config.AppID, config.OrgID, config.System); err == nil {
			allowedConfigs = append(allowedConfigs, config)
		}
	}
	return allowedConfigs, nil
}

func (a appAdmin) CreateConfig(config model.Config, claims *tokenauth.Claims) (*model.Config, error) {
	// must be a system config if applying to all orgs
	if config.OrgID == authutils.AllOrgs && !config.System {
		return nil, errors.ErrorData(logutils.StatusInvalid, "config system status", &logutils.FieldArgs{"config.org_id": authutils.AllOrgs})
	}

	err := claims.CanAccess(config.AppID, config.OrgID, config.System)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionValidate, "config access", nil, err)
	}

	config.ID = uuid.NewString()
	config.DateCreated = time.Now().UTC()
	err = a.app.storage.InsertConfig(config)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionInsert, model.TypeConfig, nil, err)
	}
	return &config, nil
}

func (a appAdmin) UpdateConfig(config model.Config, claims *tokenauth.Claims) error {
	// must be a system config if applying to all orgs
	if config.OrgID == authutils.AllOrgs && !config.System {
		return errors.ErrorData(logutils.StatusInvalid, "config system status", &logutils.FieldArgs{"config.org_id": authutils.AllOrgs})
	}

	oldConfig, err := a.app.storage.FindConfig(config.Type, config.AppID, config.OrgID)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionFind, model.TypeConfig, nil, err)
	}
	if oldConfig == nil {
		return errors.ErrorData(logutils.StatusMissing, model.TypeConfig, &logutils.FieldArgs{"type": config.Type, "app_id": config.AppID, "org_id": config.OrgID})
	}

	// cannot update a system config if not a system admin
	if !claims.System && oldConfig.System {
		return errors.ErrorData(logutils.StatusInvalid, "system claim", nil)
	}
	err = claims.CanAccess(config.AppID, config.OrgID, config.System)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionValidate, "config access", nil, err)
	}

	now := time.Now().UTC()
	config.ID = oldConfig.ID
	config.DateUpdated = &now

	err = a.app.storage.UpdateConfig(config)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionUpdate, model.TypeConfig, nil, err)
	}
	return nil
}

func (a appAdmin) DeleteConfig(id string, claims *tokenauth.Claims) error {
	config, err := a.app.storage.FindConfigByID(id)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionFind, model.TypeConfig, nil, err)
	}
	if config == nil {
		return errors.ErrorData(logutils.StatusMissing, model.TypeConfig, &logutils.FieldArgs{"id": id})
	}

	err = claims.CanAccess(config.AppID, config.OrgID, config.System)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionValidate, "config access", nil, err)
	}

	err = a.app.storage.DeleteConfig(id)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionDelete, model.TypeConfig, nil, err)
	}
	return nil
}

func (a appAdmin) AddWebtoolsBlackList(ids []string) error {
	blacklist, err := a.app.storage.FindWebtoolsBlacklistData()
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionInsert, model.TypeConfig, nil, err)
	}
	existingData := make(map[string]bool)
	for _, u := range blacklist {
		for _, val := range u.Data {
			existingData[val] = true
		}
	}
	for _, val := range ids {
		if !existingData[val] {
			for _, s := range blacklist {
				s.Data = append(s.Data, val)
				existingData[val] = true
			}
		}
	}

	var newData []string

	for key := range existingData {
		newData = append(newData, key)
	}

	err = a.app.storage.UpdateWebtoolsBlacklistData(newData)
	if err != nil {
		return nil
	}

	return nil
}

func (a appAdmin) GetWebtoolsBlackList() ([]model.WebToolsEventID, error) {

	blacklist, err := a.app.storage.FindWebtoolsBlacklistData()
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionInsert, model.TypeConfig, nil, err)
	}
	return blacklist, nil
}

func (a appAdmin) RemoveWebtoolsBlackList(ids []string) error {

	blacklist, err := a.app.storage.FindWebtoolsBlacklistData()
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionInsert, model.TypeConfig, nil, err)
	}

	idMap := make(map[string]bool)
	for _, id := range ids {
		idMap[id] = true
	}

	newData := []string{}
	if ids != nil {
		for _, j := range blacklist {
			for _, d := range j.Data {
				if !idMap[d] {
					newData = append(newData, d)
				}
			}
		}
	} else {
		ids = nil
	}

	err = a.app.storage.UpdateWebtoolsBlacklistData(newData)
	if err != nil {
		return nil
	}

	return nil
}

// newAppAdmin creates new appAdmin
func newAppAdmin(app *Application) appAdmin {
	return appAdmin{app: app}
}
