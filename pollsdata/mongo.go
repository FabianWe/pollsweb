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
	"fmt"
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
	_, insertErr := h.Collection.InsertOne(ctx, periodSettings)
	return objectId, insertErr
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

func (h *MongoPeriodSettingsHandler) GetPeriodByName(ctx context.Context, name string) (*PeriodSettingsModel, error) {
	filter := bson.M{
		"name": name,
	}

	return h.getSingle(ctx, filter, name)
}

func (h *MongoPeriodSettingsHandler) GetPeriodByID(ctx context.Context, id uuid.UUID) (*PeriodSettingsModel, error) {
	filter := bson.M{
		"_id": id,
	}

	return h.getSingle(ctx, filter, id)
}

func (h *MongoPeriodSettingsHandler) GetPeriodBySlug(ctx context.Context, slug string) (*PeriodSettingsModel, error) {
	filter := bson.M{
		"slug": slug,
	}

	return h.getSingle(ctx, filter, slug)
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

func (h *MongoMeetingHandler) InsertMeeting(ctx context.Context, meeting *MeetingModel) error {
	_, insertErr := h.Collection.InsertOne(ctx, meeting)
	return insertErr
}

func mongoParsePoll(rawDocument bson.Raw) (AbstractPollModel, error) {
	if validationErr := rawDocument.Validate(); validationErr != nil {
		return nil, validationErr
	}
	pollType, lookupErr := rawDocument.LookupErr("type")
	if lookupErr != nil {
		return nil, lookupErr
	}
	pollTypeString, pollTypeStringOk := pollType.StringValueOK()
	if !pollTypeStringOk {
		return nil, errors.New("unable to decode poll type from bson: Not a string")
	}
	var res AbstractPollModel
	switch pollTypeString {
	case BasicPollStringName:
		res = EmptyBasicPollModel()
	case MedianPollStringName:
		res = EmptyMedianPollModel()
	case SchulzePollStringName:
		res = EmptySchulzePollModel()
	default:
		return nil, fmt.Errorf("invalid poll type while parsing poll \"%s\"", pollTypeString)
	}
	if unmarshalErr := bson.Unmarshal(rawDocument, res); unmarshalErr != nil {
		return nil, unmarshalErr
	}
	return res, nil
}

type mongoPollGroupModel struct {
	*IdModel `bson:",inline"`
	Name     string
	Slug     string
	Polls    []bson.Raw
}

func (m *mongoPollGroupModel) parsePolls() ([]AbstractPollModel, error) {
	res := make([]AbstractPollModel, len(m.Polls))
	for i, pollRaw := range m.Polls {
		poll, pollErr := mongoParsePoll(pollRaw)
		if pollErr != nil {
			return nil, fmt.Errorf("unable to decode poll (position %d): %w", i, pollErr)
		}
		res[i] = poll
	}
	return res, nil
}

func (m *mongoPollGroupModel) toPollGroupModel() (*PollGroupModel, error) {
	polls, pollsErr := m.parsePolls()
	if pollsErr != nil {
		return nil, pollsErr
	}
	res := NewPollGroupModel(m.Name, m.Slug, polls)
	// set the id, it's not provided in the constructor
	res.IdModel = m.IdModel
	return res, nil
}

type mongoMeetingModel struct {
	*IdModel    `bson:",inline"`
	Name        string
	Slug        string
	Created     time.Time
	Period      string
	MeetingTime time.Time
	OnlineStart time.Time
	OnlineEnd   time.Time
	Voters      []*VoterModel
	Groups      []*mongoPollGroupModel
}

func emptyMongoMeetingModel() *mongoMeetingModel {
	return &mongoMeetingModel{
		IdModel:     EmptyIdModel(),
		Name:        "",
		Slug:        "",
		Created:     time.Time{},
		Period:      "",
		MeetingTime: time.Time{},
		OnlineStart: time.Time{},
		OnlineEnd:   time.Time{},
		Voters:      nil,
		Groups:      nil,
	}
}

func (m *mongoMeetingModel) parseGroups() ([]*PollGroupModel, error) {
	res := make([]*PollGroupModel, len(m.Groups))
	for i, internalGroup := range m.Groups {
		groupModel, groupErr := internalGroup.toPollGroupModel()
		if groupErr != nil {
			return nil, fmt.Errorf("unable to decode group (position %d): %w", i, groupErr)
		}
		res[i] = groupModel
	}
	return res, nil
}

func (m *mongoMeetingModel) toMeetingModel() (*MeetingModel, error) {
	groups, groupsErr := m.parseGroups()
	if groupsErr != nil {
		return nil, groupsErr
	}
	// create new instance with the given values
	res := NewMeetingModel(m.Name, m.Slug, m.Period, m.MeetingTime, m.OnlineStart, m.OnlineEnd,
		m.Voters, groups)
	// set the id (not provided in the constructor)
	res.IdModel = m.IdModel
	return res, nil
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

func (h *MongoMeetingHandler) GetMeetingBySlug(ctx context.Context, slug string) (*MeetingModel, error) {
	filter := bson.M{
		"slug": slug,
	}

	return h.getSingle(ctx, filter, slug)
}

func (h *MongoMeetingHandler) GetMeetingById(ctx context.Context, id uuid.UUID) (*MeetingModel, error) {
	filter := bson.M{
		"_id": id,
	}

	return h.getSingle(ctx, filter, id)
}
