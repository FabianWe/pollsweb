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
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"time"
)

func mongoDecodePollFromRaw(rawDocument bson.Raw) (AbstractPollModel, error) {
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

func (m *mongoPollGroupModel) decodePolls() ([]AbstractPollModel, error) {
	res := make([]AbstractPollModel, len(m.Polls))
	for i, pollRaw := range m.Polls {
		poll, pollErr := mongoDecodePollFromRaw(pollRaw)
		if pollErr != nil {
			return nil, fmt.Errorf("unable to decode poll (position %d): %w", i, pollErr)
		}
		res[i] = poll
	}
	return res, nil
}

func (m *mongoPollGroupModel) toPollGroupModel() (*PollGroupModel, error) {
	polls, pollsErr := m.decodePolls()
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
	LastUpdated time.Time
	UpdateToken int64
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
		LastUpdated: time.Time{},
		// no need to create a random here
		UpdateToken: -1,
	}
}

func (m *mongoMeetingModel) decodeGroups() ([]*PollGroupModel, error) {
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
	groups, groupsErr := m.decodeGroups()
	if groupsErr != nil {
		return nil, groupsErr
	}
	// create new instance with the given values
	res := NewMeetingModel(m.Name, m.Slug, m.Period, m.MeetingTime, m.OnlineStart, m.OnlineEnd,
		m.Voters, groups)
	// set the id (not provided in the constructor)
	res.IdModel = m.IdModel
	// also set last updated and update token
	res.LastUpdated = m.LastUpdated
	res.UpdateToken = m.UpdateToken
	return res, nil
}
