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
	"context"
	"time"

	"github.com/rokwire/rokwire-building-block-sdk-go/utils/errors"
	"github.com/rokwire/rokwire-building-block-sdk-go/utils/logging/logs"
	"github.com/rokwire/rokwire-building-block-sdk-go/utils/logging/logutils"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type database struct {
	mongoDBAuth  string
	mongoDBName  string
	mongoTimeout time.Duration

	db       *mongo.Database
	dbClient *mongo.Client
	logger   *logs.Logger

	globalConfigs       *collectionWrapper
	configs             *collectionWrapper
	examples            *collectionWrapper
	unitcalendars       *collectionWrapper
	appbuildingfeatures *collectionWrapper
	floorplanmarkup     *collectionWrapper

	legacyEvents           *collectionWrapper
	legacyLocations        *collectionWrapper
	webtoolsBlacklistItems *collectionWrapper
	processedImages        *collectionWrapper

	listeners []Listener
}

func (d *database) start() error {

	d.logger.Info("database -> start")

	//connect to the database
	clientOptions := options.Client().ApplyURI(d.mongoDBAuth)
	connectContext, cancel := context.WithTimeout(context.Background(), d.mongoTimeout)
	client, err := mongo.Connect(connectContext, clientOptions)
	cancel()
	if err != nil {
		return err
	}

	//ping the database
	pingContext, cancel := context.WithTimeout(context.Background(), d.mongoTimeout)
	err = client.Ping(pingContext, nil)
	cancel()
	if err != nil {
		return err
	}

	//apply checks
	db := client.Database(d.mongoDBName)

	globalConfigs := &collectionWrapper{database: d, coll: db.Collection("global_configs")}
	err = d.applyGlobalConfigsChecks(globalConfigs)
	if err != nil {
		return err
	}

	configs := &collectionWrapper{database: d, coll: db.Collection("configs")}
	err = d.applyConfigsChecks(configs)
	if err != nil {
		return err
	}

	examples := &collectionWrapper{database: d, coll: db.Collection("examples")}
	err = d.applyExamplesChecks(examples)
	if err != nil {
		return err
	}

	legacyEvents := &collectionWrapper{database: d, coll: db.Collection("legacy_events")}
	err = d.applyLegacyEventsChecks(legacyEvents)
	if err != nil {
		return err
	}

	unitcalendars := &collectionWrapper{database: d, coll: db.Collection("unitcalendars")}

	appbuildingfeatures := &collectionWrapper{database: d, coll: db.Collection("building_features")}
	floorplanmarkup := &collectionWrapper{database: d, coll: db.Collection("floorplan_markup")}

	legacyLocations := &collectionWrapper{database: d, coll: db.Collection("legacy_locations")}
	err = d.applyLegacyLocationsChecks(legacyEvents)
	if err != nil {
		return err
	}

	webtoolsBlacklistItems := &collectionWrapper{database: d, coll: db.Collection("webtools_blacklist_items")}
	err = d.applyWebtoolsBlacklistItemsChecks(webtoolsBlacklistItems)
	if err != nil {
		return err
	}

	processedImages := &collectionWrapper{database: d, coll: db.Collection("processed_images")}
	err = d.applyprocessedImagesChecks(processedImages)
	if err != nil {
		return err
	}

	//assign the db, db client and the collections
	d.db = db
	d.dbClient = client

	d.globalConfigs = globalConfigs
	d.configs = configs
	d.examples = examples
	d.legacyEvents = legacyEvents
	d.unitcalendars = unitcalendars
	d.appbuildingfeatures = appbuildingfeatures
	d.floorplanmarkup = floorplanmarkup
	d.legacyLocations = legacyLocations
	d.webtoolsBlacklistItems = webtoolsBlacklistItems
	d.processedImages = processedImages

	go d.configs.Watch(nil, d.logger)

	return nil
}

func (d *database) applyGlobalConfigsChecks(globalConfigs *collectionWrapper) error {
	d.logger.Info("apply global configs checks.....")

	err := globalConfigs.AddIndex(bson.D{primitive.E{Key: "key", Value: 1}}, true)
	if err != nil {
		return err
	}

	d.logger.Info("global configs passed")
	return nil
}

func (d *database) applyConfigsChecks(configs *collectionWrapper) error {
	d.logger.Info("apply configs checks.....")

	err := configs.AddIndex(bson.D{primitive.E{Key: "type", Value: 1}, primitive.E{Key: "app_id", Value: 1}, primitive.E{Key: "org_id", Value: 1}}, true)
	if err != nil {
		return err
	}

	d.logger.Info("apply configs passed")
	return nil
}

func (d *database) applyExamplesChecks(examples *collectionWrapper) error {
	d.logger.Info("apply examples checks.....")

	//add compound unique index - org_id + app_id
	err := examples.AddIndex(bson.D{primitive.E{Key: "org_id", Value: 1}, primitive.E{Key: "app_id", Value: 1}}, false)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionCreate, "index", nil, err)
	}

	d.logger.Info("apply examples passed")
	return nil
}

func (d *database) applyLegacyEventsChecks(legacyEvents *collectionWrapper) error {
	d.logger.Info("apply legacy events checks.....")

	//sync process source
	err := legacyEvents.AddIndex(bson.D{primitive.E{Key: "sync_process_source", Value: 1}}, false)
	if err != nil {
		return err
	}

	//id
	err = legacyEvents.AddIndex(bson.D{primitive.E{Key: "item.id", Value: 1}}, true)
	if err != nil {
		return err
	}

	//source id
	err = legacyEvents.AddIndex(bson.D{primitive.E{Key: "item.sourceId", Value: 1}}, false)
	if err != nil {
		return err
	}

	//calendar id
	err = legacyEvents.AddIndex(bson.D{primitive.E{Key: "item.calendarId", Value: 1}}, false)
	if err != nil {
		return err
	}

	//originating calendar id
	err = legacyEvents.AddIndex(bson.D{primitive.E{Key: "item.originatingCalendarId", Value: 1}}, false)
	if err != nil {
		return err
	}

	//status name
	err = legacyEvents.AddIndex(bson.D{primitive.E{Key: "status.name", Value: 1}}, false)
	if err != nil {
		return err
	}

	d.logger.Info("legacy events passed")
	return nil
}

func (d *database) applyLegacyLocationsChecks(locations *collectionWrapper) error {
	d.logger.Info("apply legacy_locations checks.....")

	err := locations.AddIndex(bson.D{primitive.E{Key: "name", Value: 1}}, false)
	if err != nil {
		return err
	}

	d.logger.Info("legacy legacy_locations passed")
	return nil
}

func (d *database) applyWebtoolsBlacklistItemsChecks(webtoolsBlacklistItems *collectionWrapper) error {
	d.logger.Info("apply webtools_blacklist_items checks.....")

	d.logger.Info("legacy webtools_blacklist_items passed")
	return nil
}

func (d *database) applyprocessedImagesChecks(webtoolsBlacklistItems *collectionWrapper) error {
	d.logger.Info("apply processed_images checks.....")

	d.logger.Info("processed_images passed")
	return nil
}

func (d *database) onDataChanged(changeDoc map[string]interface{}) {
	if changeDoc == nil {
		return
	}
	d.logger.Infof("onDataChanged: %+v\n", changeDoc)
	ns := changeDoc["ns"]
	if ns == nil {
		return
	}
	nsMap := ns.(map[string]interface{})
	coll := nsMap["coll"]

	switch coll {
	case "configs":
		d.logger.Info("configs collection changed")

		for _, listener := range d.listeners {
			go listener.OnConfigsUpdated()
		}
	case "examples":
		d.logger.Info("examples collection changed")

		for _, listener := range d.listeners {
			go listener.OnExamplesUpdated()
		}
	}
}
