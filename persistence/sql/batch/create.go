// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package batch

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/jmoiron/sqlx/reflectx"

	"github.com/ory/x/dbal"

	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
	"github.com/pkg/errors"

	"github.com/ory/x/otelx"
	"github.com/ory/x/sqlcon"

	"github.com/ory/x/sqlxx"
)

type (
	insertQueryArgs struct {
		TableName    string
		ColumnsDecl  string
		Columns      []string
		Placeholders string
	}
	quoter interface {
		Quote(key string) string
	}
	TracerConnection struct {
		Tracer     *otelx.Tracer
		Connection *pop.Connection
	}
)

func buildInsertQueryArgs[T any](ctx context.Context, dialect string, mapper *reflectx.Mapper, quoter quoter, models []*T) (insertQueryArgs, error) {
	var (
		v     T
		model = pop.NewModel(v, ctx)

		columns        []string
		quotedColumns  []string
		placeholders   []string
		placeholderRow []string
	)

	for _, col := range model.Columns().Cols {
		columns = append(columns, col.Name)
		placeholderRow = append(placeholderRow, "?")
	}

	// We sort for the sole reason that the test snapshots are deterministic.
	sort.Strings(columns)

	for _, col := range columns {
		quotedColumns = append(quotedColumns, quoter.Quote(col))
	}
	for _, model := range models {
		pl := make([]string, len(placeholderRow))
		copy(pl, placeholderRow)
		for k := range placeholderRow {
			if columns[k] != "id" || dialect != dbal.DriverCockroachDB {
				continue
			}

			m := reflect.ValueOf(model)
			el := reflect.ValueOf(model).Elem()
			fbn := el.FieldByName("ID")
			if !fbn.IsValid() {
				return insertQueryArgs{}, errors.New("model does not have a field named id")
			}

			field := mapper.FieldByName(m, "ID")
			val, ok := field.Interface().(uuid.UUID)
			if !ok {
				continue
			}

			if val != uuid.Nil {
				pl[k] = "gen_random_uuid()"
				break
			}
		}

		placeholders = append(placeholders, fmt.Sprintf("(%s)", strings.Join(pl, ", ")))
	}

	return insertQueryArgs{
		TableName:    quoter.Quote(model.TableName()),
		ColumnsDecl:  strings.Join(quotedColumns, ", "),
		Columns:      columns,
		Placeholders: strings.Join(placeholders, ",\n"),
	}, nil
}

func buildInsertQueryValues[T any](dialect string, mapper *reflectx.Mapper, columns []string, models []*T, nowFunc func() time.Time) (values []any, err error) {
	for _, m := range models {
		m := reflect.ValueOf(m)

		now := nowFunc()
		// Append model fields to args
		for _, c := range columns {
			field := mapper.FieldByName(m, c)

			switch c {
			case "created_at":
				if pop.IsZeroOfUnderlyingType(field.Interface()) {
					field.Set(reflect.ValueOf(now))
				}
			case "updated_at":
				field.Set(reflect.ValueOf(now))
			case "id":
				if field.Interface().(uuid.UUID) != uuid.Nil {
					break // breaks switch, not for
				} else if dialect == dbal.DriverCockroachDB {
					continue // We do not need a value for this column because it is set automatically
				}

				id, err := uuid.NewV4()
				if err != nil {
					return nil, err
				}
				field.Set(reflect.ValueOf(id))
			}

			values = append(values, field.Interface())

			// Special-handling for *sqlxx.NullTime: mapper.FieldByName sets this to a zero time.Time,
			// but we want a nil pointer instead.
			if i, ok := field.Interface().(*sqlxx.NullTime); ok {
				if time.Time(*i).IsZero() {
					field.Set(reflect.Zero(field.Type()))
				}
			}
		}
	}

	return values, nil
}

// Create batch-inserts the given models into the database using a single INSERT statement.
// The models are either all created or none.
func Create[T any](ctx context.Context, p *TracerConnection, models []*T) (err error) {
	ctx, span := p.Tracer.Tracer().Start(ctx, "persistence.sql.batch.Create")
	defer otelx.End(span, &err)

	if len(models) == 0 {
		return nil
	}

	var v T
	model := pop.NewModel(v, ctx)

	conn := p.Connection
	quoter, ok := conn.Dialect.(quoter)
	if !ok {
		return errors.Errorf("store is not a quoter: %T", conn.Store)
	}

	queryArgs, err := buildInsertQueryArgs(ctx, conn.Dialect.Name(), conn.TX.Mapper, quoter, models)
	if err != nil {
		return err
	}

	values, err := buildInsertQueryValues(conn.Dialect.Name(), conn.TX.Mapper, queryArgs.Columns, models, func() time.Time { return time.Now().UTC().Truncate(time.Microsecond) })
	if err != nil {
		return err
	}

	var returningClause string
	if conn.Dialect.Name() != dbal.DriverMySQL {
		// PostgreSQL, CockroachDB, SQLite support RETURNING.
		returningClause = fmt.Sprintf("RETURNING %s", model.IDField())
	}

	query := conn.Dialect.TranslateSQL(fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES\n%s\n%s",
		queryArgs.TableName,
		queryArgs.ColumnsDecl,
		queryArgs.Placeholders,
		returningClause,
	))

	rows, err := conn.TX.QueryContext(ctx, query, values...)
	if err != nil {
		return sqlcon.HandleError(err)
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		if err := rows.Err(); err != nil {
			return sqlcon.HandleError(err)
		}

		if err := setModelID(rows, pop.NewModel(models[count], ctx)); err != nil {
			return err
		}
		count++
	}

	if err := rows.Err(); err != nil {
		return sqlcon.HandleError(err)
	}

	if err := rows.Close(); err != nil {
		return sqlcon.HandleError(err)
	}

	return sqlcon.HandleError(err)
}

func setModelID(row *sql.Rows, model *pop.Model) error {
	el := reflect.ValueOf(model.Value).Elem()
	fbn := el.FieldByName("ID")
	if !fbn.IsValid() {
		return errors.New("model does not have a field named id")
	}

	pkt, err := model.PrimaryKeyType()
	if err != nil {
		return errors.WithStack(err)
	}

	switch pkt {
	case "UUID":
		var id uuid.UUID
		if err := row.Scan(&id); err != nil {
			return errors.WithStack(err)
		}
		fbn.Set(reflect.ValueOf(id))
	default:
		var id interface{}
		if err := row.Scan(&id); err != nil {
			return errors.WithStack(err)
		}
		v := reflect.ValueOf(id)
		switch fbn.Kind() {
		case reflect.Int, reflect.Int64:
			fbn.SetInt(v.Int())
		default:
			fbn.Set(reflect.ValueOf(id))
		}
	}

	return nil
}
