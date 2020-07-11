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
	"github.com/FabianWe/gopolls"
	"github.com/FabianWe/pollsweb/server"
	"testing"
	"time"
)

func TestDecodeHourMinuteFormField(t *testing.T) {
	type testCase struct {
		in          string
		expectedObj server.HourMinuteFormField
		expectsErr  bool
	}
	newTestCase := func(in string, expectedObj server.HourMinuteFormField, expectsErr bool) testCase {
		return testCase{
			in,
			expectedObj,
			expectsErr,
		}
	}
	// this type just wraps a *HourMinuteFormField we want to decode
	// we use this to use the DecodeForm method
	type formDummy struct {
		Time server.HourMinuteFormField
	}
	tests := []testCase{
		newTestCase("20:15", server.HourMinuteFormField{20, 15}, false),
		newTestCase("09:21", server.HourMinuteFormField{9, 21}, false),
		newTestCase("08:42", server.HourMinuteFormField{8, 42}, false),
		newTestCase("21", server.HourMinuteFormField{}, true),
		newTestCase(":42", server.HourMinuteFormField{}, true),
		newTestCase("21:a", server.HourMinuteFormField{}, true),
		newTestCase("8:42", server.HourMinuteFormField{}, true),
		newTestCase("21:8", server.HourMinuteFormField{}, true),
		newTestCase("24:42", server.HourMinuteFormField{}, true),
		newTestCase("21:60", server.HourMinuteFormField{}, true),
		newTestCase("-21:42", server.HourMinuteFormField{}, true),
		newTestCase("21:-42", server.HourMinuteFormField{}, true),
	}
	for _, tc := range tests {
		form := &formDummy{server.HourMinuteFormField{}}
		m := map[string][]string{"time": {tc.in}}
		tcErr := server.DecodeForm(form, m)
		if tc.expectsErr {
			if tcErr == nil {
				t.Errorf("expected error, for input \"%s\" but no error was returned", tc.in)
			}
		} else {
			if tcErr != nil {
				t.Errorf("expected no error for input \"%s\", but got error %v", tc.in, tcErr)
				continue
			}
			// compare
			if !form.Time.Equals(tc.expectedObj) {
				t.Errorf("expected %s but got %s for input \"%s\"", tc.expectedObj, form.Time, tc.in)
			}
		}
	}
}

func TestDecodeDateFormField(t *testing.T) {
	emptyDate := server.DateFormField(time.Time{})
	type formDummy struct {
		Date server.DateFormField
	}
	tests := []struct {
		in         string
		expected   server.DateFormField
		expectsErr bool
	}{
		{"2020/07/09", server.NewDateFormField(2020, 7, 9), false},
		{"1999/10/21", server.NewDateFormField(1999, 10, 21), false},
		{"20/07/09", emptyDate, true},
		{"2020/42/09", emptyDate, true},
		{"2020/07/42", emptyDate, true},
		{"2020/07/09 21:42", emptyDate, true},
	}
	for _, tc := range tests {
		form := &formDummy{emptyDate}
		m := map[string][]string{"date": {tc.in}}
		tcErr := server.DecodeForm(form, m)
		if tc.expectsErr {
			if tcErr == nil {
				t.Errorf("expected error for input \"%s\", but no error was returned", tc.in)
			}
		} else {
			if tcErr != nil {
				t.Errorf("expected no error for input \"%s\", but got error %v", tc.in, tcErr)
				continue
			}
			// compare
			if !form.Date.Equals(tc.expected) {
				t.Errorf("expected %s but got %s for input \"%s\"", tc.expected, form.Date, tc.in)
			}
			// ensure location is set to UTC
			if loc := time.Time(form.Date).Location(); loc != time.UTC {
				t.Errorf("expected location to be set to UTC, got %s", loc)
			}
		}
	}
}

func TestDecodeDateTimeFormField(t *testing.T) {
	emptyDt := server.DateTimeFormField(time.Time{})
	type formDummy struct {
		Time server.DateTimeFormField
	}
	tests := []struct {
		in         string
		expected   server.DateTimeFormField
		expectsErr bool
	}{
		{"2020/07/09 15:21", server.NewDateTimeFormField(2020, 7, 9, 15, 21), false},
		{"1999/10/21 21:09", server.NewDateTimeFormField(1999, 10, 21, 21, 9), false},
		{"1999/10/21 42:09", emptyDt, true},
		{"1999/10/21 21:99", emptyDt, true},
		{"20/07/09 15:21", emptyDt, true},
		{"2020/42/09 15:21", emptyDt, true},
		{"2020/07/42 15:21", emptyDt, true},
	}
	for _, tc := range tests {
		form := &formDummy{emptyDt}
		m := map[string][]string{"time": {tc.in}}
		tcErr := server.DecodeForm(form, m)
		if tc.expectsErr {
			if tcErr == nil {
				t.Errorf("expected error for input \"%s\", but no error was returned", tc.in)
			}
		} else {
			if tcErr != nil {
				t.Errorf("expected no error for input \"%s\", but got error %v", tc.in, tcErr)
				continue
			}
			// compare
			if !form.Time.Equals(tc.expected) {
				t.Errorf("expected %s but got %s for input \"%s\"", tc.expected, form.Time, tc.in)
			}
			// ensure location is set to UTC
			if loc := time.Time(form.Time).Location(); loc != time.UTC {
				t.Errorf("expected location to be set to UTC, got %s", loc)
			}
		}
	}
}

func TestDecodeWeekdayFormField(t *testing.T) {
	dummyWeekday := server.WeekdayFormField(-1)
	type formDummy struct {
		Weekday server.WeekdayFormField
	}
	tests := []struct {
		in         string
		expected   server.WeekdayFormField
		expectsErr bool
	}{
		{"0", server.WeekdayFormField(time.Sunday), false},
		{"1", server.WeekdayFormField(time.Monday), false},
		{"2", server.WeekdayFormField(time.Tuesday), false},
		{"3", server.WeekdayFormField(time.Wednesday), false},
		{"4", server.WeekdayFormField(time.Thursday), false},
		{"5", server.WeekdayFormField(time.Friday), false},
		{"6", server.WeekdayFormField(time.Saturday), false},
		{"7", dummyWeekday, true},
		{"42", dummyWeekday, true},
		{"abcd", dummyWeekday, true},
	}
	for _, tc := range tests {
		form := &formDummy{dummyWeekday}
		m := map[string][]string{"weekday": {tc.in}}
		tcErr := server.DecodeForm(form, m)
		if tc.expectsErr {
			if tcErr == nil {
				t.Errorf("expected error for input \"%s\", but no error was returned", tc.in)
			}
		} else {
			if tcErr != nil {
				t.Errorf("expected no error for input \"%s\", but got error %v", tc.in, tcErr)
				continue
			}
			// compare
			if form.Weekday != tc.expected {
				t.Errorf("expected %s but got %s for input \"%s\"", tc.expected, form.Weekday, tc.in)
			}
		}
	}
}

func TestDecodeVotersFormField(t *testing.T) {
	// first create a parser and register the converter in the schemaDecoder
	parser := gopolls.NewVotersParser()
	parser.MaxVotersWeight = 5
	parser.MaxVotersNameLength = 10
	parser.MaxNumVoters = 2

	schemaDecoder := server.NewSchemaDecoder()
	schemaDecoder.RegisterConverter(server.NewVotersFormField(nil), server.GetVotersFormFieldConverter(parser))
	decoder := server.NewFormDecoder()
	decoder.SchemaDecoder = schemaDecoder
	type formDummy struct {
		Voters server.VotersFormField `schema:"voters" valid:"-"`
	}
	tests := []struct {
		in             string
		expectedLength int
	}{
		{
			`* foo: 1
* bar: 5`,
			2,
		},
		{
			`# comment

* foo:1
`,
			1,
		},
		{"", 0},
		{
			"invalid!!!",
			-1,
		},
		{
			"* foo: 6",
			-1,
		},
		{
			`* foo: 1
* bar: 1
* foobar: 1`,
			-1,
		},
		{"* toolongtoolong: 1", -1},
	}
	for _, tc := range tests {
		form := &formDummy{server.NewVotersFormField(nil)}
		m := map[string][]string{"voters": {tc.in}}
		tcErr := decoder.NormalizeAndDecodeForm(form, m)
		if tc.expectedLength < 0 {
			// this means we expect an error
			if tcErr == nil {
				t.Errorf("expected error for input \"%s\", but no error was returned", tc.in)
			}
		} else {
			// no error expected
			if tcErr != nil {
				t.Errorf("expected no error for input \"%s\", but got error %v", tc.in, tcErr)
				continue
			}
			// compare length
			if len(form.Voters.Voters) != tc.expectedLength {
				t.Errorf("expected %d voters for input \"%s\", but got %d results (%v)",
					tc.expectedLength, tc.in, len(form.Voters.Voters), form.Voters.Voters)
			}
		}
	}
}
