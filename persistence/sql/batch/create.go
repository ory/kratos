// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package batch

import (
	"context"
	"database/sql"
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

	// PartialConflictError represents a partial conflict during [Create]. It always
	// wraps a [sqlcon.ErrUniqueViolation], so that the caller can either abort the
	// whole transaction, or handle the partial success.
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
func (p *PartialConflictError[T]) Unwrap() error {
	if len(p.Failed) == 0 {
		return nil
	}
	return sqlcon.ErrUniqueViolation
}

func buildInsertQueryArgs[T any](ctx context.Context, models []*T, opts *createOpts) insertQueryArgs {
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
		quotedColumns = append(quotedColumns, opts.quoter.Quote(col))
	}

	// We generate a list (for every row one) of VALUE statements here that
	// will be substituted by their column values later:
	//
	//	(?, ?, ?, ?),
	//	(?, ?, ?, ?),
	//	(?, ?, ?, ?)
	for _, m := range models {
		m := reflect.ValueOf(m)

		pl := make([]string, len(placeholderRow))
		copy(pl, placeholderRow)

		// There is a special case - when using CockroachDB we want to generate
		// UUIDs using "gen_random_uuid()" which ends up in a VALUE statement of:
		//
		//	(gen_random_uuid(), ?, ?, ?),
		for k := range placeholderRow {
			if columns[k] != "id" {
				continue
			}

			field := opts.mapper.FieldByName(m, columns[k])
			val, ok := field.Interface().(uuid.UUID)
			if !ok {
				continue
			}

			if val == uuid.Nil && opts.dialect == dbal.DriverCockroachDB && !opts.partialInserts {
				pl[k] = "gen_random_uuid()"
				break
			}
		}
		placeholders = append(placeholders, fmt.Sprintf("(%s)", strings.Join(pl, ", ")))
	}

	return insertQueryArgs{
		TableName:    opts.quoter.Quote(model.TableName()),
		ColumnsDecl:  strings.Join(quotedColumns, ", "),
		Columns:      columns,
		Placeholders: strings.Join(placeholders, ",\n"),
	}
}

func buildInsertQueryValues[T any](columns []string, models []*T, opts *createOpts) (values []any, err error) {
	for _, m := range models {
		m := reflect.ValueOf(m)

		now := opts.now()
		// Append model fields to args
		for _, c := range columns {
			field := opts.mapper.FieldByName(m, c)

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
				} else if opts.dialect == dbal.DriverCockroachDB && !opts.partialInserts {
					// This is a special case:
					// 1. We're using cockroach
					// 2. It's the primary key field ("ID")
					// 3. A UUID was not yet set.
					//
					// If all these conditions meet, the VALUE statement will look as such:
					//
					//	(gen_random_uuid(), ?, ?, ?, ...)
					//
					// For that reason, we do not add the ID value to the list of arguments,
					// because one of the arguments is using a built-in and thus doesn't need a value.
					continue // break switch, not for
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

type createOpts struct {
	partialInserts bool
	dialect        string
	mapper         *reflectx.Mapper
	quoter         quoter
	now            func() time.Time
}

type CreateOpts func(*createOpts)

// WithPartialInserts allows to insert only the models that do not conflict with
// an existing record. WithPartialInserts will also generate the IDs for the
// models before inserting them, so that the successful inserts can be correlated
// with the input models.
//
// In particular, WithPartialInserts does not work with MySQL, because it does
// not support the "RETURNING" clause.
//
// WithPartialInserts does not work with CockroachDB and gen_random_uuid(),
// because then the successful inserts cannot be correlated with the input
// models. Note: gen_random_uuid() will skip the UNIQUE constraint check, which
// needs to hit all regions in a distributed setup. Therefore, WithPartialInserts
// should not be used to insert models for only a single identity.
var WithPartialInserts CreateOpts = func(o *createOpts) {
	o.partialInserts = true
}

func newCreateOpts(conn *pop.Connection, opts ...CreateOpts) *createOpts {
	o := new(createOpts)
	o.dialect = conn.Dialect.Name()
	o.mapper = conn.TX.Mapper
	o.quoter = conn.Dialect.(quoter)
	o.now = func() time.Time { return time.Now().UTC().Truncate(time.Microsecond) }
	for _, f := range opts {
		f(o)
	}
	return o
}

// Create batch-inserts the given models into the database using a single INSERT
// statement. By default, the models are either all created or none. If
// [WithPartialInserts] is passed as an option, partial inserts are supported,
// and the models that could not be inserted are returned in an
// [PartialConflictError].
func Create[T any](ctx context.Context, p *TracerConnection, models []*T, opts ...CreateOpts) (err error) {
	ctx, span := p.Tracer.Tracer().Start(ctx, "persistence.sql.batch.Create",
		trace.WithAttributes(attribute.Int("count", len(models))))
	defer otelx.End(span, &err)

	if len(models) == 0 {
		return nil
	}

	var v T
	model := pop.NewModel(v, ctx)

	conn := p.Connection
	options := newCreateOpts(conn, opts...)

	queryArgs := buildInsertQueryArgs(ctx, models, options)
	values, err := buildInsertQueryValues(queryArgs.Columns, models, options)
	if err != nil {
		return err
	}

	var returningClause string
	if conn.Dialect.Name() != dbal.DriverMySQL {
		// PostgreSQL, CockroachDB, SQLite support RETURNING.
		if options.partialInserts {
			returningClause = fmt.Sprintf("ON CONFLICT DO NOTHING RETURNING %s", model.IDField())
		} else {
			returningClause = fmt.Sprintf("RETURNING %s", model.IDField())
		}
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

	// MySQL, which does not support RETURNING, also does not have ON CONFLICT DO
	// NOTHING, meaning that MySQL will always fail the whole transaction on a single
	// record conflict.
	if conn.Dialect.Name() == dbal.DriverMySQL {
		return nil
	}

	if options.partialInserts {
		return handlePartialInserts(queryArgs, values, models, rows)
	} else {
		return handleFullInserts(models, rows)
	}

}

func handleFullInserts[T any](models []*T, rows *sql.Rows) error {
	// Hydrate the models from the RETURNING clause.
	for i := 0; rows.Next(); i++ {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return errors.WithStack(err)
		}
		if err := setModelID(id, models[i]); err != nil {
			return err
		}
	}
	if err := rows.Err(); err != nil {
		return sqlcon.HandleError(err)
	}

	return nil
}

func handlePartialInserts[T any](queryArgs insertQueryArgs, values []any, models []*T, rows *sql.Rows) error {
	// Hydrate the models from the RETURNING clause.
	idsInDB := make(map[uuid.UUID]struct{})
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return errors.WithStack(err)
		}
		idsInDB[id] = struct{}{}
	}
	if err := rows.Err(); err != nil {
		return sqlcon.HandleError(err)
	}

	idIdx := slices.Index(queryArgs.Columns, "id")
	if idIdx == -1 {
		return errors.New("id column not found")
	}
	var idValues []uuid.UUID
	for i := idIdx; i < len(values); i += len(queryArgs.Columns) {
		idValues = append(idValues, values[i].(uuid.UUID))
	}

	var partialConflictError PartialConflictError[T]
	for i, id := range idValues {
		if _, ok := idsInDB[id]; !ok {
			partialConflictError.Failed = append(partialConflictError.Failed, models[i])
		} else {
			if err := setModelID(id, models[i]); err != nil {
				return err
			}
		}
	}

	return partialConflictError.ErrOrNil()
}

// setModelID sets the id field of the model to the id.
func setModelID(id uuid.UUID, model any) error {
	el := reflect.ValueOf(model).Elem()
	idField := el.FieldByName("ID")
	if !idField.IsValid() {
		return errors.New("model does not have a field named id")
	}
	idField.Set(reflect.ValueOf(id))

	return nil
}
