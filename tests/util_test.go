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

package tests

import (
	"github.com/FabianWe/pollsweb"
	"github.com/google/uuid"
	"testing"
)

func TestGenUUID(t *testing.T) {
	// we can't really test the actual outcome, we just make sure it does not return an error
	// and returns a valid UUID
	id, idErr := pollsweb.GenUUID()
	if idErr != nil {
		t.Fatalf("pollsweb.GenUUID should not return an error, got %s", idErr)
	}
	parsedID, parseErr := uuid.Parse(id.String())
	if parseErr != nil {
		t.Fatalf("generated UUID should be parsable, but got error %s", parseErr)
	}
	if id != parsedID {
		t.Fatalf("generated and parsed uuid must be identical! expected %s but got %s",
			id, parsedID)
	}
}
