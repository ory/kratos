// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package sql

import (
	"context"
	"fmt"
	"time"

	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
	"github.com/pkg/errors"

	"github.com/ory/kratos/selfservice/sessiontokenexchange"
	"github.com/ory/x/otelx"
	"github.com/ory/x/randx"
	"github.com/ory/x/sqlcon"
)

var _ sessiontokenexchange.Persister = new(Persister)

func updateLimitClause(conn *pop.Connection) string {
	// Not all databases support limiting in update clauses.
	switch conn.Dialect.Name() {
	case "sqlite3", "postgres":
		return ""
	default:
		return "LIMIT 1"
	}
}

func (p *Persister) CreateSessionTokenExchanger(ctx context.Context, flowID uuid.UUID) (e *sessiontokenexchange.Exchanger, err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.CreateSessionTokenExchanger")
	defer otelx.End(span, &err)

	e = &sessiontokenexchange.Exchanger{
		NID:          p.NetworkID(ctx),
		FlowID:       flowID,
		InitCode:     randx.MustString(64, randx.AlphaNum),
		ReturnToCode: randx.MustString(64, randx.AlphaNum),
	}

	return e, p.GetConnection(ctx).Create(e)
}

func (p *Persister) GetExchangerFromCode(ctx context.Context, initCode string, returnToCode string) (e *sessiontokenexchange.Exchanger, err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.GetExchangerFromCode")
	defer otelx.End(span, &err)

	e = new(sessiontokenexchange.Exchanger)
	conn := p.GetConnection(ctx)
	if err = conn.Where(`
nid = ? AND
init_code = ? AND init_code <> '' AND
return_to_code = ? AND return_to_code <> '' AND
session_id IS NOT NULL`,
		p.NetworkID(ctx), initCode, returnToCode).First(e); err != nil {
		return nil, sqlcon.HandleError(err)
	}

	return e, nil
}

func (p *Persister) UpdateSessionOnExchanger(ctx context.Context, flowID uuid.UUID, sessionID uuid.UUID) (err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.UpdateSessionOnExchanger")
	defer otelx.End(span, &err)

	conn := p.GetConnection(ctx)
	query := fmt.Sprintf("UPDATE %s SET session_id = ? WHERE flow_id = ? AND nid = ? %s",
		conn.Dialect.Quote(new(sessiontokenexchange.Exchanger).TableName()),
		updateLimitClause(conn),
	)

	return sqlcon.HandleError(conn.RawQuery(query, sessionID, flowID, p.NetworkID(ctx)).Exec())
}

func (p *Persister) CodeForFlow(ctx context.Context, flowID uuid.UUID) (codes *sessiontokenexchange.Codes, found bool, err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.CodeForFlow")
	defer otelx.End(span, &err)

	var e sessiontokenexchange.Exchanger
	switch err = sqlcon.HandleError(p.GetConnection(ctx).
		Where("flow_id = ? AND nid = ? AND init_code <> '' and return_to_code <> ''", flowID, p.NetworkID(ctx)).
		First(&e)); {
	case err == nil:
		return &sessiontokenexchange.Codes{
			InitCode:     e.InitCode,
			ReturnToCode: e.ReturnToCode,
		}, true, nil
	case errors.Is(err, sqlcon.ErrNoRows):
		return nil, false, nil
	default:
		return nil, false, err
	}
}

func (p *Persister) MoveToNewFlow(ctx context.Context, oldFlow, newFlow uuid.UUID) (err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.MoveToNewFlow")
	defer otelx.End(span, &err)

	conn := p.GetConnection(ctx)
	query := fmt.Sprintf("UPDATE %s SET flow_id = ? WHERE flow_id = ? AND nid = ? %s",
		conn.Dialect.Quote(new(sessiontokenexchange.Exchanger).TableName()),
		updateLimitClause(conn),
	)

	return sqlcon.HandleError(conn.RawQuery(query, newFlow, oldFlow, p.NetworkID(ctx)).Exec())
}

func (p *Persister) DeleteExpiredExchangers(ctx context.Context, at time.Time, limit int) error {
	expiredAfter := at.Add(1 * time.Hour)
	conn := p.GetConnection(ctx)

	//#nosec G201 -- TableName is static
	err := conn.RawQuery(fmt.Sprintf(
		"DELETE FROM %s WHERE id in (SELECT id FROM (SELECT id FROM %s c WHERE created_at <= ? and nid = ? ORDER BY created_at ASC LIMIT %d ) AS s )",
		conn.Dialect.Quote(new(sessiontokenexchange.Exchanger).TableName()),
		conn.Dialect.Quote(new(sessiontokenexchange.Exchanger).TableName()),
		limit,
	),
		expiredAfter,
		p.NetworkID(ctx),
	).Exec()

	return sqlcon.HandleError(err)
}
