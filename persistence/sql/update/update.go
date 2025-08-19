// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package update

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/trace"

	"github.com/ory/pop/v6"
	"github.com/ory/pop/v6/columns"
	"github.com/ory/x/otelx"
	"github.com/ory/x/sqlcon"
)

func Generic(ctx context.Context, c *pop.Connection, tracer trace.Tracer, v any, columnNames ...string) (err error) {
	ctx, span := tracer.Start(ctx, "persistence.sql.update")
	defer otelx.End(span, &err)

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
	stmt := fmt.Sprintf("UPDATE %s AS %s SET %s WHERE %s AND %s.nid = :nid",
		quoter.Quote(model.TableName()),
		model.Alias(),
		cols.Writeable().QuotedUpdateString(quoter),
		model.WhereNamedID(),
		model.Alias(),
	)

	if result, err := c.Store.NamedExecContext(ctx, stmt, v); err != nil {
		return sqlcon.HandleError(err)
	} else if affected, err := result.RowsAffected(); err != nil {
		return sqlcon.HandleError(err)
	} else if affected == 0 {
		return errors.WithStack(sqlcon.ErrNoRows)
	}

	return nil
}
