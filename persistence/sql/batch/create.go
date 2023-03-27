// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package batch

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
	"github.com/jmoiron/sqlx/reflectx"
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

func buildInsertQueryArgs[T any](ctx context.Context, quoter quoter, models []*T) insertQueryArgs {
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
	for range models {
		placeholders = append(placeholders, fmt.Sprintf("(%s)", strings.Join(placeholderRow, ", ")))
	}

	return insertQueryArgs{
		TableName:    quoter.Quote(model.TableName()),
		ColumnsDecl:  strings.Join(quotedColumns, ", "),
		Columns:      columns,
		Placeholders: strings.Join(placeholders, ",\n"),
	}
}

func buildInsertQueryValues[T any](mapper *reflectx.Mapper, columns []string, models []*T) (values []any, err error) {
	now := time.Now().UTC().Truncate(time.Microsecond)

	for _, m := range models {
		m := reflect.ValueOf(m)

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

	conn := p.Connection
	quoter, ok := conn.Dialect.(quoter)
	if !ok {
		return errors.Errorf("store is not a quoter: %T", conn.Store)
	}

	queryArgs := buildInsertQueryArgs(ctx, quoter, models)
	values, err := buildInsertQueryValues(conn.TX.Mapper, queryArgs.Columns, models)
	if err != nil {
		return err
	}

	query := conn.Dialect.TranslateSQL(fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES\n%s",
		queryArgs.TableName,
		queryArgs.ColumnsDecl,
		queryArgs.Placeholders,
	))

	_, err = conn.Store.ExecContext(ctx, query, values...)

	return sqlcon.HandleError(err)
}
