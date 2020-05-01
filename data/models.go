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

package data

import (
	"fmt"
	"github.com/FabianWe/pollsweb"
	"github.com/google/uuid"
	"github.com/hashicorp/go-multierror"
	"time"
	"unicode/utf8"
)

const DefaultMaxNameLength = 150
const DefaultMaxSlugLength = 150

const MaxPeriodNameLength = DefaultMaxNameLength
const MaxPeriodSlugLength = DefaultMaxSlugLength

type FieldValidationError struct {
	pollsweb.PollWebError
	FieldName    string
	FieldType    string
	ErrorMessage string
}

func NewFieldValidationError(fieldName, fieldType, errorMessage string) FieldValidationError {
	return FieldValidationError{
		FieldName:    fieldName,
		FieldType:    fieldType,
		ErrorMessage: errorMessage,
	}
}

func NewStringTooLongValidationError(fieldName string, maxLength int) FieldValidationError {
	return NewFieldValidationError(fieldName, "string", fmt.Sprintf("longer than %d charachters", maxLength))
}

func NewStringTooShortValidationError(fieldName string, minLength int) FieldValidationError {
	return NewFieldValidationError(fieldName, "string", fmt.Sprintf("shortern than %d charachters", minLength))
}

func NewIsNilValidationError(fieldName, fieldType string) FieldValidationError {
	return NewFieldValidationError(fieldName, fieldType, "is not allowed to be nil")
}

func NewTooBigValidationError(fieldName, fieldType, gotValue, maxValue string) FieldValidationError {
	errorMessage := fmt.Sprintf("value too big: %s > %s", gotValue, maxValue)
	return NewFieldValidationError(fieldName, fieldType, errorMessage)
}

func NewTooSmallValidationError(fieldName, fieldType, gotValue, maxValue string) FieldValidationError {
	errorMessage := fmt.Sprintf("value too small: %s < %s", gotValue, maxValue)
	return NewFieldValidationError(fieldName, fieldType, errorMessage)
}

func (err FieldValidationError) Error() string {
	return fmt.Sprintf("field validation error: field %s of type %s: %s",
		err.FieldName, err.FieldType, err.ErrorMessage)
}

// TODO test this!
type ModelValidationError struct {
	FieldErrors *multierror.Error
}

func (err ModelValidationError) Error() string {
	return err.FieldErrors.Error()
}

func (err ModelValidationError) Unwrap() error {
	return err.FieldErrors.Unwrap()
}

func NewModelValidationError() *ModelValidationError {
	return &ModelValidationError{
		nil,
	}
}

func (err *ModelValidationError) ErrorOrNil() error {
	return err.FieldErrors.ErrorOrNil()
}

func (err ModelValidationError) AppendTo(errs ...error) {
	err.FieldErrors = multierror.Append(err.FieldErrors, errs...)
}

func ValidateStringLen(s, fieldName string, minLength, maxLength int) error {
	if minLength < 0 && maxLength < 0 {
		return nil
	}
	length := utf8.RuneCountInString(s)
	if minLength > 0 && length < minLength {
		return NewStringTooShortValidationError(fieldName, minLength)
	}
	if maxLength > 0 && length > maxLength {
		return NewStringTooLongValidationError(fieldName, maxLength)
	}
	return nil
}

type Model interface {
	ValidateFields() error
}

type PeriodModel struct {
	ID          uuid.UUID
	Name        string
	Slug        string
	MeetingTime time.Time
	PeriodStart time.Time
	PeriodEnd   time.Time
	Created     time.Time
}

func GeneratePeriodModel(appContext *pollsweb.AppContext, name, slug string, meetingTime, periodStart, periodEnd time.Time) (*PeriodModel, error) {
	id, idErr := pollsweb.GenUUID()
	if idErr != nil {
		return nil, idErr
	}
	if slug == "" {
		slug = appContext.Generator.GenSlug(name)
	}
	res := PeriodModel{
		ID:          id,
		Name:        name,
		Slug:        slug,
		MeetingTime: meetingTime,
		PeriodStart: periodStart,
		PeriodEnd:   periodEnd,
		Created:     time.Time{},
	}
	return &res, nil
}

func (period *PeriodModel) ValidateFields() error {
	err := NewModelValidationError()

	if nameErr := ValidateStringLen(period.Name, "name", -1, MaxPeriodNameLength); nameErr != nil {
		err.AppendTo(nameErr)
	}

	if slugErr := ValidateStringLen(period.Slug, "slug", -1, MaxPeriodSlugLength); slugErr != nil {
		err.AppendTo(slugErr)
	}

	return err.ErrorOrNil()
}

type VotersRevisionModel struct {
	ID       uuid.UUID
	PeriodID uuid.UUID
	Created  time.Time
	Note     string
	IsActive bool
}

func (rev *VotersRevisionModel) ValidateFields() error {
	return nil
}
