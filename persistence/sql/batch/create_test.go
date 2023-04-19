// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package batch

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/jmoiron/sqlx/reflectx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/identity"
	"github.com/ory/x/snapshotx"
	"github.com/ory/x/sqlxx"
)

type (
	testModel struct {
		ID          uuid.UUID       `db:"id"`
		NID         uuid.UUID       `db:"nid"`
		String      string          `db:"string"`
		Int         int             `db:"int"`
		Traits      identity.Traits `db:"traits"`
		NullTimePtr *sqlxx.NullTime `db:"null_time_ptr"`
		CreatedAt   time.Time       `json:"created_at" db:"created_at"`
		UpdatedAt   time.Time       `json:"updated_at" db:"updated_at"`
	}
	testQuoter struct{ name string }
)

func (testModel) TableName() string {
	return "test_models"
}

func (tq testQuoter) Quote(s string) string { return fmt.Sprintf("%q", s) }

func (tq testQuoter) Name() string { return tq.name }

func Test_buildInsertQueryArgs(t *testing.T) {
	ctx := context.Background()

	t.Run("case=testModel/quoter=generic", func(t *testing.T) {
		models := make([]*testModel, 10)
		args := buildInsertQueryArgs(ctx, testQuoter{"generic"}, models)
		snapshotx.SnapshotT(t, args)

		query := fmt.Sprintf("INSERT INTO %s (%s) VALUES\n%s", args.TableName, args.ColumnsDecl, args.Placeholders)
		assert.Equal(t, `INSERT INTO "test_models" ("created_at", "id", "int", "nid", "null_time_ptr", "string", "traits", "updated_at") VALUES
(?, ?, ?, ?, ?, ?, ?, ?),
(?, ?, ?, ?, ?, ?, ?, ?),
(?, ?, ?, ?, ?, ?, ?, ?),
(?, ?, ?, ?, ?, ?, ?, ?),
(?, ?, ?, ?, ?, ?, ?, ?),
(?, ?, ?, ?, ?, ?, ?, ?),
(?, ?, ?, ?, ?, ?, ?, ?),
(?, ?, ?, ?, ?, ?, ?, ?),
(?, ?, ?, ?, ?, ?, ?, ?),
(?, ?, ?, ?, ?, ?, ?, ?)`, query)
	})

	t.Run("case=testModel/quoter=cockroach", func(t *testing.T) {
		models := make([]*testModel, 10)
		args := buildInsertQueryArgs(ctx, testQuoter{"cockroach"}, models)
		snapshotx.SnapshotT(t, args)

		query := fmt.Sprintf("INSERT INTO %s (%s) VALUES\n%s", args.TableName, args.ColumnsDecl, args.Placeholders)
		assert.Equal(t, `INSERT INTO "test_models" ("id", "created_at", "int", "nid", "null_time_ptr", "string", "traits", "updated_at") VALUES
(gen_random_uuid(), ?, ?, ?, ?, ?, ?, ?),
(gen_random_uuid(), ?, ?, ?, ?, ?, ?, ?),
(gen_random_uuid(), ?, ?, ?, ?, ?, ?, ?),
(gen_random_uuid(), ?, ?, ?, ?, ?, ?, ?),
(gen_random_uuid(), ?, ?, ?, ?, ?, ?, ?),
(gen_random_uuid(), ?, ?, ?, ?, ?, ?, ?),
(gen_random_uuid(), ?, ?, ?, ?, ?, ?, ?),
(gen_random_uuid(), ?, ?, ?, ?, ?, ?, ?),
(gen_random_uuid(), ?, ?, ?, ?, ?, ?, ?),
(gen_random_uuid(), ?, ?, ?, ?, ?, ?, ?)`, query)
	})

	for _, tc := range []testQuoter{{"generic"}, {"cockroach"}} {
		t.Run("case=Identities/quoter="+tc.name, func(t *testing.T) {
			models := make([]*identity.Identity, 10)
			args := buildInsertQueryArgs(ctx, tc, models)
			snapshotx.SnapshotT(t, args)
		})

		t.Run("case=RecoveryAddress/quoter="+tc.name, func(t *testing.T) {
			models := make([]*identity.RecoveryAddress, 10)
			args := buildInsertQueryArgs(ctx, tc, models)
			snapshotx.SnapshotT(t, args)
		})

		t.Run("case=RecoveryAddress/quoter="+tc.name, func(t *testing.T) {
			models := make([]*identity.RecoveryAddress, 10)
			args := buildInsertQueryArgs(ctx, tc, models)
			snapshotx.SnapshotT(t, args)
		})
	}
}

func Test_buildInsertQueryValues(t *testing.T) {
	t.Run("case=testModel", func(t *testing.T) {
		model := &testModel{
			String: "string",
			Int:    42,
			Traits: []byte(`{"foo": "bar"}`),
		}
		mapper := reflectx.NewMapper("db")
		values, err := buildInsertQueryValues(mapper, []string{"created_at", "updated_at", "id", "string", "int", "null_time_ptr", "traits"}, []*testModel{model})
		require.NoError(t, err)

		assert.NotNil(t, model.CreatedAt)
		assert.Equal(t, model.CreatedAt, values[0])

		assert.NotNil(t, model.UpdatedAt)
		assert.Equal(t, model.UpdatedAt, values[1])

		assert.NotNil(t, model.ID)
		assert.Equal(t, model.ID, values[2])

		assert.Equal(t, model.String, values[3])
		assert.Equal(t, model.Int, values[4])

		assert.Nil(t, model.NullTimePtr)
	})
}
