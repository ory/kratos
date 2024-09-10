// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package batch

import (
	"context"
	"fmt"
	"reflect"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
	"github.com/jmoiron/sqlx/reflectx"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/ory/x/dbal"
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

	PartialConflictError[T any] struct {
		Failed []*T
	}
)

func (p *PartialConflictError[T]) Error() string {
	return fmt.Sprintf("partial conflict error: %d models failed to insert", len(p.Failed))
}
func (p *PartialConflictError[T]) ErrOrNil() error {
	if len(p.Failed) == 0 {
		return nil
	}
	return p
}

func buildInsertQueryArgs[T any](ctx context.Context, dialect string, mapper *reflectx.Mapper, quoter quoter, models []*T) insertQueryArgs {
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

	// We generate a list (for every row one) of VALUE statements here that
	// will be substituted by their column values later:
	//
	//	(?, ?, ?, ?),
	//	(?, ?, ?, ?),
	//	(?, ?, ?, ?)
	for range models {
		pl := make([]string, len(placeholderRow))
		copy(pl, placeholderRow)

		placeholders = append(placeholders, fmt.Sprintf("(%s)", strings.Join(pl, ", ")))
	}

	return insertQueryArgs{
		TableName:    quoter.Quote(model.TableName()),
		ColumnsDecl:  strings.Join(quotedColumns, ", "),
		Columns:      columns,
		Placeholders: strings.Join(placeholders, ",\n"),
	}
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
	ctx, span := p.Tracer.Tracer().Start(ctx, "persistence.sql.batch.Create",
		trace.WithAttributes(attribute.Int("count", len(models))))
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

	queryArgs := buildInsertQueryArgs(ctx, conn.Dialect.Name(), conn.TX.Mapper, quoter, models)
	values, err := buildInsertQueryValues(conn.Dialect.Name(), conn.TX.Mapper, queryArgs.Columns, models, func() time.Time { return time.Now().UTC().Truncate(time.Microsecond) })
	if err != nil {
		return err
	}

	var returningClause string
	if conn.Dialect.Name() != dbal.DriverMySQL {
		// PostgreSQL, CockroachDB, SQLite support RETURNING.
		returningClause = fmt.Sprintf("ON CONFLICT DO NOTHING RETURNING %s", model.IDField())
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
	// MySQL, which does not support RETURNING, also does not have ON CONFLICT DO
	// NOTHING, meaning that MySQL will always fail the whole transaction on a single
	// record conflict.
	if conn.Dialect.Name() == dbal.DriverMySQL {
		return sqlcon.HandleError(rows.Close())
	}

	idIdx := slices.Index(queryArgs.Columns, "id")
	if idIdx == -1 {
		return errors.New("id column not found")
	}
	var idValues []uuid.UUID
	for i := idIdx; i < len(values); i += len(queryArgs.Columns) {
		idValues = append(idValues, values[i].(uuid.UUID))
	}

	// Hydrate the models from the RETURNING clause.
	idsInDB := make(map[uuid.UUID]struct{})
	for rows.Next() {
		if err := rows.Err(); err != nil {
			return sqlcon.HandleError(err)
		}
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return errors.WithStack(err)
		}
		idsInDB[id] = struct{}{}
	}
	if err := rows.Err(); err != nil {
		return sqlcon.HandleError(err)
	}

	if err := rows.Close(); err != nil {
		return sqlcon.HandleError(err)
	}

	var partialConflictError PartialConflictError[T]
	for i, id := range idValues {
		if _, ok := idsInDB[id]; !ok {
			partialConflictError.Failed = append(partialConflictError.Failed, models[i])
		} else {
			if err := setModelID(id, pop.NewModel(models[i], ctx)); err != nil {
				return err
			}
		}
	}

	if len(partialConflictError.Failed) > 0 {
		return sqlcon.ErrUniqueViolation.WithWrap(&partialConflictError)
	}

	return nil
}

// setModelID was copy & pasted from pop. It basically sets
// the primary key to the given value read from the SQL row.
func setModelID(id uuid.UUID, model *pop.Model) error {
	el := reflect.ValueOf(model.Value).Elem()
	fbn := el.FieldByName("ID")
	if !fbn.IsValid() {
		return errors.New("model does not have a field named id")
	}
	fbn.Set(reflect.ValueOf(id))

	return nil
}
