package errorx

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"

	"github.com/ory/x/sqlcon"
	"github.com/ory/x/sqlxx"

	"github.com/google/uuid"

	"github.com/ory/herodot"
)

var _ Manager = new(ManagerSQL)

type (
	containerSQL struct {
		PK      uint64    `db:"pk"`
		ID      string    `db:"id"`
		Errors  string    `db:"errors"`
		SeenAt  time.Time `db:"seen_at"`
		WasSeen bool      `db:"was_seen"`
	}

	ManagerSQL struct {
		db *sqlx.DB
		*BaseManager
	}
)

func NewManagerSQL(
	db *sqlx.DB,
	d baseManagerDependencies,
	c baseManagerConfiguration,
) *ManagerSQL {
	m := &ManagerSQL{db: db}
	m.BaseManager = NewBaseManager(d, c, m)
	return m
}

func (m *ManagerSQL) Add(ctx context.Context, errs ...error) (string, error) {
	b, err := m.encode(errs)
	if err != nil {
		return "", err
	}

	container := &containerSQL{ID: uuid.New().String(), Errors: b.String()}
	columns, arguments := sqlxx.NamedInsertArguments(container, "pk", "seen_at")
	query := fmt.Sprintf(`INSERT INTO self_service_error (%s) VALUES (%s)`, columns, arguments)
	if _, err := m.db.NamedExecContext(
		ctx,
		query,
		container,
	); err != nil {
		return "", errors.WithStack(herodot.ErrInternalServerError.WithReason("Unable to store error messages in SQL datastore.").WithDebug(err.Error()))
	}

	return container.ID, nil
}

func (m *ManagerSQL) Read(ctx context.Context, id string) ([]json.RawMessage, error) {
	var c string

	tx, err := m.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, sqlcon.HandleError(err)
	}

	if err := tx.GetContext(context.Background(), &c, m.db.Rebind("SELECT errors FROM self_service_error WHERE id=?"), id); err != nil {
		if err := tx.Rollback(); err != nil {
			return nil, sqlcon.HandleError(err)
		}
		return nil, sqlcon.HandleError(err)
	}

	var errs []json.RawMessage
	if err := json.NewDecoder(bytes.NewBufferString(c)).Decode(&errs); err != nil {
		if err := tx.Rollback(); err != nil {
			return nil, sqlcon.HandleError(err)
		}
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReason("Unable to decode stored error messages from SQL datastore.").WithDebug(err.Error()))
	}

	if _, err := tx.ExecContext(context.Background(), m.db.Rebind("UPDATE self_service_error SET was_seen = true, seen_at = ? WHERE id = ?"), time.Now().UTC(), id); err != nil {
		if err := tx.Rollback(); err != nil {
			return nil, sqlcon.HandleError(err)
		}
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReason("Unable to update seen status for error message in SQL datastore.").WithDebug(err.Error()))
	}

	if err := tx.Commit(); err != nil {
		return nil, sqlcon.HandleError(err)
	}

	return errs, nil
}

func (m *ManagerSQL) Clear(ctx context.Context, olderThan time.Duration, force bool) (err error) {
	if force {
		_, err = m.db.ExecContext(ctx, m.db.Rebind("DELETE FROM self_service_error WHERE seen_at < ?"), olderThan)
	} else {
		_, err = m.db.ExecContext(ctx, m.db.Rebind("DELETE FROM self_service_error WHERE was_seen=true AND seen_at < ?"), time.Now().UTC().Add(-olderThan))
	}

	return sqlcon.HandleError(err)
}
