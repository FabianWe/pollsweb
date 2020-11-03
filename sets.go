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
	"strings"
)

// StringSet ia a set of strings.
// Concurrent calls to Add are not allowed.
type StringSet map[string]struct{}

func NewStringSet(initialSize int) StringSet {
	if initialSize <= 0 {
		return make(StringSet)
	}
	return make(StringSet, initialSize)
}

// Add adds an element to the set, it returns true of the string didn't exist before and false otherwise.
func (s StringSet) Add(e string) bool {
	oldLen := len(s)
	s[e] = struct{}{}
	return oldLen != len(s)
}

func (s StringSet) Contains(e string) bool {
	_, has := s[e]
	return has
}

func (s StringSet) String() string {
	var buf strings.Builder

	buf.WriteRune('{')
	first := true
	for e, _ := range s {
		if first {
			first = false
		} else {
			buf.WriteString(", ")
		}
		buf.WriteString(e)
	}
	buf.WriteRune('}')
	return buf.String()
}

// UUIDSet is a set of uuids.
// Concurrent calls to Add are not allowed.
type UUIDSet map[uuid.UUID]struct{}

func NewUUIDSet(initialSize int) UUIDSet {
	if initialSize <= 0 {
		return make(UUIDSet)
	}
	return make(UUIDSet, initialSize)
}

func (s UUIDSet) Add(e uuid.UUID) bool {
	oldLen := len(s)
	s[e] = struct{}{}
	return oldLen != len(s)
}

func (s UUIDSet) Contains(e uuid.UUID) bool {
	_, has := s[e]
	return has
}

func (s UUIDSet) String() string {
	var buf strings.Builder

	buf.WriteRune('{')
	first := true
	for e, _ := range s {
		if first {
			first = false
		} else {
			buf.WriteString(", ")
		}
		buf.WriteString(e.String())
	}
	buf.WriteRune('}')
	return buf.String()
}
