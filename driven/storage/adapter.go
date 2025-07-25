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
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/rokwire/rokwire-building-block-sdk-go/utils/errors"
	"github.com/rokwire/rokwire-building-block-sdk-go/utils/logging/logs"
	"github.com/rokwire/rokwire-building-block-sdk-go/utils/logging/logutils"
	"golang.org/x/sync/syncmap"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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
func (a *Adapter) RegisterStorageListener(listener Listener) {
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
	err := a.db.configs.FindWithContext(a.context, filter, &configs, nil)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionFind, model.TypeConfig, nil, err)
	}

	return configs, nil
}

// FindGlobalConfig finds global config by key
func (a *Adapter) FindGlobalConfig(context TransactionContext, key string) (*model.GlobalConfigEntry, error) {
	var err error

	filter := bson.D{
		bson.E{Key: "key", Value: key},
	}

	var globalConfig model.GlobalConfigEntry
	err = a.db.globalConfigs.FindOneWithContext(context, filter, &globalConfig, nil)
	if err != nil {
		return nil, err
	}

	return &globalConfig, nil
}

// SaveGlobalConfig saves global config
func (a *Adapter) SaveGlobalConfig(context TransactionContext, globalConfig model.GlobalConfigEntry) error {
	filter := bson.D{primitive.E{Key: "_id", Value: globalConfig.ID}}
	err := a.db.globalConfigs.ReplaceOneWithContext(context, filter, globalConfig, nil)
	if err != nil {
		return err
	}
	return nil
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
	_, err := a.db.configs.DeleteManyWithContext(a.context, delFilter, nil)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionDelete, model.TypeConfig, &logutils.FieldArgs{"id": id}, err)
	}

	return nil
}

// FindLegacyEventItems finds legacy events items
func (a *Adapter) FindLegacyEventItems(context TransactionContext, source *string, statuses *[]string, dataSourceEventID *string, calendarID *string, originatingCalendarID *string) ([]model.LegacyEventItem, error) {
	filter := bson.D{}

	//source
	if source != nil {
		filter = append(filter, primitive.E{Key: "sync_process_source", Value: *source})
	}

	//statuses
	if statuses != nil {
		filter = append(filter, primitive.E{Key: "status.name", Value: primitive.M{"$in": *statuses}})
	}

	//dataSourceEventID
	if dataSourceEventID != nil {
		filter = append(filter, primitive.E{Key: "item.dataSourceEventId", Value: *dataSourceEventID})
	}

	//calendarID
	if calendarID != nil {
		filter = append(filter, primitive.E{Key: "item.calendarId", Value: *calendarID})
	}

	//originatingCalendarID
	if originatingCalendarID != nil {
		filter = append(filter, primitive.E{Key: "item.originatingCalendarId", Value: *originatingCalendarID})
	}

	var data []model.LegacyEventItem
	err := a.db.legacyEvents.FindWithContext(context, filter, &data, nil)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionFind, model.TypeExample, filterArgs(nil), err)
	}

	return data, nil
}

// FindLegacyEventItemsBySourceID finds legacy events items by source id
func (a *Adapter) FindLegacyEventItemsBySourceID(context TransactionContext, sourceID string) ([]model.LegacyEventItem, error) {
	filter := bson.D{primitive.E{Key: "item.sourceId", Value: sourceID}}
	var data []model.LegacyEventItem
	timeout := 15 * time.Second //15 seconds timeout
	err := a.db.legacyEvents.FindWithParams(context, filter, &data, nil, &timeout)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionFind, model.TypeExample, filterArgs(nil), err)
	}

	return data, nil
}

// InsertLegacyEvents inserts legacy events
func (a *Adapter) InsertLegacyEvents(context TransactionContext, items []model.LegacyEventItem) ([]model.LegacyEventItem, error) {

	storageItems := make([]interface{}, len(items))
	for i, p := range items {
		storageItems[i] = p
	}

	timeout := 15 * time.Second //15 seconds timeout
	_, err := a.db.legacyEvents.InsertManyWithParams(context, storageItems, nil, &timeout)
	if err != nil {
		return nil, errors.WrapErrorAction("insert", "legacy events", nil, err)
	}

	return nil, nil
}

// DeleteLegacyEvents Deletes a reminder
func (a *Adapter) DeleteLegacyEvents() error {
	filter := bson.M{}
	_, err := a.db.legacyEvents.DeleteManyWithContext(nil, filter, nil)
	return err
}

// DeleteLegacyEventsByIDs deletes all items by dataSourceEventIds ????
func (a *Adapter) DeleteLegacyEventsByIDs(context TransactionContext, Ids map[string]string) error {

	var valueIds []string
	for _, value := range Ids {
		valueIds = append(valueIds, value)
	}

	filter := bson.D{
		primitive.E{Key: "item.id", Value: primitive.M{"$in": valueIds}},
	}
	timeout := 15 * time.Second //15 seconds timeout
	_, err := a.db.legacyEvents.DeleteManyWithParams(context, filter, nil, &timeout)
	return err
}

// DeleteLegacyEventsBySourceID deletes all legacy events by source id
func (a *Adapter) DeleteLegacyEventsBySourceID(context TransactionContext, sourceID string) error {
	filter := bson.D{
		primitive.E{Key: "item.sourceId", Value: sourceID},
	}
	timeout := 15 * time.Second //15 seconds timeout
	_, err := a.db.legacyEvents.DeleteManyWithParams(context, filter, nil, &timeout)
	return err
}

// DeleteLegacyEventsByIDsAndCreator deletes legacy events by ids and creator
func (a *Adapter) DeleteLegacyEventsByIDsAndCreator(context TransactionContext, ids []string, accountID string) error {
	var valueIds []string
	for _, value := range ids {
		valueIds = append(valueIds, value)
	}

	filter := bson.D{
		primitive.E{Key: "sync_process_source", Value: "events-tps-api"},
		primitive.E{Key: "create_info.account_id", Value: accountID},
	}

	if ids != nil {
		filter = append(filter, primitive.E{Key: "item.id", Value: primitive.M{"$in": valueIds}})
	}

	_, err := a.db.legacyEvents.DeleteManyWithContext(context, filter, nil)
	return err
}

// FindLegacyEvents finds legacy events by params
func (a *Adapter) FindLegacyEvents(source *string, status *string) ([]model.LegacyEvent, error) {
	filter := bson.D{}

	//source
	if source != nil {
		filter = append(filter, primitive.E{Key: "sync_process_source", Value: *source})
	}

	//status
	if status != nil {
		filter = append(filter, primitive.E{Key: "status.name", Value: *status})
	}

	var list []model.LegacyEventItem
	timeout := 15 * time.Second //15 seconds timeout
	err := a.db.legacyEvents.FindWithParams(nil, filter, &list, nil, &timeout)
	if err != nil {
		return nil, err
	}

	//this processing should happen in the core module, not here
	var legacyEvents []model.LegacyEvent
	for _, l := range list {
		le := l.Item

		legacyEvents = append(legacyEvents, le)
	}

	return legacyEvents, err
}

// AddWebtoolsBlacklistData update data from the database
func (a *Adapter) AddWebtoolsBlacklistData(dataSourceIDs []string, dataCalendarIDs []string, dataOriginatingCalendarIDs []string) error {
	if dataSourceIDs != nil {
		filterSource := bson.M{"name": "webtools_events_ids"}
		updateSource := bson.M{
			"$addToSet": bson.M{
				"data": bson.M{"$each": dataSourceIDs},
			},
		}

		_, err := a.db.webtoolsBlacklistItems.UpdateOne(a.context, filterSource, updateSource, nil)
		if err != nil {
			return errors.WrapErrorAction(logutils.ActionUpdate, "", filterArgs(filterSource), err)
		}
	}
	if dataCalendarIDs != nil {
		filterCalendar := bson.M{"name": "webtools_calendar_ids"}
		updateCalendar := bson.M{
			"$addToSet": bson.M{
				"data": bson.M{"$each": dataCalendarIDs},
			},
		}

		_, err := a.db.webtoolsBlacklistItems.UpdateOne(a.context, filterCalendar, updateCalendar, nil)
		if err != nil {
			return errors.WrapErrorAction(logutils.ActionUpdate, "", filterArgs(filterCalendar), err)
		}
	}

	if dataOriginatingCalendarIDs != nil {
		filterCalendar := bson.M{
			"_id":  "3",
			"name": "webtools_originating_calendar_ids",
		}

		updateCalendar := bson.M{
			"$addToSet": bson.M{
				"data": bson.M{"$each": dataOriginatingCalendarIDs},
			},
		}

		opts := options.Update().SetUpsert(true) //create webtools_originating_calendar_ids if it does not exist

		_, err := a.db.webtoolsBlacklistItems.UpdateOne(a.context, filterCalendar, updateCalendar, opts)
		if err != nil {
			return errors.WrapErrorAction(logutils.ActionUpdate, "", filterArgs(filterCalendar), err)
		}

	}

	return nil

}

// RemoveWebtoolsBlacklistData update data from the database
func (a *Adapter) RemoveWebtoolsBlacklistData(dataSourceIDs []string, dataCalendarIDs []string, dataOriginatingCalendarIdsList []string) error {
	if dataSourceIDs != nil {
		filterSource := bson.M{"name": "webtools_events_ids"}
		updateSource := bson.M{
			"$pull": bson.M{
				"data": bson.M{"$in": dataSourceIDs},
			},
		}

		_, err := a.db.webtoolsBlacklistItems.UpdateOne(a.context, filterSource, updateSource, nil)
		if err != nil {
			return errors.WrapErrorAction(logutils.ActionUpdate, model.TypeExample, filterArgs(filterSource), err)
		}
	}
	if dataCalendarIDs != nil {
		filterCalendar := bson.M{"name": "webtools_calendar_ids"}
		updateCalendar := bson.M{
			"$pull": bson.M{
				"data": bson.M{"$in": dataCalendarIDs},
			},
		}

		_, err := a.db.webtoolsBlacklistItems.UpdateOne(a.context, filterCalendar, updateCalendar, nil)
		if err != nil {
			return errors.WrapErrorAction(logutils.ActionUpdate, model.TypeExample, filterArgs(filterCalendar), err)
		}

	}

	if dataOriginatingCalendarIdsList != nil {
		filterCalendar := bson.M{"name": "webtools_originating_calendar_ids"}
		updateCalendar := bson.M{
			"$pull": bson.M{
				"data": bson.M{"$in": dataOriginatingCalendarIdsList},
			},
		}

		_, err := a.db.webtoolsBlacklistItems.UpdateOne(a.context, filterCalendar, updateCalendar, nil)
		if err != nil {
			return errors.WrapErrorAction(logutils.ActionUpdate, model.TypeExample, filterArgs(filterCalendar), err)
		}

	}

	return nil

}

// FindWebtoolsBlacklistData finds all webtools blacklist from the database
func (a *Adapter) FindWebtoolsBlacklistData(context TransactionContext) ([]model.Blacklist, error) {
	filterSource := bson.M{}
	var dataSource []model.Blacklist
	err := a.db.webtoolsBlacklistItems.FindWithContext(context, filterSource, &dataSource, nil)
	if err != nil {
		return nil, err
	}

	return dataSource, nil
}

// FindWebtoolsOriginatingCalendarIDsBlacklistData finds all webtools blacklist from the database
func (a *Adapter) FindWebtoolsOriginatingCalendarIDsBlacklistData() ([]model.Blacklist, error) {
	filterSource := bson.M{"name": "webtools_originating_calendar_ids"}
	var dataSource []model.Blacklist
	err := a.db.webtoolsBlacklistItems.FindWithContext(a.context, filterSource, &dataSource, nil)
	if err != nil {
		return nil, err
	}

	return dataSource, nil
}

// PerformTransaction performs a transaction
func (a *Adapter) PerformTransaction(transaction func(context TransactionContext) error, timeoutMilliSeconds int64) error {
	// transaction
	timeout := time.Millisecond * time.Duration(timeoutMilliSeconds)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	opts := &options.SessionOptions{}
	opts.SetDefaultMaxCommitTime(&timeout)

	err := a.db.dbClient.UseSessionWithOptions(ctx, opts, func(sessionContext mongo.SessionContext) error {
		err := sessionContext.StartTransaction()
		if err != nil {
			a.abortTransaction(sessionContext)
			return errors.WrapErrorAction(logutils.ActionStart, logutils.TypeTransaction, nil, err)
		}

		err = transaction(sessionContext)
		if err != nil {
			a.abortTransaction(sessionContext)
			return errors.WrapErrorAction("performing", logutils.TypeTransaction, nil, err)
		}

		err = sessionContext.CommitTransaction(sessionContext)
		if err != nil {
			a.abortTransaction(sessionContext)
			return errors.WrapErrorAction(logutils.ActionCommit, logutils.TypeTransaction, nil, err)
		}
		return nil
	})

	return err
}

func (a *Adapter) abortTransaction(sessionContext mongo.SessionContext) {
	err := sessionContext.AbortTransaction(sessionContext)
	if err != nil {
		log.Printf("error aborting a transaction - %s", err)
	}
}

func filterArgs(filter bson.M) *logutils.FieldArgs {
	args := logutils.FieldArgs{}
	for k, v := range filter {
		args[k] = v
	}
	return &args
}

// FindImageItems finds all images stored in the database
func (a *Adapter) FindImageItems() ([]model.ContentImagesURL, error) {
	filter := bson.M{}
	var data []model.ContentImagesURL
	timeout := 15 * time.Second //15 seconds timeout
	err := a.db.processedImages.FindWithParams(nil, filter, &data, nil, &timeout)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionFind, model.TypeExample, filterArgs(nil), err)
	}

	return data, nil
}

// InsertImageItem insert content image url
func (a *Adapter) InsertImageItem(items model.ContentImagesURL) error {
	_, err := a.db.processedImages.InsertOne(nil, items)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionInsert, model.TypeExample, nil, err)
	}
	return nil
}

// FindLegacyLocationItems finds all legacy locations stored in the database
func (a *Adapter) FindLegacyLocationItems() ([]model.LegacyLocation, error) {
	filter := bson.M{}
	var data []model.LegacyLocation
	timeout := 15 * time.Second //15 seconds timeout
	err := a.db.legacyLocations.FindWithParams(nil, filter, &data, nil, &timeout)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionFind, model.TypeExample, filterArgs(nil), err)
	}

	return data, nil
}

// InsertLegacyLocationItem insertthe location of the event
func (a *Adapter) InsertLegacyLocationItem(items model.LegacyLocation) error {
	_, err := a.db.legacyLocations.InsertOne(nil, items)
	if err != nil {
		return nil
	}
	return nil
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

// Listener represents storage listener
type Listener interface {
	OnConfigsUpdated()
	OnExamplesUpdated()
}

// TransactionContext represents storage transaction interface
type TransactionContext interface {
	mongo.SessionContext
}
