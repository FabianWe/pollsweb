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
	"reflect"
	"strings"
	"time"
)

// EntryNotFoundError is an error returned if an entry could not be found in the database.
//
// It embeds PollWebError and is thus an internal error.
// It can wrap another error (for example the original database error), Wrapped can also be nil.
// It also references the key / query args for which the lookup failed by wrapping it in a reflect.Value.
type EntryNotFoundError struct {
	pollsweb.PollWebError
	Model    reflect.Type
	KeyValue reflect.Value
	Wrapped  error
}

// NewEntryNotFoundError returns a new error given the model type, the key that caused the failure
// and a wrapped error (which may be nil).
func NewEntryNotFoundError(model reflect.Type, key reflect.Value, wrapped error) EntryNotFoundError {
	return EntryNotFoundError{
		Model:    model,
		KeyValue: key,
		Wrapped:  wrapped,
	}
}

func (e EntryNotFoundError) Error() string {
	return fmt.Sprintf("entry of type \"%v\" not found, key / query specification %v: \"%v\"",
		e.Model, e.KeyValue.Type(), e.KeyValue)
}

func (e EntryNotFoundError) Unwrap() error {
	return e.Wrapped
}

type InvalidQueryArgsError struct {
	pollsweb.PollWebError
	Message string
}

func NewInvalidQueryArgsError(message string) InvalidQueryArgsError {
	return InvalidQueryArgsError{Message: message}
}

func (e InvalidQueryArgsError) Error() string {
	return e.Message
}

func (e InvalidQueryArgsError) Unwrap() error {
	return nil
}

func formatSimpleQueryArgs(argsType reflect.Type, arguments []string) string {
	var buf strings.Builder
	buf.WriteString(argsType.String())
	buf.WriteString(" : {\n")
	for _, arg := range arguments {
		buf.WriteRune('\t')
		buf.WriteString(arg)
		buf.WriteString(",\n")
	}
	buf.WriteRune('}')
	return buf.String()
}

type PeriodSettingsQueryArgs struct {
	Id   *uuid.UUID
	Name *string
	Slug *string
}

func NewPeriodSettingsQueryArgs() *PeriodSettingsQueryArgs {
	return &PeriodSettingsQueryArgs{
		Id:   nil,
		Name: nil,
		Slug: nil,
	}
}

func (args *PeriodSettingsQueryArgs) SetId(id *uuid.UUID) *PeriodSettingsQueryArgs {
	args.Id = id
	return args
}

func (args *PeriodSettingsQueryArgs) SetName(name *string) *PeriodSettingsQueryArgs {
	args.Name = name
	return args
}

func (args *PeriodSettingsQueryArgs) SetSlug(slug *string) *PeriodSettingsQueryArgs {
	args.Slug = slug
	return args
}

func (args *PeriodSettingsQueryArgs) String() string {
	asStrings := make([]string, 0, 1)
	if args.Id != nil {
		asStrings = append(asStrings, fmt.Sprintf("Id = \"%s\"", *args.Id))
	}
	if args.Slug != nil {
		asStrings = append(asStrings, fmt.Sprintf("Slug = \"%s\"", *args.Slug))
	}
	if args.Name != nil {
		asStrings = append(asStrings, fmt.Sprintf("Name = \"%s\"", *args.Name))
	}
	return formatSimpleQueryArgs(reflect.TypeOf(args), asStrings)
}

type MeetingQueryArgs struct {
	Id          *uuid.UUID
	Name        *string
	Slug        *string
	LastUpdated *time.Time
	UpdateToken *int64
}

func NewMeetingQueryArgs() *MeetingQueryArgs {
	return &MeetingQueryArgs{
		Id:          nil,
		Name:        nil,
		Slug:        nil,
		LastUpdated: nil,
		UpdateToken: nil,
	}
}

func (args *MeetingQueryArgs) SetId(id *uuid.UUID) *MeetingQueryArgs {
	args.Id = id
	return args
}

func (args *MeetingQueryArgs) SetName(name *string) *MeetingQueryArgs {
	args.Name = name
	return args
}

func (args *MeetingQueryArgs) SetSlug(slug *string) *MeetingQueryArgs {
	args.Slug = slug
	return args
}

func (args *MeetingQueryArgs) SetLastUpdated(lastUpdated *time.Time) *MeetingQueryArgs {
	args.LastUpdated = lastUpdated
	return args
}

func (args *MeetingQueryArgs) SetUpdateToken(updateToken *int64) *MeetingQueryArgs {
	args.UpdateToken = updateToken
	return args
}

func (args *MeetingQueryArgs) String() string {
	asStrings := make([]string, 0, 3)
	if args.Id != nil {
		asStrings = append(asStrings, fmt.Sprintf("Id = \"%s\"", *args.Id))
	}
	if args.Slug != nil {
		asStrings = append(asStrings, fmt.Sprintf("Slug = \"%s\"", *args.Slug))
	}
	if args.Name != nil {
		asStrings = append(asStrings, fmt.Sprintf("Name = \"%s\"", *args.Name))
	}
	if args.LastUpdated != nil {
		asStrings = append(asStrings, fmt.Sprintf("LastUpdated = \"%s\"", *args.LastUpdated))
	}
	if args.UpdateToken != nil {
		asStrings = append(asStrings, fmt.Sprintf("UpdateToken = %d", *args.UpdateToken))
	}
	return formatSimpleQueryArgs(reflect.TypeOf(args), asStrings)
}

var ErrInvalidPeriodSettingsQuery = NewInvalidQueryArgsError("invalid query for PeriodSettingsModel: Id, Name or Slug must be given")
var ErrInvalidMeetingQuery = NewInvalidQueryArgsError("invalid query for MeetingModel: Id, Name or Slug must be given")

type PeriodSettingsHandler interface {
	InsertPeriod(ctx context.Context, meetingTime *PeriodSettingsModel) (uuid.UUID, error)

	GetPeriod(ctx context.Context, args *PeriodSettingsQueryArgs) (*PeriodSettingsModel, error)
	GetActivePeriods(ctx context.Context, referenceTime time.Time) ([]*PeriodSettingsModel, error)

	DeletePeriod(ctx context.Context, args *PeriodSettingsQueryArgs) (int64, error)
}

type MeetingsHandler interface {
	InsertMeeting(ctx context.Context, meeting *MeetingModel) error

	GetMeeting(ctx context.Context, args *MeetingQueryArgs) (*MeetingModel, error)

	DeleteMeeting(ctx context.Context, args *MeetingQueryArgs) (int64, error)
}

// TODO clarify when UUIDs are generated
// 	should we disallow 00000... uuid? nearly impossible this happens ;)

type DataHandler interface {
	PeriodSettingsHandler
	MeetingsHandler
	Close(ctx context.Context) error
}
