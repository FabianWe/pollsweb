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
)

// EntryNotFoundError is an error returned if an entry could not be found in the database.
//
// It embeds PollWebError and is thus an internal error.
// It can wrap another error (for example the original database error), Wrapped can also be nil.
// It also references the key for which the lookup failed by wrapping it in a reflect.Value.
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
	return fmt.Sprintf("entry of type \"%v\" not found, key %v: \"%v\" does not exist",
		e.Model, e.KeyValue.Type(), e.KeyValue)
}

func (e EntryNotFoundError) Unwrap() error {
	return e.Wrapped
}

type PeriodSettingsHandler interface {
	GetByName(ctx context.Context, name string) (*PeriodSettingsModel, error)
	GetByID(ctx context.Context, id uuid.UUID) (*PeriodSettingsModel, error)

	Insert(ctx context.Context, meetingTime *PeriodSettingsModel) (uuid.UUID, error)
}
