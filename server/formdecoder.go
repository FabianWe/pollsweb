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
	"github.com/asaskevich/govalidator"
	"github.com/gorilla/schema"
	"golang.org/x/text/unicode/norm"
	"log"
	"time"
	"unicode/utf8"
)

type CustomFormValidator interface {
	ValidateForm() error
}

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

func NewSchemaDecoder() *schema.Decoder {
	res := schema.NewDecoder()
	res.RegisterConverter(HourMinuteFormField{}, decodeHourMinuteFormField)
	res.RegisterConverter(DateFormField{}, decodeDateFormField)
	res.RegisterConverter(DateTimeFormField{}, decodeDateTimeFormField)
	res.RegisterConverter(WeekdayFormField(time.Sunday), decodeWeekdayFormField)
	return res
}

// TODO is it a good idea to re-use encoders? or should a new one always be created? not clear from doc...
var DefaultSchemaDecoder = NewSchemaDecoder()

type FormDecoder struct {
	UTF8Form      norm.Form
	SchemaDecoder *schema.Decoder
}

func NewFormDecoder() *FormDecoder {
	return &FormDecoder{
		UTF8Form:      norm.NFKC,
		SchemaDecoder: DefaultSchemaDecoder,
	}
}

func (decoder *FormDecoder) ValidateAndNormalizeFormStrings(src map[string][]string) (map[string][]string, error) {
	res := make(map[string][]string, len(src))
	for key, values := range src {
		if !utf8.ValidString(key) {
			return nil, NewFormValidationError("invalid string in form key")
		}
		newValues := make([]string, len(values))
		for i, value := range values {
			if !utf8.ValidString(value) {
				return nil, NewFormValidationError("invalid string in form value")
			}
			newValues[i] = decoder.UTF8Form.String(value)
		}
		res[decoder.UTF8Form.String(key)] = newValues
	}
	return res, nil
}

func (decoder *FormDecoder) DecodeForm(dst interface{}, src map[string][]string) error {
	decodeErr := decoder.SchemaDecoder.Decode(dst, src)
	if decodeErr != nil {
		// test if it's a conversion error
		if asConversionErr, ok := decodeErr.(schema.ConversionError); ok {
			return NewFormValidationError("unable to decode form").SetWrapped(asConversionErr)
		} else {
			return decodeErr
		}
	}
	// validate struct
	// TODO form validation: iterate errors?
	if ok, validateErr := govalidator.ValidateStruct(dst); ok {
		if validateErr != nil {
			// log this (using the normal logger)
			log.Printf("unexepcted result from form validation: got an error: %v", validateErr)
			return NewFormValidationError("form validation failed").SetWrapped(validateErr)
		}
		// in this case we continue after the outer if
	} else {
		if validateErr == nil {
			log.Printf("unexpected result from validaton: result is not okay, but no error was given")
			return NewFormValidationError("form validation return not okay, but no error was given")
		}
		return NewFormValidationError("form validation failed").SetWrapped(validateErr)
	}
	// validator package succeeded, if applicable run custom form validation
	if formValidator, isFormValidator := dst.(CustomFormValidator); isFormValidator {
		// perform custom validation logic of the form
		return formValidator.ValidateForm()
	}
	return nil
}

func (decoder *FormDecoder) NormalizeAndDecodeForm(dst interface{}, src map[string][]string) error {
	normalizedSrc, stringValidationErr := decoder.ValidateAndNormalizeFormStrings(src)
	if stringValidationErr != nil {
		return stringValidationErr
	}
	return decoder.DecodeForm(dst, normalizedSrc)
}

var DefaultFormDecoder = NewFormDecoder()

func DecodeForm(dst interface{}, src map[string][]string) error {
	return DefaultFormDecoder.NormalizeAndDecodeForm(dst, src)
}
