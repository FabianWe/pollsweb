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
	"context"
	"errors"
	"github.com/FabianWe/pollsweb"
	"github.com/google/uuid"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/nleof/goyesql"
	"time"
)

func convertToPGXUUID(uuid uuid.UUID) interface{} {
	internal := pgtype.UUID{
		Bytes:  [16]byte(uuid),
		Status: pgtype.Present,
	}
	return internal
}

func getPGXUUIDScanType() pgtype.UUID {
	return pgtype.UUID{
		Bytes:  [16]byte{},
		Status: pgtype.Null,
	}
}

func assignToUUID(pgxUUID pgtype.UUID) uuid.UUID {
	asBytes := pgxUUID.Bytes
	return asBytes
}

type PostgresPeriodDataProvider struct {
	AppContext *pollsweb.AppContext
	Tx         pgx.Tx
	Queries    goyesql.Queries
}

func NewPostgresPeriodDataProvider(appContext *pollsweb.AppContext, tx pgx.Tx, queries goyesql.Queries) *PostgresPeriodDataProvider {
	return &PostgresPeriodDataProvider{
		AppContext: appContext,
		Tx:         tx,
		Queries:    queries,
	}
}

func (pg *PostgresPeriodDataProvider) InsertPeriod(ctx context.Context, period *PeriodModel) error {
	// use query to store entry
	query := pg.Queries["period_add"]
	internalUUID := convertToPGXUUID(period.ID)
	_, insertErr := pg.Tx.Exec(ctx, query, internalUUID, period.Name, period.Slug, period.MeetingTime,
		period.PeriodStart, period.PeriodEnd)
	return insertErr
}

func (pg *PostgresPeriodDataProvider) ScanPeriod(row pgx.Row) (*PeriodModel, error) {
	// create all variables that are required
	pgxID := getPGXUUIDScanType()
	var name, slug string
	var created, meetingTime, periodStart, periodEnd time.Time
	scanErr := row.Scan(&pgxID, &name, &slug, &meetingTime, &periodStart, &periodEnd, &created)
	if scanErr != nil {
		if errors.Is(scanErr, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, scanErr
	}
	id := assignToUUID(pgxID)
	period := PeriodModel{
		ID:          id,
		Name:        name,
		Slug:        slug,
		MeetingTime: meetingTime,
		PeriodStart: periodStart,
		PeriodEnd:   periodEnd,
		Created:     created,
	}

	return &period, nil
}

func (pg *PostgresPeriodDataProvider) GetPeriodByID(ctx context.Context, id uuid.UUID) (*PeriodModel, error) {
	query := pg.Queries["period_get_by_id"]
	return pg.ScanPeriod(pg.Tx.QueryRow(ctx, query, convertToPGXUUID(id)))
}

func (pg *PostgresPeriodDataProvider) GetPeriodBySlug(ctx context.Context, slug string) (*PeriodModel, error) {
	query := pg.Queries["period_get_by_slug"]
	return pg.ScanPeriod(pg.Tx.QueryRow(ctx, query, slug))
}

func (pg *PostgresPeriodDataProvider) GetLatestPeriod(ctx context.Context) (*PeriodModel, error) {
	query := pg.Queries["period_get_latest"]
	return pg.ScanPeriod(pg.Tx.QueryRow(ctx, query))
}

func (pg *PostgresPeriodDataProvider) GetLatestNPeriods(ctx context.Context, n int) ([]*PeriodModel, error) {
	res := make([]*PeriodModel, 0, n)
	query := pg.Queries["period_get_latest_n"]
	rows, err := pg.Tx.Query(ctx, query, n)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		period, scanErr := pg.ScanPeriod(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		if period == nil {
			panic("internal error: scanned period should not be nil")
		}
		res = append(res, period)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return res, nil
}
