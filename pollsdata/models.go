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
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/google/uuid"
	"github.com/hashicorp/go-multierror"
	"math/rand"
	"reflect"
	"time"
)

// TODO remove govalidate and use https://github.com/go-ozzo/ozzo-validation

const (
	BasicPollStringName   = "basic"
	MedianPollStringName  = "median"
	SchulzePollStringName = "schulze"
)

var (
	periodSettingsModelType = reflect.TypeOf(EmptyPeriodSettingsModel())
	meetingModelType        = reflect.TypeOf(EmptyMeetingModel())
)

type AbstractIdModel interface {
	GetId() uuid.UUID
	SetId(id uuid.UUID)
}

type IdModel struct {
	Id uuid.UUID `bson:"_id"`
}

func EmptyIdModel() *IdModel {
	return &IdModel{
		Id: uuid.Nil,
	}
}

func NewIdModel(id uuid.UUID) *IdModel {
	return &IdModel{Id: id}
}

func (m *IdModel) GetId() uuid.UUID {
	return m.Id
}

func (m *IdModel) SetId(id uuid.UUID) {
	m.Id = id
}

// Add more slugs / ids?
// Add set methods for ids

type MeetingTimeTemplateModel struct {
	Weekday time.Weekday // valid with range doesn't work here, tested in ValidateModel
	Hour    uint8        `valid:"range(0|23)"`
	Minute  uint8        `valid:"range(0|59)"`
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

var weekdays = []interface{}{time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday, time.Saturday, time.Sunday}

func (m *MeetingTimeTemplateModel) ValidateModel() error {
	switch m.Weekday {
	case time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday, time.Saturday, time.Sunday:
		return nil
	default:
		return NewModelValidationError("invalid weekday, must be a value between 0 and 6").
			SetFieldName("Weekday")
	}
}

func (m *MeetingTimeTemplateModel) Validate() error {
	return validation.Errors{
		"Weekday": validation.Validate(m.Weekday, validation.In(weekdays...)),
		"Hour":    validation.Validate(m.Hour, validation.Max(uint8(23))),
		"Minute":  validation.Validate(m.Minute, validation.Max(uint8(59))),
	}.Filter()
}

type PeriodSettingsModel struct {
	*IdModel            `bson:",inline"`
	Name                string                    `valid:"-"`
	Slug                string                    `valid:"-"`
	MeetingDateTemplate *MeetingTimeTemplateModel `bson:"time" valid:""`
	Voters              []*VoterModel             // TODO valid?
	Start               time.Time
	End                 time.Time
	Created             time.Time
	LastUpdated         time.Time
}

func (m *PeriodSettingsModel) Validate() error {
	return validation.Errors{
		// TODO insert name etc.
	}.Filter()
}

func EmptyPeriodSettingsModel() *PeriodSettingsModel {
	return &PeriodSettingsModel{
		IdModel:             EmptyIdModel(),
		Name:                "",
		Slug:                "",
		MeetingDateTemplate: EmptyMeetingTimeTemplateModel(),
		Voters:              nil,
		Start:               time.Time{},
		End:                 time.Time{},
		Created:             time.Time{},
		LastUpdated:         time.Time{},
	}
}

// TODO where are ids for voters generated?
func NewPeriodSettingsModel(name, slug string, meetingDateTemplate *MeetingTimeTemplateModel, voters []*VoterModel, start, end time.Time) *PeriodSettingsModel {
	now := pollsweb.UTCNow()
	return &PeriodSettingsModel{
		IdModel:             EmptyIdModel(),
		Name:                name,
		Slug:                slug,
		MeetingDateTemplate: meetingDateTemplate,
		Voters:              voters,
		Start:               start,
		End:                 end,
		Created:             now,
		LastUpdated:         now,
	}
}

func (m *PeriodSettingsModel) String() string {
	return fmt.Sprintf("PeriodSettingsModel(Id=%s, Name=%s, Slug=%s, MettingDateTemplate=%s, Voters=%v, Start=%s, End=%s, Created=%s, LastUpdated=%s)",
		m.Id, m.Name, m.Slug, m.MeetingDateTemplate, m.Voters, m.Start, m.End, m.Created, m.LastUpdated)
}

type PeriodSettingsValidator struct {
	NameMin, NameMax int
	SlugMin, SlugMax int
}

func NewPeriodSettingsValidator() *PeriodSettingsValidator {
	return &PeriodSettingsValidator{
		NameMin: -1,
		NameMax: -1,
		SlugMin: -1,
		SlugMax: -1,
	}
}

// TODO voters?
func (validator *PeriodSettingsValidator) Validator() CustomValidator {
	return func(model interface{}, modelValidator *ModelValidator) error {
		asSettings, isSettings := model.(*PeriodSettingsModel)
		if !isSettings {
			panic(fmt.Sprintf("can't validate model: must be a *PeriodSettingsModel, got type %v", reflect.TypeOf(model)))
		}
		var res *multierror.Error
		if nameValidateErr := runeLengthValidator(asSettings.Name, validator.NameMin, validator.NameMax); nameValidateErr != nil {
			res = multierror.Append(nameValidateErr.SetFieldName("Name"))
		}
		if slugValidationErr := slugValidator(asSettings.Slug); slugValidationErr != nil {
			res = multierror.Append(res, slugValidationErr.SetFieldName("Slug"))
		} else {
			// only test length if it is a valid slug
			if slugLenValidateErr := runeLengthValidator(asSettings.Slug, validator.SlugMin, validator.SlugMax); slugLenValidateErr != nil {
				res = multierror.Append(res, slugLenValidateErr)
			}
		}
		if asSettings.End.Before(asSettings.Start) {
			res = multierror.Append(res, NewModelValidationError("Start date must be before end date").SetFieldName("Start"))
		}
		return res.ErrorOrNil()
	}
}

type VoterModel struct {
	*IdModel `bson:",inline"`
	Name     string
	Slug     string
	Weight   gopolls.Weight
}

func EmptyVoterModel() *VoterModel {
	return &VoterModel{
		IdModel: EmptyIdModel(),
		Name:    "",
		Slug:    "",
		Weight:  gopolls.NoWeight,
	}
}

func NewVoterModel(name, slug string, weight gopolls.Weight) *VoterModel {
	return &VoterModel{
		IdModel: EmptyIdModel(),
		Name:    name,
		Slug:    slug,
		Weight:  weight,
	}
}

func (m *VoterModel) String() string {
	return fmt.Sprintf("VoterModel(Id=%s, Name=%s, Slug=%s, Weight=%d)",
		m.Id, m.Name, m.Slug, m.Weight)
}

type VoterValidator struct {
	NameMin, NameMax int
	SlugMin, SlugMax int
	MaxWeight        gopolls.Weight
}

func NewVoterValidator() *VoterValidator {
	return &VoterValidator{
		NameMin:   -1,
		NameMax:   -1,
		SlugMin:   -1,
		SlugMax:   -1,
		MaxWeight: gopolls.NoWeight,
	}
}

func (validator *VoterValidator) Validator() CustomValidator {
	return func(model interface{}, modelValidator *ModelValidator) error {
		asVoter, isVoter := model.(*VoterModel)
		if !isVoter {
			panic(fmt.Sprintf("can't validate model: must be a *VoterModel, got type %v", reflect.TypeOf(model)))
		}
		var res *multierror.Error
		if nameValidateErr := runeLengthValidator(asVoter.Name, validator.NameMin, validator.NameMax); nameValidateErr != nil {
			res = multierror.Append(nameValidateErr.SetFieldName("Name"))
		}
		if slugValidationErr := slugValidator(asVoter.Slug); slugValidationErr != nil {
			res = multierror.Append(res, slugValidationErr.SetFieldName("Slug"))
		} else {
			// only test length if it is a valid slug
			if slugLenValidateErr := runeLengthValidator(asVoter.Slug, validator.SlugMin, validator.SlugMax); slugLenValidateErr != nil {
				res = multierror.Append(res, slugLenValidateErr)
			}
		}
		if weightValidationErr := weightValidator(asVoter.Weight, validator.MaxWeight); weightValidationErr != nil {
			res = multierror.Append(res, weightValidationErr.SetFieldName("Weight"))
		}
		return res.ErrorOrNil()
	}
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

func (m *MajorityModel) ValidateModel() error {
	var res *multierror.Error
	if positiveErr := strictlyPositiveInt64Validator(m.Denominator); positiveErr != nil {
		res = multierror.Append(res, positiveErr.SetFieldName("Denominator"))
	}
	// test if Numerator > Denominator, this would not be allowed for majorities
	if m.Numerator > m.Denominator {
		res = multierror.Append(res, NewModelValidationError("invalid majority: numerator must be ≤ denominator").SetFieldName("Numerator"))
	}
	return res.ErrorOrNil()
}

type AbstractVoteModel interface {
	AbstractIdModel
	ModelVoteForType() string
}

type VoteModel struct {
	*IdModel  `bson:",inline"`
	VoterName string
	// unique in the poll, so probably just use slug of voter name
	Slug string
}

func EmptyVoteModel() *VoteModel {
	return &VoteModel{
		IdModel:   EmptyIdModel(),
		VoterName: "",
		Slug:      "",
	}
}

func NewVoteModel(name, slug string) *VoteModel {
	return &VoteModel{
		IdModel:   EmptyIdModel(),
		VoterName: name,
		Slug:      slug,
	}
}

func (m *VoteModel) String() string {
	return fmt.Sprintf("VoteModel(Id=%s, VoterName=%s, Slug=%s)",
		m.Id, m.VoterName, m.Slug)
}

type VoteValidator struct {
	NameMin, NameMax int
	SlugMin, SlugMax int
}

func NewVoteValidator() *VoteValidator {
	return &VoteValidator{
		NameMin: -1,
		NameMax: -1,
		SlugMin: -1,
		SlugMax: -1,
	}
}

func (validator *VoteValidator) Validator() CustomValidator {
	return func(model interface{}, modelValidator *ModelValidator) error {
		asVote, isVote := model.(*VoteModel)
		if !isVote {
			panic(fmt.Sprintf("can't validate model: must be a *VoteModel, got type %v", reflect.TypeOf(model)))
		}
		var res *multierror.Error
		if nameValidateErr := runeLengthValidator(asVote.VoterName, validator.NameMin, validator.NameMax); nameValidateErr != nil {
			res = multierror.Append(nameValidateErr.SetFieldName("VoterName"))
		}
		if slugValidationErr := slugValidator(asVote.Slug); slugValidationErr != nil {
			res = multierror.Append(res, slugValidationErr.SetFieldName("Slug"))
		} else {
			// only test length if it is a valid slug
			if slugLenValidateErr := runeLengthValidator(asVote.Slug, validator.SlugMin, validator.SlugMax); slugLenValidateErr != nil {
				res = multierror.Append(res, slugLenValidateErr)
			}
		}
		return res.ErrorOrNil()
	}
}

type BasicPollVoteModel struct {
	*VoteModel `bson:",inline"`
	Answer     gopolls.BasicPollAnswer `valid:"range(0|2)"`
}

func EmptyBasicPollVoteModel() *BasicPollVoteModel {
	return &BasicPollVoteModel{
		VoteModel: EmptyVoteModel(),
		Answer:    -1,
	}
}

func NewBasicPollVoteModel(name, slug string, answer gopolls.BasicPollAnswer) *BasicPollVoteModel {
	return &BasicPollVoteModel{
		VoteModel: NewVoteModel(name, slug),
		Answer:    answer,
	}
}

func (vote *BasicPollVoteModel) ModelVoteForType() string {
	return BasicPollStringName
}

func (vote *BasicPollVoteModel) String() string {
	return fmt.Sprintf("BasicPollVoteModel(VoteModel=%s, Answer=%s)",
		vote.VoteModel, vote.Answer)
}

type MedianPollVoteModel struct {
	*VoteModel `bson:",inline"`
	Value      gopolls.MedianUnit // Note: this is not validated, only the poll is validated and then the answers to the poll
}

func EmptyMedianPollVoteModel() *MedianPollVoteModel {
	return &MedianPollVoteModel{
		VoteModel: EmptyVoteModel(),
		Value:     gopolls.NoMedianUnitValue,
	}
}

func NewMedianPollVoteModel(name, slug string, value gopolls.MedianUnit) *MedianPollVoteModel {
	return &MedianPollVoteModel{
		VoteModel: NewVoteModel(name, slug),
		Value:     value,
	}
}

func (vote *MedianPollVoteModel) ModelVoteForType() string {
	return MedianPollStringName
}

func (vote *MedianPollVoteModel) String() string {
	return fmt.Sprintf("MedianPollVoteModel(VoteModel=%s, Value=%d)",
		vote.VoteModel, vote.Value)
}

type SchulzePollVoteModel struct {
	*VoteModel `bson:",inline"`
	Ranking    gopolls.SchulzeRanking // Note: this is not validated, only the poll is validated and then the answers to the poll
}

func EmptySchulzePollVoteModel() *SchulzePollVoteModel {
	return &SchulzePollVoteModel{
		VoteModel: EmptyVoteModel(),
		Ranking:   nil,
	}
}

func NewSchulzePollVoteModel(name, slug string, ranking gopolls.SchulzeRanking) *SchulzePollVoteModel {
	return &SchulzePollVoteModel{
		VoteModel: NewVoteModel(name, slug),
		Ranking:   ranking,
	}
}

func (vote *SchulzePollVoteModel) ModelVoteForType() string {
	return SchulzePollStringName
}

func (vote *SchulzePollVoteModel) String() string {
	return fmt.Sprintf("SchulzePollVoteModel(VoteModel=%s, Ranking=%v)",
		vote.VoteModel, vote.Ranking)
}

type AbstractPollModel interface {
	AbstractIdModel
	ModelPollForType() string
	// GenId for model itself and also for all votes
	GenIds() error
}

type PollModel struct {
	*IdModel         `bson:",inline"`
	Name             string `valid:"runelength(5|250)"`
	Slug             string `valid:"-"`
	Majority         *MajorityModel
	AbsoluteMajority bool   `valid:"-"`
	Type             string `valid:"in(basic|median|schulze)"`
}

func EmptyPollModel() *PollModel {
	return &PollModel{
		IdModel:          EmptyIdModel(),
		Name:             "",
		Slug:             "",
		Majority:         EmptyMajorityModel(),
		AbsoluteMajority: false,
		Type:             "",
	}
}

func NewPollModel(name, slug string, majority *MajorityModel, absoluteMajority bool, _type string) *PollModel {
	return &PollModel{
		IdModel:          EmptyIdModel(),
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
	*PollModel `bson:",inline"`
	Votes      []*BasicPollVoteModel
}

func EmptyBasicPollModel() *BasicPollModel {
	return &BasicPollModel{
		PollModel: EmptyPollModel(),
		Votes:     nil,
	}
}

func NewBasicPollModel(name, slug string, majority *MajorityModel, absoluteMajority bool, votes []*BasicPollVoteModel) *BasicPollModel {
	return &BasicPollModel{
		PollModel: NewPollModel(name, slug, majority, absoluteMajority, BasicPollStringName),
		Votes:     votes,
	}
}

func (poll *BasicPollModel) ModelPollForType() string {
	return BasicPollStringName
}

func (poll *BasicPollModel) String() string {
	return fmt.Sprintf("BasicPollModel(PollModel=%s, Votes=%v)",
		poll.PollModel, poll.Votes)
}

func (poll *BasicPollModel) GenIds() error {
	// re-use variables
	var genId uuid.UUID
	var genErr error
	genId, genErr = pollsweb.GenUUID()
	if genErr != nil {
		return genErr
	}
	poll.SetId(genId)
	for _, vote := range poll.Votes {
		genId, genErr = pollsweb.GenUUID()
		if genErr != nil {
			return genErr
		}
		vote.SetId(genId)
	}
	return nil
}

type MedianPollModel struct {
	*PollModel `bson:",inline"`
	Value      gopolls.MedianUnit `valid:"range(0|2147483647)"` // TODO custom
	Currency   string
	Votes      []*MedianPollVoteModel
}

func EmptyMedianPollModel() *MedianPollModel {
	return &MedianPollModel{
		PollModel: EmptyPollModel(),
		Value:     gopolls.NoMedianUnitValue,
		Currency:  "",
		Votes:     nil,
	}
}

func NewMedianPollModel(name, slug string, majority *MajorityModel, absoluteMajority bool, value gopolls.MedianUnit, currency string, votes []*MedianPollVoteModel) *MedianPollModel {
	return &MedianPollModel{
		PollModel: NewPollModel(name, slug, majority, absoluteMajority, MedianPollStringName),
		Value:     value,
		Currency:  currency,
		Votes:     votes,
	}
}

func (poll *MedianPollModel) ModelPollForType() string {
	return MedianPollStringName
}

func (poll *MedianPollModel) String() string {
	return fmt.Sprintf("MedianPollModel(PollModel=%s, Value=%d, Currency=%s, Votes=%v)",
		poll.PollModel, poll.Value, poll.Currency, poll.Votes)
}

func (poll *MedianPollModel) GenIds() error {
	// re-use variables
	var genId uuid.UUID
	var genErr error
	genId, genErr = pollsweb.GenUUID()
	if genErr != nil {
		return genErr
	}
	poll.SetId(genId)
	for _, vote := range poll.Votes {
		genId, genErr = pollsweb.GenUUID()
		if genErr != nil {
			return genErr
		}
		vote.SetId(genId)
	}
	return nil
}

type SchulzePollModel struct {
	*PollModel `bson:",inline"`
	Options    []string
	Votes      []*SchulzePollVoteModel
}

// TODO custom

func EmptySchulzePollModel() *SchulzePollModel {
	return &SchulzePollModel{
		PollModel: EmptyPollModel(),
		Options:   nil,
		Votes:     nil,
	}
}

func NewSchulzePollModel(name, slug string, majority *MajorityModel, absoluteMajority bool, options []string, votes []*SchulzePollVoteModel) *SchulzePollModel {
	return &SchulzePollModel{
		PollModel: NewPollModel(name, slug, majority, absoluteMajority, SchulzePollStringName),
		Options:   options,
		Votes:     votes,
	}
}

func (poll *SchulzePollModel) ModelPollForType() string {
	return SchulzePollStringName
}

func (poll *SchulzePollModel) String() string {
	return fmt.Sprintf("SchulzePollModel(PollModel=%s, Options=%v, Votes=%v)",
		poll.PollModel, poll.Options, poll.Votes)
}

func (poll *SchulzePollModel) GenIds() error {
	// re-use variables
	var genId uuid.UUID
	var genErr error
	genId, genErr = pollsweb.GenUUID()
	if genErr != nil {
		return genErr
	}
	poll.SetId(genId)
	for _, vote := range poll.Votes {
		genId, genErr = pollsweb.GenUUID()
		if genErr != nil {
			return genErr
		}
		vote.SetId(genId)
	}
	return nil
}

type PollGroupModel struct {
	*IdModel `bson:",inline"`
	Name     string // TODO custom
	Slug     string
	Polls    []AbstractPollModel
}

func EmptyPollGroupModel() *PollGroupModel {
	return &PollGroupModel{
		IdModel: EmptyIdModel(),
		Name:    "",
		Slug:    "",
		Polls:   nil,
	}
}

func NewPollGroupModel(name, slug string, polls []AbstractPollModel) *PollGroupModel {
	return &PollGroupModel{
		IdModel: EmptyIdModel(),
		Name:    name,
		Slug:    slug,
		Polls:   polls,
	}
}

func (group *PollGroupModel) String() string {
	return fmt.Sprintf("PollGroupModel(Id=%s, Name=%s, Slug=%s, Polls=%v)",
		group.Id, group.Name, group.Slug, group.Polls)
}

func (group *PollGroupModel) GenIds() error {
	var genId uuid.UUID
	var genErr error

	genId, genErr = pollsweb.GenUUID()
	if genErr != nil {
		return genErr
	}
	group.SetId(genId)
	// now for all polls
	for _, poll := range group.Polls {
		genErr = poll.GenIds()
		if genErr != nil {
			return genErr
		}
	}
	return nil
}

type MeetingModel struct {
	*IdModel    `bson:",inline"`
	Name        string
	Slug        string
	Created     time.Time
	Period      string
	MeetingTime time.Time
	OnlineStart time.Time
	OnlineEnd   time.Time
	Voters      []*VoterModel
	Groups      []*PollGroupModel
	LastUpdated time.Time
	UpdateToken int64
}

// TODO custom

func EmptyMeetingModel() *MeetingModel {
	now := pollsweb.UTCNow()
	return &MeetingModel{
		IdModel:     EmptyIdModel(),
		Name:        "",
		Slug:        "",
		Created:     now,
		Period:      "",
		MeetingTime: time.Time{},
		OnlineStart: time.Time{},
		OnlineEnd:   time.Time{},
		Voters:      nil,
		Groups:      nil,
		LastUpdated: time.Time{},
		UpdateToken: rand.Int63(),
	}
}

func NewMeetingModel(name, slug string, period string, meetingTime, onlineStart, onlineEnd time.Time, voters []*VoterModel, groups []*PollGroupModel) *MeetingModel {
	now := pollsweb.UTCNow()
	return &MeetingModel{
		IdModel:     EmptyIdModel(),
		Name:        name,
		Slug:        slug,
		Created:     now,
		Period:      period,
		MeetingTime: meetingTime,
		OnlineStart: onlineStart,
		OnlineEnd:   onlineEnd,
		Voters:      voters,
		Groups:      groups,
		LastUpdated: now,
		UpdateToken: rand.Int63(),
	}
}

func (meeting *MeetingModel) String() string {
	return fmt.Sprintf("MeetingModel(Id=%s, Name=%s, Slug=%s, Created=%s, Period=%s, MeetingTime=%s, OnlineStart=%s, OnlineEnd=%s, Voters=%v, Groups=%v, LastUpdated=%s, UpdateToken=%d)",
		meeting.Id, meeting.Name, meeting.Slug, meeting.Created, meeting.Period, meeting.MeetingTime,
		meeting.OnlineStart, meeting.OnlineEnd, meeting.Voters, meeting.Groups, meeting.LastUpdated,
		meeting.UpdateToken)
}

func (meeting *MeetingModel) GenIds() error {
	// we re-use these variables
	var genId uuid.UUID
	var genErr error

	// first on the instance
	genId, genErr = pollsweb.GenUUID()
	if genErr != nil {
		return genErr
	}
	meeting.SetId(genId)

	// all voters
	for _, voter := range meeting.Voters {
		genId, genErr = pollsweb.GenUUID()
		if genErr != nil {
			return genErr
		}
		voter.SetId(genId)
	}
	// for all groups
	for _, group := range meeting.Groups {
		genErr = group.GenIds()
		if genErr != nil {
			return genErr
		}
	}
	return nil
}
