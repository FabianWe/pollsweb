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

func (h *MongoPeriodSettingsHandler) CreateIndexes(ctx context.Context) ([]string, error) {
	indexes := []mongo.IndexModel{h.endStartIndex(), h.nameIndex(), h.slugIndex(), h.createdIndex()}
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

func (h *MongoPeriodSettingsHandler) slugIndex() mongo.IndexModel {
	return mongo.IndexModel{
		Keys: bson.D{
			{"slug", 1},
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
	_, insertErr := h.Collection.InsertOne(ctx, periodSettings)
	return objectId, insertErr
}

func (h *MongoPeriodSettingsHandler) generateFilter(args *PeriodSettingsQueryArgs) (bson.M, error) {
	res := make(bson.M, 1)
	if args.Id != nil {
		res["_id"] = *args.Id
	}
	if args.Slug != nil {
		res["slug"] = *args.Slug
	}
	if args.Name != nil {
		res["name"] = *args.Slug
	}
	if len(res) == 0 {
		return nil, ErrInvalidPeriodSettingsQuery
	}
	return res, nil
}

func (h *MongoPeriodSettingsHandler) getSingle(ctx context.Context, filter, key interface{}) (*PeriodSettingsModel, error) {
	modelInstance := EmptyPeriodSettingsModel()
	err := h.Collection.FindOne(ctx, filter).Decode(modelInstance)
	if err != nil {
		// check if it is ErrNoDocuments, if so return a not found error
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, NewEntryNotFoundError(periodSettingsModelType, reflect.ValueOf(key), err)
		}
		return nil, err
	}
	return modelInstance, nil
}

func (h *MongoPeriodSettingsHandler) GetPeriod(ctx context.Context, args *PeriodSettingsQueryArgs) (*PeriodSettingsModel, error) {
	filter, queryErr := h.generateFilter(args)
	if queryErr != nil {
		return nil, queryErr
	}
	return h.getSingle(ctx, filter, args)
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

func (h *MongoPeriodSettingsHandler) deleteOnePeriod(ctx context.Context, filter interface{}) (int64, error) {
	deleteRes, deleteErr := h.Collection.DeleteOne(ctx, filter, options.Delete())
	if deleteErr != nil {
		return -1, deleteErr
	}
	return deleteRes.DeletedCount, nil
}

func (h *MongoPeriodSettingsHandler) DeletePeriod(ctx context.Context, args *PeriodSettingsQueryArgs) (int64, error) {
	filter, queryErr := h.generateFilter(args)
	if queryErr != nil {
		return -1, queryErr
	}
	return h.deleteOnePeriod(ctx, filter)
}

type MongoMeetingHandler struct {
	Collection *mongo.Collection
}

func NewMongoMeetingHandler(collection *mongo.Collection) *MongoMeetingHandler {
	return &MongoMeetingHandler{
		Collection: collection,
	}
}

func (h *MongoMeetingHandler) CreateIndexes(ctx context.Context) ([]string, error) {
	indexes := []mongo.IndexModel{
		h.nameIndex(),
		h.slugIndex(),
		h.createdIndex(),
		h.periodIndex(),
		h.meetingTimeIndex(),
		h.onlineVoteIndex(),
		h.voterNameIndex(),
		h.groupNameIndex(),
		h.pollNameIndex(),
	}
	return h.Collection.Indexes().CreateMany(ctx, indexes, options.CreateIndexes())
}

func (h *MongoMeetingHandler) voterNameIndex() mongo.IndexModel {
	return mongo.IndexModel{
		Keys: bson.D{
			{"voters.name", 1},
		},
		Options: options.Index(),
	}
}

func (h *MongoMeetingHandler) groupNameIndex() mongo.IndexModel {
	return mongo.IndexModel{
		Keys: bson.D{
			{"groups.name", 1},
		},
		Options: options.Index(),
	}
}

func (h *MongoMeetingHandler) pollNameIndex() mongo.IndexModel {
	return mongo.IndexModel{
		Keys: bson.D{
			{"groups.poll.name", 1},
		},
		Options: options.Index(),
	}
}

func (h *MongoMeetingHandler) nameIndex() mongo.IndexModel {
	return mongo.IndexModel{
		Keys: bson.D{
			{"name", 1},
		},
		Options: options.Index().SetUnique(true),
	}
}

func (h *MongoMeetingHandler) slugIndex() mongo.IndexModel {
	return mongo.IndexModel{
		Keys: bson.D{
			{"slug", 1},
		},
		Options: options.Index().SetUnique(true),
	}
}

func (h *MongoMeetingHandler) createdIndex() mongo.IndexModel {
	return mongo.IndexModel{
		Keys: bson.D{
			{"created", -1},
		},
		Options: options.Index(),
	}
}

func (h *MongoMeetingHandler) periodIndex() mongo.IndexModel {
	return mongo.IndexModel{
		Keys: bson.D{
			{"period", 1},
		},
		Options: options.Index(),
	}
}

func (h *MongoMeetingHandler) meetingTimeIndex() mongo.IndexModel {
	return mongo.IndexModel{
		Keys: bson.D{
			{"meetingtime", -1},
		},
		Options: options.Index(),
	}
}

func (h *MongoMeetingHandler) onlineVoteIndex() mongo.IndexModel {
	return mongo.IndexModel{
		Keys: bson.D{
			{"onlineend", -1},
			{"onlinestart", -1},
		},
		Options: options.Index(),
	}
}

func (h *MongoMeetingHandler) InsertMeeting(ctx context.Context, meeting *MeetingModel) error {
	_, insertErr := h.Collection.InsertOne(ctx, meeting)
	return insertErr
}

func (h *MongoMeetingHandler) getSingle(ctx context.Context, filter, key interface{}) (*MeetingModel, error) {
	internalModel := emptyMongoMeetingModel()
	err := h.Collection.FindOne(ctx, filter).Decode(internalModel)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, NewEntryNotFoundError(meetingModelType, reflect.ValueOf(key), err)
		}
		return nil, err
	}
	return internalModel.toMeetingModel()
}

func (h *MongoMeetingHandler) generateFilter(args *MeetingQueryArgs) (bson.M, error) {
	res := make(bson.M, 3)
	if args.Id != nil {
		res["_id"] = *args.Id
	}
	if args.Slug != nil {
		res["slug"] = *args.Slug
	}
	if args.Name != nil {
		res["name"] = *args.Slug
	}
	if len(res) == 0 {
		return nil, ErrInvalidMeetingQuery
	}
	// check for optional args
	if args.LastUpdated != nil {
		res["lastupdated"] = *args.LastUpdated
	}
	if args.UpdateToken != nil {
		res["updatetoken"] = *args.UpdateToken
	}
	return res, nil
}

func (h *MongoMeetingHandler) GetMeeting(ctx context.Context, args *MeetingQueryArgs) (*MeetingModel, error) {
	filter, queryErr := h.generateFilter(args)
	if queryErr != nil {
		return nil, queryErr
	}
	return h.getSingle(ctx, filter, args)
}

func (h *MongoMeetingHandler) deleteOneMeeting(ctx context.Context, filter interface{}) (int64, error) {
	deleteRes, deleteErr := h.Collection.DeleteOne(ctx, filter, options.Delete())
	if deleteErr != nil {
		return -1, deleteErr
	}
	return deleteRes.DeletedCount, nil
}

func (h *MongoMeetingHandler) DeleteMeeting(ctx context.Context, args *MeetingQueryArgs) (int64, error) {
	filter, queryErr := h.generateFilter(args)
	if queryErr != nil {
		return -1, queryErr
	}
	return h.deleteOneMeeting(ctx, filter)
}

type MongoDataHandler struct {
	MongoPeriodSettingsHandler
	MongoMeetingHandler
}
