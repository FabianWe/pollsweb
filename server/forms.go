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

package server

import (
	"fmt"
	"github.com/FabianWe/gopolls"
	"github.com/FabianWe/pollsweb"
	"github.com/gorilla/schema"
	"reflect"
	"regexp"
	"strconv"
	"time"
)

type HourMinuteFormField struct {
	Hour, Minute uint8
}

func (f HourMinuteFormField) Equals(other HourMinuteFormField) bool {
	return f.Hour == other.Hour && f.Minute == other.Minute
}

func (f HourMinuteFormField) String() string {
	return fmt.Sprintf("HourMinuteFormField(Hour=%d, Minute=%d)", f.Hour, f.Minute)
}

var hourMinuteRegex = regexp.MustCompile(`^([0-9]{2}):([0-9]{2})$`)

func ParseHoursAndMinutes(s string) (field HourMinuteFormField, err error) {
	// makes sure to return nil in case of an error
	defer func() {
		if err != nil {
			field = HourMinuteFormField{}
		}
	}()
	match := hourMinuteRegex.FindStringSubmatch(s)
	if match == nil {
		err = NewFormValidationError(fmt.Sprintf("can't parse hour and minute, must be of the form \"HH:mm\", got %s", s))
		return
	}
	hourString, minuteString := match[1], match[2]
	var placeholder uint64
	var parseErr error
	placeholder, parseErr = strconv.ParseUint(hourString, 10, 8)
	if parseErr != nil {
		err = NewFormValidationError(fmt.Sprintf("can't parse hour as integer, got hour \"%s\"", hourString)).
			SetWrapped(parseErr)
		return
	}
	field.Hour = uint8(placeholder)
	placeholder, parseErr = strconv.ParseUint(minuteString, 10, 8)
	if parseErr != nil {
		err = NewFormValidationError(fmt.Sprintf("can't parse minute as integer, got hour \"%s\"", minuteString)).
			SetWrapped(parseErr)
		return
	}
	field.Minute = uint8(placeholder)
	// validate values: hour must be <= 23, minute <= 59
	if field.Hour > 23 {
		err = NewFormValidationError(fmt.Sprintf("invalid hour: must be <= 23, got %d", field.Hour))
		return
	}
	if field.Minute > 59 {
		err = NewFormValidationError(fmt.Sprintf("invalid minute: must be <= 59, got %d", field.Minute))
		return
	}
	// everything ok
	return
}

func decodeHourMinuteFormField(s string) reflect.Value {
	res, err := ParseHoursAndMinutes(s)
	if err == nil {
		return reflect.ValueOf(res)
	}
	return reflect.Value{}
}

// The following formats are used to format / parse files in forms, forms must make sure
// not to use the display format but these formats when sending form data.
// All data sent is expected to be in UTC

const InternalDateFormat = "2006/01/02"

var InternalDateFormatMomentJS = pollsweb.MomentJSDateFormatter.ConvertFormat(InternalDateFormat)

const InternalDateTimeFormat = "2006/01/02 15:04"

var InternalDateTimeFormatMomentJS = pollsweb.MomentJSDateFormatter.ConvertFormat(InternalDateTimeFormat)

type DateFormField time.Time

func NewDateFormField(year int, month time.Month, day int) DateFormField {
	return DateFormField(time.Date(year, month, day, 0, 0, 0, 0, time.UTC))
}

func (d DateFormField) Equals(other DateFormField) bool {
	return time.Time(d).Equal(time.Time(other))
}

func (d DateFormField) String() string {
	return time.Time(d).Format(InternalDateFormat)
}

func ParseDateFormField(s string) (DateFormField, error) {
	res, err := time.ParseInLocation(InternalDateFormat, s, time.UTC)
	if err != nil {
		return DateFormField(res), NewFormValidationError(fmt.Sprintf("can't parse as date: invalid format (for \"%s\")", s)).
			SetWrapped(err)
	}
	return DateFormField(res), nil
}

func decodeDateFormField(s string) reflect.Value {
	res, err := ParseDateFormField(s)
	if err == nil {
		return reflect.ValueOf(res)
	}
	return reflect.Value{}
}

type DateTimeFormField time.Time

func NewDateTimeFormField(year int, month time.Month, day, hour, min int) DateTimeFormField {
	return DateTimeFormField(time.Date(year, month, day, hour, min, 0, 0, time.UTC))
}

func (dt DateTimeFormField) Equals(other DateTimeFormField) bool {
	return time.Time(dt).Equal(time.Time(other))
}

func (dt DateTimeFormField) String() string {
	return time.Time(dt).Format(InternalDateTimeFormat)
}

func ParseDateTimeFormField(s string) (DateTimeFormField, error) {
	res, err := time.ParseInLocation(InternalDateTimeFormat, s, time.UTC)
	if err != nil {
		return DateTimeFormField(res), NewFormValidationError(fmt.Sprintf("can't parse as datetime: invalid format (for \"%s\")", s)).
			SetWrapped(err)
	}
	return DateTimeFormField(res), nil
}

func decodeDateTimeFormField(s string) reflect.Value {
	res, err := ParseDateTimeFormField(s)
	if err == nil {
		return reflect.ValueOf(res)
	}
	return reflect.Value{}
}

type WeekdayFormField time.Weekday

func ParseWeekdayFormField(s string) (WeekdayFormField, error) {
	var weekday time.Weekday
	var err error
	switch s {
	case "0":
		weekday = time.Sunday
	case "1":
		weekday = time.Monday
	case "2":
		weekday = time.Tuesday
	case "3":
		weekday = time.Wednesday
	case "4":
		weekday = time.Thursday
	case "5":
		weekday = time.Friday
	case "6":
		weekday = time.Saturday
	default:
		weekday = -1
		err = NewFormValidationError(fmt.Sprintf("weekday must be an int between 0 and 6, got %s", s))
	}
	return WeekdayFormField(weekday), err
}

func (w WeekdayFormField) String() string {
	return time.Weekday(w).String()
}

func decodeWeekdayFormField(s string) reflect.Value {
	res, err := ParseWeekdayFormField(s)
	if err == nil {
		return reflect.ValueOf(res)
	}
	return reflect.Value{}
}

type VotersFormField struct {
	Voters []*gopolls.Voter
}

func NewVotersFormField(voters []*gopolls.Voter) VotersFormField {
	return VotersFormField{Voters: voters}
}

func GetVotersFormFieldConverter(parser *gopolls.VotersParser) schema.Converter {
	return func(s string) reflect.Value {
		voters, err := parser.ParseVotersFromString(s)
		if err == nil {
			res := NewVotersFormField(voters)
			return reflect.ValueOf(res)
		}
		return reflect.Value{}
	}
}

type PeriodForm struct {
	Name        string              `schema:"period_name" valid:"runelength(5|200)"`
	Start       DateTimeFormField   `schema:"period_start" valid:"-"`
	End         DateTimeFormField   `schema:"period_end" valid:"-"`
	Weekday     WeekdayFormField    `schema:"weekday" valid:"-"`
	MeetingTime HourMinuteFormField `schema:"time" valid:"-"`
}

func (form PeriodForm) ValidateForm() error {
	startAsTime := time.Time(form.Start)
	endAsTime := time.Time(form.End)
	if endAsTime.Before(startAsTime) {
		return NewFormValidationError(fmt.Sprintf("end date is after start date: start=\"%s\", end=\"%s\"",
			form.Start, form.End))
	}
	return nil
}

func DecodePeriodForm(src map[string][]string) (*PeriodForm, error) {
	res := PeriodForm{}
	err := DecodeForm(&res, src)
	return &res, err
}
