// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package update

import (
	"context"
	"fmt"

	"github.com/gobuffalo/pop/v6"
	"github.com/gobuffalo/pop/v6/columns"
	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/trace"

	"github.com/ory/x/sqlcon"
)

type Model interface {
	GetID() uuid.UUID
	GetNID() uuid.UUID
}

func Generic(ctx context.Context, c *pop.Connection, tracer trace.Tracer, v Model, columnNames ...string) error {
	ctx, span := tracer.Start(ctx, "persistence.sql.update")
	defer span.End()

	quoter, ok := c.Dialect.(interface{ Quote(key string) string })
	if !ok {
		return errors.Errorf("store is not a quoter: %T", c.Store)
	}

	model := pop.NewModel(v, ctx)
	tn := model.TableName()

	cols := columns.Columns{}
	if len(columnNames) > 0 && tn == model.TableName() {
		cols = columns.NewColumnsWithAlias(tn, model.As, model.IDField())
		cols.Add(columnNames...)
	} else {
		cols = columns.ForStructWithAlias(v, tn, model.As, model.IDField())
	}

	//#nosec G201 -- TableName is static
	stmt := fmt.Sprintf("SELECT COUNT(id) FROM %s AS %s WHERE %s.id = ? AND %s.nid = ?",
		quoter.Quote(model.TableName()),
		model.Alias(),
		model.Alias(),
		model.Alias(),
	)

	var count int
	if err := c.Store.GetContext(ctx, &count, c.Dialect.TranslateSQL(stmt), v.GetID(), v.GetNID()); err != nil {
		return sqlcon.HandleError(err)
	} else if count == 0 {
		return errors.WithStack(sqlcon.ErrNoRows)
	}

	//#nosec G201 -- TableName is static
	stmt = fmt.Sprintf("UPDATE %s AS %s SET %s WHERE %s AND %s.nid = :nid",
		quoter.Quote(model.TableName()),
		model.Alias(),
		cols.Writeable().QuotedUpdateString(quoter),
		model.WhereNamedID(),
		model.Alias(),
	)

	if _, err := c.Store.NamedExecContext(ctx, stmt, v); err != nil {
		return sqlcon.HandleError(err)
	}
	return nil
}
