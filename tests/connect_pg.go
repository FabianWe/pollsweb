// +build integration

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
	"context"
	"flag"
	"fmt"
	"github.com/jackc/pgx/v4"
	"time"
)

var (
	pgHost     = flag.String("pg-host", "localhost", "host of postgres to connect to (default \"localhost\")")
	pgPort     = flag.Int("pg-port", 5432, "port of postgres to connect to, default 5432")
	pgUser     = flag.String("pg-user", "root", "name of the postgres user, default \"root\"")
	pgPassword = flag.String("pg-password", "password", "password for the postgres connection, default \"password\"")
	pgDatabase = flag.String("pg-db", "gopolls-test", "name of the database to connect to, default \"gopolls-test\"")
	pgSSLMode  = flag.String("pg-ssl-mode", "disable", "postgres SSL mode, default \"disable\"")
	pgTimeout  = flag.Duration("pg-timeout", 30*time.Second, "timeout for postgres operations, default is 30s")
)

func prepareTestRevisions() {

}

func preparePGDB() *pgx.Conn {
	connectionStr := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		*pgUser,
		*pgPassword,
		*pgHost,
		*pgPort,
		*pgDatabase,
		*pgSSLMode)
	fmt.Println(*pgPassword)
	ctx, cancel := context.WithTimeout(context.Background(), *pgTimeout)
	defer cancel()
	conn, connErr := pgx.Connect(ctx, connectionStr)
	if connErr != nil {
		panic(connErr)
	}
	return conn
}
