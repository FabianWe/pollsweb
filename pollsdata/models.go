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
	"fmt"
	"github.com/FabianWe/gopolls"
	"github.com/FabianWe/pollsweb"
	"github.com/google/uuid"
	"time"
)

type MeetingTimeTemplateModel struct {
	Weekday time.Weekday
	Hour    uint8
	Minute  uint8
}

func EmptyMeetingTimeTemplateModel() *MeetingTimeTemplateModel {
	return &MeetingTimeTemplateModel{
		Weekday: 0,
		Hour:    0,
		Minute:  0,
	}
}

func NewMeetingTimeTemplateModel(weekday time.Weekday, hour, minute uint8) *MeetingTimeTemplateModel {
	return &MeetingTimeTemplateModel{
		Weekday: weekday,
		Hour:    hour,
		Minute:  minute,
	}
}

func (m *MeetingTimeTemplateModel) String() string {
	return fmt.Sprintf("MeetingTimeTemplateModel(Weekday=%d, Hour=%d, Minute=%d)", m.Weekday, m.Hour, m.Minute)
}

type PeriodSettingsModel struct {
	Id                  uuid.UUID `bson:"_id"`
	Name                string
	MeetingDateTemplate *MeetingTimeTemplateModel `bson:"time"`
	Start               time.Time
	End                 time.Time
	Created             time.Time
}

func EmptyPeriodSettingsModel() *PeriodSettingsModel {
	return &PeriodSettingsModel{
		Id:                  uuid.Nil,
		Name:                "",
		MeetingDateTemplate: EmptyMeetingTimeTemplateModel(),
		Start:               time.Time{},
		End:                 time.Time{},
		Created:             time.Time{},
	}
}

func NewPeriodSettingsModel(name string, meetingDateTemplate *MeetingTimeTemplateModel, start, end time.Time) *PeriodSettingsModel {
	now := pollsweb.UTCNow()
	return &PeriodSettingsModel{
		Id:                  uuid.Nil,
		Name:                name,
		MeetingDateTemplate: meetingDateTemplate,
		Start:               start,
		End:                 end,
		Created:             now,
	}
}

func (m *PeriodSettingsModel) String() string {
	return fmt.Sprintf("PeriodSettingsModel(Id=%s, Name=%s, MettingDateTemplate=%s, Start=%s, End=%s, Created=%s)",
		m.Id, m.Name, m.MeetingDateTemplate, m.Start, m.End, m.Created)
}

type VoterModel struct {
	Id     uuid.UUID
	Name   string
	Weight gopolls.Weight
}

func EmptyVoterModel() *VoterModel {
	return &VoterModel{
		Id:     uuid.Nil,
		Name:   "",
		Weight: gopolls.NoWeight,
	}
}

func NewVoterModel(name string, weight gopolls.Weight) *VoterModel {
	return &VoterModel{
		Id:     uuid.Nil,
		Name:   name,
		Weight: weight,
	}
}

func (m *VoterModel) String() string {
	return fmt.Sprintf("VoterModel(Id=%s, Name=%s, Weight=%d)",
		m.Id, m.Name, m.Weight)
}

type MajorityModel struct {
	Numerator   int64
	Denominator int64
}

func EmptyMajorityModel() *MajorityModel {
	return &MajorityModel{
		Numerator:   -1,
		Denominator: -1,
	}
}

func NewMajorityModel(numerator, denominator int64) *MajorityModel {
	return &MajorityModel{
		Numerator:   numerator,
		Denominator: denominator,
	}
}

func (m *MajorityModel) String() string {
	return fmt.Sprintf("MajorityModel(Numerator=%d, Denominator=%d)",
		m.Numerator, m.Denominator)
}

type AbstractVoteModel interface {
	ModelVoteForType() string
}

type VoteModel struct {
	Id        uuid.UUID
	VoterName string
}

func EmptyVoteModel() *VoteModel {
	return &VoteModel{
		Id:        uuid.Nil,
		VoterName: "",
	}
}

func NewVoteModel(name string) *VoteModel {
	return &VoteModel{
		Id:        uuid.Nil,
		VoterName: name,
	}
}

type BasicPollVoteModel struct {
	*VoteModel
	Answer gopolls.BasicPollAnswer
}

func EmptyBasicPollVoteModel() *BasicPollVoteModel {
	return &BasicPollVoteModel{
		VoteModel: EmptyVoteModel(),
		Answer:    -1,
	}
}

func NewBasicPollVoteModel(name string, answer gopolls.BasicPollAnswer) *BasicPollVoteModel {
	return &BasicPollVoteModel{
		VoteModel: NewVoteModel(name),
		Answer:    answer,
	}
}

func (vote *BasicPollVoteModel) ModelVoteForType() string {
	return "basic"
}

func (vote *BasicPollVoteModel) String() string {
	return fmt.Sprintf("BasicPollVoteModel(VoteModel=%s, Answer=%s)",
		vote.VoteModel, vote.Answer)
}

type MedianPollVoteModel struct {
	*VoteModel
	Value gopolls.MedianUnit
}

func EmptyMedianPollVoteModel() *MedianPollVoteModel {
	return &MedianPollVoteModel{
		VoteModel: EmptyVoteModel(),
		Value:     gopolls.NoMedianUnitValue,
	}
}

func NewMedianPollVoteModel(name string, value gopolls.MedianUnit) *MedianPollVoteModel {
	return &MedianPollVoteModel{
		VoteModel: NewVoteModel(name),
		Value:     value,
	}
}

func (vote *MedianPollVoteModel) ModelVoteForType() string {
	return "median"
}

func (vote *MedianPollVoteModel) String() string {
	return fmt.Sprintf("MedianPollVoteModel(VoteModel=%s, Value=%d)",
		vote.VoteModel, vote.Value)
}

type SchulzePollVoteModel struct {
	*VoteModel
	Ranking gopolls.SchulzeRanking
}

func EmptySchulzePollVoteModel() *SchulzePollVoteModel {
	return &SchulzePollVoteModel{
		VoteModel: EmptyVoteModel(),
		Ranking:   nil,
	}
}

func NewSchulzePollVoteModel(name string, ranking gopolls.SchulzeRanking) *SchulzePollVoteModel {
	return &SchulzePollVoteModel{
		VoteModel: NewVoteModel(name),
		Ranking:   ranking,
	}
}

func (vote *SchulzePollVoteModel) ModelVoteForType() string {
	return "schulze"
}

func (vote *SchulzePollVoteModel) String() string {
	return fmt.Sprintf("SchulzePollVoteModel(VoteModel=%s, Ranking=%v)",
		vote.VoteModel, vote.Ranking)
}

type AbstractPollModel interface {
	ModelPollForType() string
}

type PollModel struct {
	Id               uuid.UUID
	Name             string
	Slug             string
	Majority         *MajorityModel
	AbsoluteMajority bool
	Type             string
}

func EmptyPollModel() *PollModel {
	return &PollModel{
		Id:               uuid.Nil,
		Name:             "",
		Slug:             "",
		Majority:         EmptyMajorityModel(),
		AbsoluteMajority: false,
		Type:             "",
	}
}

func NewPollModel(name, slug string, majority *MajorityModel, absoluteMajority bool, _type string) *PollModel {
	return &PollModel{
		Id:               uuid.Nil,
		Name:             name,
		Slug:             slug,
		Majority:         majority,
		AbsoluteMajority: absoluteMajority,
		Type:             _type,
	}
}

func (poll *PollModel) String() string {
	return fmt.Sprintf("PollModel(Id=%s, Name=%s, Slug=%s, Majority=%s, AbsoluteMajority=%v, Type=%s)",
		poll.Id, poll.Name, poll.Slug, poll.Majority, poll.AbsoluteMajority, poll.Type)
}

type BasicPollModel struct {
	*PollModel
	Votes []*BasicPollVoteModel
}

func EmptyBasicPollModel() *BasicPollModel {
	return &BasicPollModel{
		PollModel: EmptyPollModel(),
		Votes:     nil,
	}
}

func NewBasicPollModel(name, slug string, majority *MajorityModel, absoluteMajority bool, votes []*BasicPollVoteModel) *BasicPollModel {
	return &BasicPollModel{
		PollModel: NewPollModel(name, slug, majority, absoluteMajority, "basic"),
		Votes:     votes,
	}
}

func (poll *BasicPollModel) ModelPollForType() string {
	return "basic"
}

func (poll *BasicPollModel) String() string {
	return fmt.Sprintf("BasicPollModel(PollModel=%s, Votes=%v)",
		poll.PollModel, poll.Votes)
}

type MedianPollModel struct {
	*PollModel
	Value    gopolls.MedianUnit
	Currency string
	Votes    []*MedianPollVoteModel
}

func EmptyMedianPollModel() MedianPollModel {
	return MedianPollModel{
		PollModel: EmptyPollModel(),
		Value:     gopolls.NoMedianUnitValue,
		Currency:  "",
		Votes:     nil,
	}
}

func NewMedianPollModel(name, slug string, majority *MajorityModel, absoluteMajority bool, value gopolls.MedianUnit, currency string, votes []*MedianPollVoteModel) MedianPollModel {
	return MedianPollModel{
		PollModel: NewPollModel(name, slug, majority, absoluteMajority, "median"),
		Value:     value,
		Currency:  currency,
		Votes:     votes,
	}
}

func (poll *MedianPollModel) ModelPollForType() string {
	return "median"
}

func (poll *MedianPollModel) String() string {
	return fmt.Sprintf("MedianPollModel(PollModel=%s, Value=%d, Currency=%s, Votes=%v)",
		poll.PollModel, poll.Value, poll.Currency, poll.Votes)
}

type SchulzePollModel struct {
	*PollModel
	Options []string
	Votes   []*SchulzePollVoteModel
}

func EmptySchulzePollModel() *SchulzePollModel {
	return &SchulzePollModel{
		PollModel: EmptyPollModel(),
		Options:   nil,
		Votes:     nil,
	}
}

func NewSchulzePollModel(name, slug string, majority *MajorityModel, absoluteMajority bool, options []string, votes []*SchulzePollVoteModel) *SchulzePollModel {
	return &SchulzePollModel{
		PollModel: NewPollModel(name, slug, majority, absoluteMajority, "schulze"),
		Options:   options,
		Votes:     votes,
	}
}

func (poll *SchulzePollModel) ModelPollForType() string {
	return "schulze"
}

func (poll *SchulzePollModel) String() string {
	return fmt.Sprintf("SchulzePollModel(PollModel=%s, Options=%v, Votes=%v)",
		poll.PollModel, poll.Options, poll.Votes)
}

type PollGroupModel struct {
	Id    uuid.UUID
	Name  string
	Slug  string
	Polls []AbstractPollModel
}

func EmptyPollGroupModel() *PollGroupModel {
	return &PollGroupModel{
		Id:    uuid.Nil,
		Name:  "",
		Slug:  "",
		Polls: nil,
	}
}

func NewPollGroupModel(name, slug string, polls []AbstractPollModel) *PollGroupModel {
	return &PollGroupModel{
		Id:    uuid.Nil,
		Name:  name,
		Slug:  slug,
		Polls: polls,
	}
}

func (group *PollGroupModel) String() string {
	return fmt.Sprintf("PollGroupModel(Id=%s, Name=%s, Slug=%s, Polls=%v)",
		group.Id, group.Name, group.Slug, group.Polls)
}

type MeetingModel struct {
	Id          uuid.UUID `bson:"_id"`
	Name        string
	Slug        string
	Created     time.Time
	Period      string
	MeetingTime time.Time
	OnlineStart time.Time
	OnlineEnd   time.Time
	Voters      []*VoterModel
	Groups      []*PollGroupModel
}

func EmptyMeetingModel() *MeetingModel {
	now := pollsweb.UTCNow()
	return &MeetingModel{
		Id:          uuid.Nil,
		Name:        "",
		Slug:        "",
		Created:     now,
		Period:      "",
		MeetingTime: time.Time{},
		OnlineStart: time.Time{},
		OnlineEnd:   time.Time{},
		Voters:      nil,
		Groups:      nil,
	}
}

func NewMeetingModel(name, slug string, created time.Time, period string, meetingTime, onlineStart, onlineEnd time.Time, voters []*VoterModel, groups []*PollGroupModel) *MeetingModel {
	return &MeetingModel{
		Id:          uuid.Nil,
		Name:        name,
		Slug:        slug,
		Created:     created,
		Period:      period,
		MeetingTime: meetingTime,
		OnlineStart: onlineStart,
		OnlineEnd:   onlineEnd,
		Voters:      voters,
		Groups:      groups,
	}
}

func (meeting *MeetingModel) String() string {
	return fmt.Sprintf("MeetingModel(Id=%s, Name=%s, Slug=%s, Created=%s, Period=%s, MeetingTime=%s, OnlineStart=%s, OnlineEnd=%s, Voters=%v, Groups=%v)",
		meeting.Id, meeting.Name, meeting.Slug, meeting.Created, meeting.Period, meeting.MeetingTime,
		meeting.OnlineStart, meeting.OnlineEnd, meeting.Voters, meeting.Groups)
}
