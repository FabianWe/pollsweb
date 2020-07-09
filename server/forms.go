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
	"github.com/FabianWe/pollsweb"
	"github.com/gorilla/schema"
	"reflect"
	"regexp"
	"strconv"
	"time"
)

type FormValidationError struct {
	pollsweb.PollWebError
	FieldName string
	Message   string
	Wrapped   error
}

func NewFormValidationError(message string) *FormValidationError {
	return &FormValidationError{
		FieldName: "",
		Message:   message,
		Wrapped:   nil,
	}
}

func (e *FormValidationError) SetFieldName(fieldName string) *FormValidationError {
	e.FieldName = fieldName
	return e
}

func (e *FormValidationError) SetWrapped(wrapped error) *FormValidationError {
	e.Wrapped = wrapped
	return e
}

func (e *FormValidationError) Error() string {
	msg := "form validation error"
	if e.FieldName != "" {
		msg += fmt.Sprintf(" for field \"%s\"", e.FieldName)
	}
	msg += ": "
	msg += e.Message
	if e.Wrapped != nil {
		msg += ". Cause: " + e.Wrapped.Error()
	}
	return msg
}

func (e *FormValidationError) Unwrap() error {
	return e.Wrapped
}

const parseHoursAndMinutesErrMsgFormat = "invalid time format, must be \"HH:mm\", got string %s"

// TODO is it a good idea to re-use encoders? or should a new one always be created? not clear from doc...

var FormDecoder = schema.NewDecoder()

func DecodeForm(dst interface{}, src map[string][]string) error {
	decodeErr := FormDecoder.Decode(dst, src)
	if decodeErr != nil {
		// test if it's a conversion error
		if asConversionErr, ok := decodeErr.(schema.ConversionError); ok {
			return NewFormValidationError("unable to decode form:").SetWrapped(asConversionErr)
		} else {
			return decodeErr
		}
	}
	return nil
}

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
	return DateFormField(res), err
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
	return DateTimeFormField(res), err
}

func decodeDateTimeFormField(s string) reflect.Value {
	res, err := ParseDateTimeFormField(s)
	if err == nil {
		return reflect.ValueOf(res)
	}
	return reflect.Value{}
}

func init() {
	FormDecoder.RegisterConverter(HourMinuteFormField{}, decodeHourMinuteFormField)
	FormDecoder.RegisterConverter(DateFormField{}, decodeDateFormField)
	FormDecoder.RegisterConverter(DateTimeFormField{}, decodeDateTimeFormField)
}
