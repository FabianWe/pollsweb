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

package pollsweb

import (
	"errors"
	"fmt"
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/google/uuid"
	"strconv"
	"strings"
)

const (
	BasicPollName   = "basic"
	MedianPollName  = "median"
	SchulzePollName = "schulze"
)

func nameAndSlugFieldValidator(fieldPtr interface{}) *validation.FieldRules {
	return validation.Field(fieldPtr, validation.Required, validation.RuneLength(2, 250))
}

var ErrIDAlreadySet = errors.New("ID (uuid) field is already set")

type BaseModel struct {
	ID uuid.UUID
}

func (model *BaseModel) GenerateID() (uuid.UUID, error) {
	if model.ID != uuid.Nil {
		return uuid.Nil, ErrIDAlreadySet
	}
	id, err := uuid.NewRandom()
	if err != nil {
		return uuid.Nil, err
	}
	model.ID = id
	return id, nil
}

func (model *BaseModel) GetID() uuid.UUID {
	return model.ID
}

type Model interface {
	validation.Validatable
}

type IDModel interface {
	Model
	GetID() uuid.UUID
	GenerateID() (uuid.UUID, error)
}

type VoterModel struct {
	BaseModel
	Name   string
	Slug   string
	Weight uint32
}

func NewVoterModel(name string, weight uint32) *VoterModel {
	return &VoterModel{
		Name:   name,
		Slug:   slugGenerator.GenerateSlug(name),
		Weight: weight,
	}
}

func (voter *VoterModel) Validate() error {
	return validation.ValidateStruct(voter,
		nameAndSlugFieldValidator(&voter.Name),
		nameAndSlugFieldValidator(&voter.Slug),
		// the weight in the database is not allowed to be too big, we need want it to fit in a "standard" db integer
		// thus the max value would be 2147483647, we subtract two because some databases might use something else
		validation.Field(&voter.Weight, validation.Required, validation.Max(uint32(2147483647-2))),
	)
}

type VotersRevision struct {
	BaseModel
	Name   string
	Slug   string
	Voters []*VoterModel
}

func NewVotersRevision(name string) *VotersRevision {
	return &VotersRevision{
		Name:   name,
		Slug:   slugGenerator.GenerateSlug(name),
		Voters: nil,
	}
}

func (revision *VotersRevision) Validate() error {
	return validation.ValidateStruct(revision,
		nameAndSlugFieldValidator(&revision.Name),
		nameAndSlugFieldValidator(&revision.Slug),
		validation.Field(&revision.Voters, validation.Each()),
	)
}

type MajorityModel struct {
	Numerator, Denominator uint32
}

func NewMajorityModel(numerator, denominator uint32) *MajorityModel {
	return &MajorityModel{
		Numerator:   numerator,
		Denominator: denominator,
	}
}

func (majority *MajorityModel) Validate() error {
	return validation.ValidateStruct(majority,
		validation.Field(&majority.Numerator, validation.Required, validation.Min(uint32(1)), validation.Max(uint32(2147483647-2))),
		validation.Field(&majority.Denominator, validation.Required, validation.Min(uint32(1)), validation.Max(uint32(2147483647-2))),
	)
}

func (majority *MajorityModel) FormatDBString() string {
	return fmt.Sprintf("%d/%d", majority.Numerator, majority.Denominator)
}

func ParseMajorityModelDBString(s string) (*MajorityModel, error) {
	split := strings.Split(s, "/")
	if len(split) != 2 {
		return nil, fmt.Errorf("invalid majority string in database, must be in the format \"a/b\" where a and b are integers, got %s", s)
	}
	numeratorStr, denominatorStr := split[0], split[1]
	numerator, numeratorErr := strconv.ParseUint(numeratorStr, 10, 32)
	if numeratorErr != nil {
		return nil, fmt.Errorf("can't parse majority numerator from string %s: %w", numeratorStr, numeratorErr)
	}
	denominator, denominatorErr := strconv.ParseUint(denominatorStr, 10, 32)
	if denominatorErr != nil {
		return nil, fmt.Errorf("can't parse majority denominator from string %s: %w", denominatorStr, denominatorErr)
	}
	return NewMajorityModel(uint32(numerator), uint32(denominator)), nil
}

type PollModel struct {
	BaseModel
	Name             string
	Slug             string
	Majority         *MajorityModel
	AbsoluteMajority bool
	Type             string
}

func NewPollModel(name string, majority *MajorityModel, absoluteMajority bool, _type string) *PollModel {
	return &PollModel{
		Name:             name,
		Slug:             slugGenerator.GenerateSlug(name),
		Majority:         majority,
		AbsoluteMajority: absoluteMajority,
		Type:             _type,
	}
}

func (poll *PollModel) Validate() error {
	return validation.ValidateStruct(poll,
		nameAndSlugFieldValidator(&poll.Name),
		nameAndSlugFieldValidator(&poll.Slug),
		validation.Field(&poll.Majority),
		validation.Field(&poll.Type, validation.In(BasicPollName, MedianPollName, SchulzePollName)),
	)
}

type BasicPollModel struct {
	*PollModel
}

func NewBasicPollModel(name string, majority *MajorityModel, absoluteMajority bool) *BasicPollModel {
	return &BasicPollModel{NewPollModel(name, majority, absoluteMajority, BasicPollName)}
}

type MedianPollModel struct {
	*PollModel
	Value    uint32
	Currency string
}

func NewMedianPollModel(name string, majority *MajorityModel, absoluteMajority bool, value uint32, currency string) *MedianPollModel {
	return &MedianPollModel{
		PollModel: NewPollModel(name, majority, absoluteMajority, MedianPollName),
		Value:     value,
		Currency:  currency,
	}
}

func (poll *MedianPollModel) Validate() error {
	return validation.ValidateStruct(poll,
		validation.Field(&poll.PollModel),
		validation.Field(&poll.Value, validation.Required, validation.Min(uint32(1)), validation.Max(uint32(2147483647-2))),
		validation.Field(&poll.Currency, validation.Required, validation.RuneLength(1, 5)),
	)
}

type SchulzeOption struct {
	BaseModel
	Option string
}

func NewSchulzeOption(option string) *SchulzeOption {
	return &SchulzeOption{
		Option: option,
	}
}

func (option *SchulzeOption) Validate() error {
	return validation.ValidateStruct(option,
		validation.Field(&option.Option, validation.Required, validation.RuneLength(1, 300)),
	)
}

type SchulzePollModel struct {
	*PollModel
	Options []*SchulzeOption
}

func (poll *SchulzePollModel) Validate() error {
	return validation.ValidateStruct(poll,
		validation.Field(&poll.Options, validation.Each()),
	)
}
