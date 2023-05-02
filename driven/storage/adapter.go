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
	"application/core/interfaces"
	"application/core/model"
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/rokwire/logging-library-go/v2/errors"
	"github.com/rokwire/logging-library-go/v2/logs"
	"github.com/rokwire/logging-library-go/v2/logutils"
	"golang.org/x/sync/syncmap"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// Adapter implements the Storage interface
type Adapter struct {
	db *database

	context mongo.SessionContext

	cachedConfigs *syncmap.Map
	configsLock   *sync.RWMutex
}

// Start starts the storage
func (a *Adapter) Start() error {
	err := a.db.start()
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionInitialize, "storage adapter", nil, err)
	}

	//register storage listener
	sl := storageListener{adapter: a}
	a.RegisterStorageListener(&sl)

	//cache the configs
	err = a.cacheConfigs()
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionCache, model.TypeConfig, nil, err)
	}

	return nil
}

// RegisterStorageListener registers a data change listener with the storage adapter
func (a *Adapter) RegisterStorageListener(listener interfaces.StorageListener) {
	a.db.listeners = append(a.db.listeners, listener)
}

// Creates a new Adapter with provided context
func (a *Adapter) withContext(context mongo.SessionContext) *Adapter {
	return &Adapter{db: a.db, context: context, cachedConfigs: a.cachedConfigs, configsLock: a.configsLock}
}

// cacheConfigs caches the configs from the DB
func (a *Adapter) cacheConfigs() error {
	a.db.logger.Info("cacheConfigs...")

	configs, err := a.loadConfigs()
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionLoad, model.TypeConfig, nil, err)
	}

	a.setCachedConfigs(configs)

	return nil
}

func (a *Adapter) setCachedConfigs(configs []model.Config) {
	a.configsLock.Lock()
	defer a.configsLock.Unlock()

	a.cachedConfigs = &syncmap.Map{}

	for _, config := range configs {
		var err error
		switch config.Type {
		case model.ConfigTypeEnv:
			err = parseConfigsData[model.EnvConfigData](&config)
		default:
			err = parseConfigsData[map[string]interface{}](&config)
		}
		if err != nil {
			a.db.logger.Warn(err.Error())
		}
		a.cachedConfigs.Store(config.ID, config)
		a.cachedConfigs.Store(fmt.Sprintf("%s_%s_%s", config.Type, config.AppID, config.OrgID), config)
	}
}

func parseConfigsData[T model.ConfigData](config *model.Config) error {
	bsonBytes, err := bson.Marshal(config.Data)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionUnmarshal, model.TypeConfig, nil, err)
	}

	var data T
	err = bson.Unmarshal(bsonBytes, &data)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionUnmarshal, model.TypeConfigData, &logutils.FieldArgs{"type": config.Type}, err)
	}

	config.Data = data
	return nil
}

func (a *Adapter) getCachedConfig(id string, configType string, appID string, orgID string) (*model.Config, error) {
	a.configsLock.RLock()
	defer a.configsLock.RUnlock()

	var item any
	var errArgs logutils.FieldArgs
	if id != "" {
		errArgs = logutils.FieldArgs{"id": id}
		item, _ = a.cachedConfigs.Load(id)
	} else {
		errArgs = logutils.FieldArgs{"type": configType, "app_id": appID, "org_id": orgID}
		item, _ = a.cachedConfigs.Load(fmt.Sprintf("%s_%s_%s", configType, appID, orgID))
	}

	if item != nil {
		config, ok := item.(model.Config)
		if !ok {
			return nil, errors.ErrorAction(logutils.ActionCast, model.TypeConfig, &errArgs)
		}
		return &config, nil
	}
	return nil, nil
}

func (a *Adapter) getCachedConfigs(configType *string) ([]model.Config, error) {
	a.configsLock.RLock()
	defer a.configsLock.RUnlock()

	var err error
	configList := make([]model.Config, 0)
	a.cachedConfigs.Range(func(key, item interface{}) bool {
		keyStr, ok := key.(string)
		if !ok || item == nil {
			return false
		}
		if !strings.Contains(keyStr, "_") {
			return true
		}

		config, ok := item.(model.Config)
		if !ok {
			err = errors.ErrorAction(logutils.ActionCast, model.TypeConfig, &logutils.FieldArgs{"key": key})
			return false
		}

		if configType == nil || strings.HasPrefix(keyStr, fmt.Sprintf("%s_", *configType)) {
			configList = append(configList, config)
		}

		return true
	})

	return configList, err
}

// loadConfigs loads configs
func (a *Adapter) loadConfigs() ([]model.Config, error) {
	filter := bson.M{}

	var configs []model.Config
	err := a.db.configs.Find(a.context, filter, &configs, nil)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionFind, model.TypeConfig, nil, err)
	}

	return configs, nil
}

// FindConfig finds the config for the specified type, appID, and orgID
func (a *Adapter) FindConfig(configType string, appID string, orgID string) (*model.Config, error) {
	return a.getCachedConfig("", configType, appID, orgID)
}

// FindConfigByID finds the config for the specified ID
func (a *Adapter) FindConfigByID(id string) (*model.Config, error) {
	return a.getCachedConfig(id, "", "", "")
}

// FindConfigs finds all configs for the specified type
func (a *Adapter) FindConfigs(configType *string) ([]model.Config, error) {
	return a.getCachedConfigs(configType)
}

// InsertConfig inserts a new config
func (a *Adapter) InsertConfig(config model.Config) error {
	_, err := a.db.configs.InsertOne(a.context, config)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionInsert, model.TypeConfig, nil, err)
	}

	return nil
}

// UpdateConfig updates an existing config
func (a *Adapter) UpdateConfig(config model.Config) error {
	filter := bson.M{"_id": config.ID}
	update := bson.D{
		primitive.E{Key: "$set", Value: bson.D{
			primitive.E{Key: "type", Value: config.Type},
			primitive.E{Key: "app_id", Value: config.AppID},
			primitive.E{Key: "org_id", Value: config.OrgID},
			primitive.E{Key: "system", Value: config.System},
			primitive.E{Key: "data", Value: config.Data},
			primitive.E{Key: "date_updated", Value: config.DateUpdated},
		}},
	}
	_, err := a.db.configs.UpdateOne(a.context, filter, update, nil)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionUpdate, model.TypeConfig, &logutils.FieldArgs{"id": config.ID}, err)
	}

	return nil
}

// DeleteConfig deletes a configuration from storage
func (a *Adapter) DeleteConfig(id string) error {
	delFilter := bson.M{"_id": id}
	_, err := a.db.configs.DeleteMany(a.context, delFilter, nil)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionDelete, model.TypeConfig, &logutils.FieldArgs{"id": id}, err)
	}

	return nil
}

// PerformTransaction performs a transaction
func (a *Adapter) PerformTransaction(transaction func(storage interfaces.Storage) error) error {
	// transaction
	callback := func(sessionContext mongo.SessionContext) (interface{}, error) {
		adapter := a.withContext(sessionContext)

		err := transaction(adapter)
		if err != nil {
			if wrappedErr, ok := err.(*errors.Error); ok && wrappedErr.Internal() != nil {
				return nil, wrappedErr.Internal()
			}
			return nil, err
		}

		return nil, nil
	}

	session, err := a.db.dbClient.StartSession()
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionStart, "mongo session", nil, err)
	}
	context := context.Background()
	defer session.EndSession(context)

	_, err = session.WithTransaction(context, callback)
	if err != nil {
		return errors.WrapErrorAction("performing", logutils.TypeTransaction, nil, err)
	}
	return nil
}

func filterArgs(filter bson.M) *logutils.FieldArgs {
	args := logutils.FieldArgs{}
	for k, v := range filter {
		args[k] = v
	}
	return &args
}

// NewStorageAdapter creates a new storage adapter instance
func NewStorageAdapter(mongoDBAuth string, mongoDBName string, mongoTimeout string, logger *logs.Logger) *Adapter {
	timeout, err := strconv.Atoi(mongoTimeout)
	if err != nil {
		logger.Infof("Set default timeout - 2000")
		timeout = 2000
	}

	cachedConfigs := &syncmap.Map{}
	configsLock := &sync.RWMutex{}

	db := &database{mongoDBAuth: mongoDBAuth, mongoDBName: mongoDBName, mongoTimeout: time.Millisecond * time.Duration(timeout), logger: logger}
	return &Adapter{db: db, cachedConfigs: cachedConfigs, configsLock: configsLock}
}
