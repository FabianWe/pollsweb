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
	"github.com/FabianWe/pollsweb"
	"time"
)
import "github.com/google/uuid"

type MeetingTimeTemplateModel struct {
	Weekday time.Weekday
	Hour uint8
	Minute uint8
}

func EmptyMeetingTimeTemplateModel() *MeetingTimeTemplateModel {
	return &MeetingTimeTemplateModel{
		Weekday: 0,
		Hour:    0,
		Minute:  0,
	}
}

func NewMeetingTimeTemplateModel(weekday time.Weekday, hour, minute uint8) *MeetingTimeTemplateModel {
	return &MeetingTimeTemplateModel{
		Weekday: weekday,
		Hour:    hour,
		Minute:  minute,
	}
}

func (m *MeetingTimeTemplateModel) String() string {
	return fmt.Sprintf("MeetingTimeTemplateModel(Weekday=%d, Hour=%d, Minute=%d)", m.Weekday, m.Hour, m.Minute)
}

type PeriodSettingsModel struct {
	Id uuid.UUID `bson:"_id"`
	Name string
	MeetingDateTemplate *MeetingTimeTemplateModel `bson:"time"`
	Start time.Time
	End time.Time
	Created time.Time
}

func EmptyPeriodSettingsModel() *PeriodSettingsModel {
	return &PeriodSettingsModel{
		Id:                  uuid.Nil,
		Name:                "",
		MeetingDateTemplate: EmptyMeetingTimeTemplateModel(),
		Start:               time.Time{},
		End:                 time.Time{},
		Created:             time.Time{},
	}
}

func NewPeriodSettingsModel(name string, meetingDateTemplate *MeetingTimeTemplateModel, start, end time.Time) *PeriodSettingsModel {
	now := pollsweb.UTCNow()
	return &PeriodSettingsModel{
		Id:                  uuid.Nil,
		Name:                name,
		MeetingDateTemplate: meetingDateTemplate,
		Start:               start,
		End:                 end,
		Created:             now,
	}
}

func (m *PeriodSettingsModel) String() string {
	return fmt.Sprintf("PeriodSettingsModel(Id=%s, Name=%s, MettingDateTemplate=%s, Start=%s, End=%s, Created=%s)",
		m.Id, m.Name, m.MeetingDateTemplate, m.Start, m.End, m.Created)
}
