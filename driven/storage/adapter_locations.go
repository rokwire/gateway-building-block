package storage

import (
	"application/core/model"

	"github.com/rokwire/logging-library-go/v2/errors"
	"github.com/rokwire/logging-library-go/v2/logutils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// InitializeLegacyLocations Initialize the locations db only if it's empty
func (a *Adapter) InitializeLegacyLocations() error {

	err := a.PerformTransaction(func(context TransactionContext) error {
		count, err := a.db.legacyLocations.CountDocuments(context, bson.D{})
		if err != nil {
			return err
		}

		if count == 0 {
			_, err := a.db.legacyLocations.InsertManyWithContext(context, model.DefaultLegacyLocations.ToBsonRecords(), nil)
			if err != nil {
				return err
			}
		}
		return nil
	}, 10000)

	if err != nil {
		return errors.WrapErrorAction(logutils.ActionInsert, model.TypeEventLocations, nil, err)
	}

	return nil
}

// FindLegacyLocations finds legacy locations
func (a *Adapter) FindLegacyLocations() (model.LegacyLocationsListType, error) {

	var list model.LegacyLocationsListType
	err := a.db.legacyLocations.FindWithContext(a.context, bson.D{}, &list, options.Find().SetSort(bson.D{{Key: "name", Value: 1}}))
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionFind, model.TypeEventLocations, nil, err)
	}
	return list, nil
}
