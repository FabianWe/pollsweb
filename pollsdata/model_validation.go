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
	"github.com/asaskevich/govalidator"
	"github.com/hashicorp/go-multierror"
	"log"
	"reflect"
	"regexp"
	"unicode/utf8"
)

type ModelValidationError struct {
	pollsweb.PollWebError
	FieldName string
	Message   string
	Wrapped   error
}

func NewModelValidationError(message string) *ModelValidationError {
	return &ModelValidationError{
		FieldName: "",
		Message:   message,
		Wrapped:   nil,
	}
}

func (e *ModelValidationError) SetFieldName(fieldName string) *ModelValidationError {
	e.FieldName = fieldName
	return e
}

func (e *ModelValidationError) SetWrapped(wrapped error) *ModelValidationError {
	e.Wrapped = wrapped
	return e
}

func (e *ModelValidationError) Error() string {
	msg := "model validation error"
	if e.FieldName != "" {
		msg += fmt.Sprintf(" for field \"%s\"", e.FieldName)
	}
	if e.Message != "" {
		msg += ": "
		msg += e.Message
	}
	if e.Wrapped != nil {
		msg += ". Cause: " + e.Wrapped.Error()
	}
	return msg
}

func (e *ModelValidationError) Unwrap() error {
	return e.Wrapped
}

type ValidatorModel interface {
	ValidateModel() error
}

type CustomValidator func(model interface{}, validator *ModelValidator) error

type ModelValidator struct {
	CustomValidators map[reflect.Type][]CustomValidator
}

func NewModelValidator() *ModelValidator {
	return &ModelValidator{
		CustomValidators: make(map[reflect.Type][]CustomValidator),
	}
}

// TODO test this
// TODO does the govalidator package validate all elements in an array / embedded fields?
//		we must iterate in our own code anyway... I would like to avoid double effort?
//		we have to deal with embedded types!
func (validator *ModelValidator) Validate(val interface{}) error {
	var result *multierror.Error
	if ok, validateErr := govalidator.ValidateStruct(val); ok {
		if validateErr != nil {
			log.Printf("unexepcted result from model validation: got an error: %v", validateErr)
			result = multierror.Append(result, NewModelValidationError("unknown reason").SetWrapped(validateErr))
			return result
		}
	} else {
		if validateErr == nil {
			log.Printf("unexpected result from model validation: result is not okay, but no error was given")
			result = multierror.Append(result, NewModelValidationError(""))
			return result
		}
		// try to convert to govalidator.Error
		if asValidatorErrs, isValidatorErrs := validateErr.(govalidator.Errors); isValidatorErrs {
			// append field errors, but make sure we append at least one error... I don't think
			// that the error list can become empty, but just to be sure
			if len(asValidatorErrs.Errors()) == 0 {
				result = multierror.Append(result, NewModelValidationError("").SetWrapped(asValidatorErrs))
				return result
			}
			for _, innerErr := range asValidatorErrs.Errors() {
				// check if it is a govalidation.Error
				if asValidatorErr, isValidatorErr := innerErr.(govalidator.Error); isValidatorErr {
					result = multierror.Append(result, NewModelValidationError("").SetFieldName(asValidatorErr.Name).SetWrapped(asValidatorErr))
				} else {
					// append as a genric error
					result = multierror.Append(result, NewModelValidationError("").SetWrapped(innerErr))
				}
			}
		} else {
			// just create a generic error
			result = multierror.Append(result, NewModelValidationError("").SetWrapped(validateErr))
		}
		// should never be nil here...
		return result.ErrorOrNil()
	}
	// run all custom validators for this type
	customValidators := validator.CustomValidators[reflect.TypeOf(val)]
	for _, validatorFunc := range customValidators {
		if customValidationErr := validatorFunc(val, validator); customValidationErr != nil {
			result = multierror.Append(result, customValidationErr)
		}
	}
	// if the model is also a ValidatorModel run this validation too
	if asValidatorModel, isValidatorModel := val.(ValidatorModel); isValidatorModel {
		if modelValidationErr := asValidatorModel.ValidateModel(); modelValidationErr != nil {
			result = multierror.Append(result, modelValidationErr)
		}
	}
	return result.ErrorOrNil()
}

func runeLengthValidator(s string, min, max int) *ModelValidationError {
	n := utf8.RuneCountInString(s)
	switch {
	case min > 0 && n < min:
		return NewModelValidationError(fmt.Sprintf("string is too short, must have at least length of %d", min))
	case max > 0 && n > max:
		return NewModelValidationError(fmt.Sprintf("string is too long, must have at most length of %d", max))
	default:
		return nil
	}
}

// this is a very much simplified regex, but it should be enough
var slugRegex = regexp.MustCompile(`^[a-z0-9-_]+$`)

func slugValidator(s string) *ModelValidationError {
	if match := slugRegex.FindStringSubmatch(s); match == nil {
		return NewModelValidationError("slug has an invalid form")
	}
	return nil
}

func weightValidator(w, max gopolls.Weight) *ModelValidationError {
	if w == gopolls.NoWeight {
		return NewModelValidationError("not a valid weight")
	}
	if max != gopolls.NoWeight && w > max {
		return NewModelValidationError(fmt.Sprintf("weight must be <= %d", max))
	}
	return nil
}

func strictlyPositiveInt64Validator(i int64) *ModelValidationError {
	if i <= 0 {
		return NewModelValidationError("must be ≥ 0")
	}
	return nil
}
