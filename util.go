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
	"github.com/google/uuid"
	"github.com/gosimple/slug"
	"time"
)

// UUIDGenError is an error returned whenever we're not able a UUID.
// This should never happen.
type UUIDGenError struct {
	PollWebError
	Wrapped error
}

// NewUUIDGenError returns a new UUIDGenError given the wrapped error.
func NewUUIDGenError(err error) UUIDGenError {
	return UUIDGenError{
		PollWebError: PollWebError{},
		Wrapped:      err,
	}
}

func (err UUIDGenError) Error() string {
	return "can't generate UUID: " + err.Wrapped.Error()
}

func (err UUIDGenError) Unwrap() error {
	return err.Wrapped
}

// GenUUID generates a new UUID.
// The returned UUID is a random id, for consistent usage this function should always be called
// to generate UUIDs.
//
// The returned error is (when not nil) of type UUIDGenError.
func GenUUID() (uuid.UUID, error) {
	res, err := uuid.NewRandom()
	if err != nil {
		return res, NewUUIDGenError(err)
	}
	return res, nil
}

// GenNow returns the current time in UTC.
// For consistent usage this function should always be called to generate the current time.
func GenNow() time.Time {
	return time.Now().UTC()
}

// SlugGenerator is used to create new slugs from a given string.
// It can be customized with language codes, see github.com/gosimple/slug.
type SlugGenerator struct {
	Lang string
}

// NewSlugGenerator returns a new SlugGenerator given the language to be used.
func NewSlugGenerator(lang string) *SlugGenerator {
	return &SlugGenerator{lang}
}

// GenSlug generates a slug string for s, given the langugage of the generator.
func (gen SlugGenerator) GenSlug(s string) string {
	return slug.MakeLang(s, gen.Lang)
}
