// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package update_test

import (
	"testing"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace/noop"

	"github.com/ory/kratos/persistence/sql/update"
	"github.com/ory/pop/v6"
	"github.com/ory/x/dbal"
)

type updateTestModel struct {
	ID        uuid.UUID `db:"id"`
	NID       uuid.UUID `db:"nid"`
	Name      string    `db:"name"`
	UniqueVal string    `db:"unique_val"`
}

func (updateTestModel) TableName() string { return "update_test_models" }

func newUpdateTestConn(t *testing.T) *pop.Connection {
	t.Helper()

	c, err := pop.NewConnection(&pop.ConnectionDetails{URL: dbal.NewSQLiteTestDatabase(t)})
	require.NoError(t, err)
	require.NoError(t, c.Open())
	t.Cleanup(func() { _ = c.Close() })

	require.NoError(t, c.RawQuery(`CREATE TABLE update_test_models (
		id TEXT PRIMARY KEY,
		nid TEXT NOT NULL,
		name TEXT NOT NULL,
		unique_val TEXT NOT NULL
	)`).Exec())
	require.NoError(t, c.RawQuery(
		`CREATE UNIQUE INDEX update_test_models_nid_unique_val ON update_test_models (nid, unique_val)`).Exec())

	return c
}

func TestGeneric(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	tracer := noop.NewTracerProvider().Tracer("")

	seed := func(t *testing.T, c *pop.Connection) (nid, id uuid.UUID) {
		nid = uuid.Must(uuid.NewV4())
		id = uuid.Must(uuid.NewV4())
		require.NoError(t, c.Create(&updateTestModel{
			ID: id, NID: nid, Name: "original", UniqueVal: "original-unique",
		}))
		return nid, id
	}

	reload := func(t *testing.T, c *pop.Connection, nid, id uuid.UUID) updateTestModel {
		var m updateTestModel
		require.NoError(t, c.Where("id = ? AND nid = ?", id, nid).First(&m))
		return m
	}

	t.Run("case=GenericExcept leaves the excluded column untouched", func(t *testing.T) {
		t.Parallel()
		c := newUpdateTestConn(t)
		nid, id := seed(t, c)

		// The in-memory model carries a different unique_val, but excluding it
		// must keep the column out of the SET list so the stored value stays.
		m := &updateTestModel{ID: id, NID: nid, Name: "changed", UniqueVal: "must-not-be-written"}
		require.NoError(t, update.GenericExcept(ctx, c, tracer, m, "unique_val"))

		got := reload(t, c, nid, id)
		assert.Equal(t, "changed", got.Name, "a non-excluded column is written")
		assert.Equal(t, "original-unique", got.UniqueVal, "the excluded column is left untouched")
	})

	t.Run("case=Generic writes every column including the unique one", func(t *testing.T) {
		t.Parallel()
		c := newUpdateTestConn(t)
		nid, id := seed(t, c)

		m := &updateTestModel{ID: id, NID: nid, Name: "changed", UniqueVal: "now-written"}
		require.NoError(t, update.Generic(ctx, c, tracer, m))

		got := reload(t, c, nid, id)
		assert.Equal(t, "changed", got.Name)
		assert.Equal(t, "now-written", got.UniqueVal, "Generic writes the unique column")
	})

	t.Run("case=Generic with an explicit column list writes only those columns", func(t *testing.T) {
		t.Parallel()
		c := newUpdateTestConn(t)
		nid, id := seed(t, c)

		m := &updateTestModel{ID: id, NID: nid, Name: "changed", UniqueVal: "must-not-be-written"}
		require.NoError(t, update.Generic(ctx, c, tracer, m, "name"))

		got := reload(t, c, nid, id)
		assert.Equal(t, "changed", got.Name)
		assert.Equal(t, "original-unique", got.UniqueVal, "a column absent from the list is left untouched")
	})

	t.Run("case=returns ErrNoRows when the row does not exist", func(t *testing.T) {
		t.Parallel()
		c := newUpdateTestConn(t)

		m := &updateTestModel{ID: uuid.Must(uuid.NewV4()), NID: uuid.Must(uuid.NewV4()), Name: "x", UniqueVal: "y"}
		assert.Error(t, update.GenericExcept(ctx, c, tracer, m, "unique_val"))
		assert.Error(t, update.Generic(ctx, c, tracer, m))
	})
}
