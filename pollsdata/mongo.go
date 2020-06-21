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
	"fmt"
	"github.com/FabianWe/pollsweb"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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

func (h *MongoPeriodSettingsHandler) Insert(ctx context.Context, periodSettings *PeriodSettingsModel) (uuid.UUID, error) {
	objectId, uuidErr := pollsweb.GenUUID()
	if uuidErr != nil {
		return objectId, uuidErr
	}
	periodSettings.Id = objectId
	bsonObj := bson.M{
		"_id":  periodSettings.Id,
		"name": periodSettings.Name,
		"time": bson.M{
			"weekday": periodSettings.MeetingDateTemplate.Weekday,
			"hour":    periodSettings.MeetingDateTemplate.Hour,
			"minute":  periodSettings.MeetingDateTemplate.Minute,
		},
		"start":   periodSettings.Start,
		"end":     periodSettings.End,
		"created": periodSettings.Created,
	}
	insertRes, insertErr := h.Collection.InsertOne(ctx, bsonObj)
	fmt.Println(insertRes)
	return objectId, insertErr
}

func (h *MongoPeriodSettingsHandler) GetByName(ctx context.Context, name string) (*PeriodSettingsModel, error) {
	filter := bson.M{
		"name": name,
	}
	bsonMap := EmptyPeriodSettingsModel()
	err := h.Collection.FindOne(ctx, filter).Decode(bsonMap)
	if err != nil {
		return nil, err
	}
	fmt.Println(bsonMap)
	return nil, nil
}
