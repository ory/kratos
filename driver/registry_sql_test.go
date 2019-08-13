/*
 * Copyright © 2015-2018 Aeneas Rekkas <aeneas+oss@aeneas.io>
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * @author		Aeneas Rekkas <aeneas+oss@aeneas.io>
 * @Copyright 	2017-2018 Aeneas Rekkas <aeneas+oss@aeneas.io>
 * @license 	Apache-2.0
 */

package driver

import (
	"context"
	"fmt"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"

	"github.com/ory/x/dbal"
	"github.com/ory/x/dbal/migratest"

	"github.com/ory/hive/selfservice"
)

func TestXXMigrations(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
		return
	}

	migratest.RunPackrMigrationTests(
		t,
		migratest.MigrationSchemas{Migrations},
		migratest.MigrationSchemas{dbal.FindMatchingTestMigrations("../contrib/sql/migrations/tests/", Migrations, AssetNames(), Asset)},
		SQLPurgeTestDatabase, SQLPurgeTestDatabase,
		func(t *testing.T, dbName string, db *sqlx.DB, _, step, steps int) {
			id := fmt.Sprintf("%d-data", step+1)
			t.Run("poll="+id, func(t *testing.T) {
				t.Run("service=selfservice.NewRequestManagerSQL", func(t *testing.T) {
					m := selfservice.NewRequestManagerSQL(db, requestManagerFactories)
					_, err := m.GetLoginRequest(context.Background(), "1")
					require.NoError(t, err)
				})
			})
		},
	)
}
