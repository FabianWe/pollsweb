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

type UUIDGenError struct {
	PollWebError
	Wrapped error
}

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

func GenUUID() (uuid.UUID, error) {
	res, err := uuid.NewRandom()
	if err != nil {
		return res, NewUUIDGenError(err)
	}
	return res, nil
}

func GenNow() time.Time {
	return time.Now().UTC()
}

type SlugGenerator struct {
	Lang string
}

func NewSlugGenerator(lang string) *SlugGenerator {
	return &SlugGenerator{lang}
}

func (gen SlugGenerator) GenSlug(s string) string {
	return slug.MakeLang(s, gen.Lang)
}
