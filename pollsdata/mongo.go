// Copyright 2020 Fabian Wenzelmann <fabianwen@posteo.eu>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package pollsdata

import (
	"context"
	"errors"
	"github.com/FabianWe/pollsweb"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"reflect"
	"time"
)

type MongoPeriodSettingsHandler struct {
	Collection *mongo.Collection
}

func NewMongoMeetingTimeHandler(collection *mongo.Collection) *MongoPeriodSettingsHandler {
	return &MongoPeriodSettingsHandler{
		Collection: collection,
	}
}

func (h *MongoPeriodSettingsHandler) CreateIndices(ctx context.Context) ([]string, error) {
	indexes := []mongo.IndexModel{h.endStartIndex(), h.nameIndex(), h.createdIndex()}
	return h.Collection.Indexes().CreateMany(ctx, indexes, options.CreateIndexes())
}

func (h *MongoPeriodSettingsHandler) endStartIndex() mongo.IndexModel {
	return mongo.IndexModel{
		Keys: bson.D{
			{"end", -1},
			{"start", -1},
		},
		Options: options.Index(),
	}
}

func (h *MongoPeriodSettingsHandler) nameIndex() mongo.IndexModel {
	return mongo.IndexModel{
		Keys: bson.D{
			{"name", 1},
		},
		Options: options.Index().SetUnique(true),
	}
}

func (h *MongoPeriodSettingsHandler) createdIndex() mongo.IndexModel {
	return mongo.IndexModel{
		Keys: bson.D{
			{"created", -1},
		},
		Options: options.Index(),
	}
}

func (h *MongoPeriodSettingsHandler) InsertPeriod(ctx context.Context, periodSettings *PeriodSettingsModel) (uuid.UUID, error) {
	objectId, uuidErr := pollsweb.GenUUID()
	if uuidErr != nil {
		return objectId, uuidErr
	}
	periodSettings.Id = objectId
	//bsonObj := bson.M{
	//	"_id":  periodSettings.Id,
	//	"name": periodSettings.Name,
	//	"time": bson.M{
	//		"weekday": periodSettings.MeetingDateTemplate.Weekday,
	//		"hour":    periodSettings.MeetingDateTemplate.Hour,
	//		"minute":  periodSettings.MeetingDateTemplate.Minute,
	//	},
	//	"start":   periodSettings.Start,
	//	"end":     periodSettings.End,
	//	"created": periodSettings.Created,
	//}
	_, insertErr := h.Collection.InsertOne(ctx, periodSettings)
	return objectId, insertErr
}

func (h *MongoPeriodSettingsHandler) GetPeriodByName(ctx context.Context, name string) (*PeriodSettingsModel, error) {
	filter := bson.M{
		"name": name,
	}
	modelInstance := EmptyPeriodSettingsModel()
	err := h.Collection.FindOne(ctx, filter).Decode(modelInstance)
	if err != nil {
		// check if it is ErrNoDocuments, if so return a not found error
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, NewEntryNotFoundError(reflect.TypeOf(modelInstance), reflect.ValueOf(name), err)
		}
		return nil, err
	}
	return modelInstance, nil
}

func (h *MongoPeriodSettingsHandler) GetPeriodByID(ctx context.Context, id uuid.UUID) (*PeriodSettingsModel, error) {
	filter := bson.M{
		"_id": id,
	}
	modelInstance := EmptyPeriodSettingsModel()
	err := h.Collection.FindOne(ctx, filter).Decode(modelInstance)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, NewEntryNotFoundError(reflect.TypeOf(modelInstance), reflect.ValueOf(id), err)
		}
		return nil, err
	}
	return modelInstance, nil
}

func (h *MongoPeriodSettingsHandler) GetActivePeriods(ctx context.Context, referenceTime time.Time) (res []*PeriodSettingsModel, err error) {
	filter := bson.D{
		{"$and", bson.A{
			bson.D{
				{"end", bson.D{
					{"$gte", referenceTime},
				}},
			},
			bson.D{
				{"start", bson.D{
					{"$lte", referenceTime},
				}},
			},
		},
		}}
	cur, curErr := h.Collection.Find(ctx, filter)
	if curErr != nil {
		err = curErr
		return
	}
	// in most cases we expect exactly one entry
	res = make([]*PeriodSettingsModel, 0, 1)
	// takes care of closing the cursor
	defer func() {
		closeErr := cur.Close(ctx)
		// only if no error occurred earlier set err to closeErr
		if err == nil {
			err = closeErr
		}
		// in case of error always set result to nil
		if err != nil {
			res = nil
		}
	}()
	// read entries
	for cur.Next(ctx) {
		next := EmptyPeriodSettingsModel()
		err = cur.Decode(next)
		if err != nil {
			return
		}
		res = append(res, next)
	}
	err = cur.Err()
	return
}

type MongoMeetingHandler struct {
	Collection *mongo.Collection
}

func NewMongoMeetingHandler(collection *mongo.Collection) *MongoMeetingHandler {
	return &MongoMeetingHandler{
		Collection: collection,
	}
}

func (h *MongoMeetingHandler) InsertMeeting(ctx context.Context, meeting *MeetingModel) (uuid.UUID, error) {
	objectId, uuidErr := pollsweb.GenUUID()
	if uuidErr != nil {
		return objectId, uuidErr
	}
	meeting.Id = objectId
	_, insertErr := h.Collection.InsertOne(ctx, meeting)
	return objectId, insertErr
}
