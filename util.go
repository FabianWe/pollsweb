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
	"golang.org/x/text/unicode/norm"
	"strings"
	"sync"
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

// UTCNow returns the current time in UTC.
// For consistent usage this function should always be called to generate the current time.
func UTCNow() time.Time {
	return time.Now().UTC()
}

type TimeFormatTranslator struct {
	once          *sync.Once
	replacer      *strings.Replacer
	YearLong      string
	YearShort     string
	LongMonthStr  string
	ShortMonthStr string
	NumMonthLong  string
	NumMonthShort string
	WeekdayLong   string
	WeekdayShort  string
	DayLong       string
	DayShort      string
	Hour24        string
	Hour12Long    string
	Hour12Short   string
	MinuteLong    string
	MinuteShort   string
	SecondLong    string
	SecondShort   string
	PMCapital     string
	PMLower       string
	TZ            string
	NumColonTZ    string
	NumTZLong     string
	NumTZShort    string
}

func NewTimeFormatTranslator() *TimeFormatTranslator {
	return &TimeFormatTranslator{
		once: &sync.Once{},
	}
}

func (f *TimeFormatTranslator) ConvertFormat(goFormat string) string {
	f.once.Do(func() {
		// note that the order matters: "2006" must be before "06" for example
		replaceList := []string{
			"2006", f.YearLong,
			"06", f.YearShort,
			"January", f.LongMonthStr,
			"Jan", f.ShortMonthStr,
			"15", f.Hour24, // this must be this high in the list because the "1" (numMonthShort) would replace it
			"01", f.NumMonthLong,
			"1", f.NumMonthShort,
			"Monday", f.WeekdayLong,
			"Mon", f.WeekdayShort,
			"02", f.DayLong,
			"2", f.DayShort,
			"03", f.Hour12Long,
			"3", f.Hour12Short,
			"04", f.MinuteLong,
			"4", f.MinuteShort,
			"05", f.SecondLong,
			"5", f.SecondShort,
			"PM", f.PMCapital,
			"pm", f.PMLower,
			"MST", f.TZ,
			"-07:00", f.NumColonTZ,
			"-0700", f.NumTZLong,
			"-07", f.NumTZShort,
		}
		f.replacer = strings.NewReplacer(replaceList...)
	})
	return f.replacer.Replace(goFormat)
}

var MomentJSDateFormatter *TimeFormatTranslator

func init() {
	MomentJSDateFormatter = &TimeFormatTranslator{
		once:          &sync.Once{},
		replacer:      nil,
		YearLong:      "YYYY",
		YearShort:     "YY",
		LongMonthStr:  "MMMM",
		ShortMonthStr: "MMM",
		NumMonthLong:  "MM",
		NumMonthShort: "M",
		WeekdayLong:   "dddd",
		WeekdayShort:  "ddd",
		DayLong:       "DD",
		DayShort:      "D",
		Hour24:        "HH",
		Hour12Long:    "hh",
		Hour12Short:   "h",
		MinuteLong:    "mm",
		MinuteShort:   "m",
		SecondLong:    "ss",
		SecondShort:   "s",
		PMCapital:     "A",
		PMLower:       "a",
		TZ:            "zz",
		NumColonTZ:    "Z",
		NumTZLong:     "ZZ",
		NumTZShort:    "ZZ", // not really supported
	}
}

// TODO remove
var Utf8Form = norm.NFKC

func NormalizeString(s string) string {
	return Utf8Form.String(s)
}
