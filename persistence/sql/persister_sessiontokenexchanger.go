// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package sql

import (
	"context"
	"fmt"

	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
	"github.com/pkg/errors"

	"github.com/ory/kratos/selfservice/sessiontokenexchange"
	"github.com/ory/x/otelx"
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

func (p *Persister) CreateSessionTokenExchanger(ctx context.Context, flowID uuid.UUID, code string) (err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.CreateSessionTokenExchanger")
	defer otelx.End(span, &err)

	e := sessiontokenexchange.Exchanger{
		NID:    p.NetworkID(ctx),
		FlowID: flowID,
		Code:   code,
	}

	return p.GetConnection(ctx).Create(&e)
}

func (p *Persister) GetExchangerFromCode(ctx context.Context, code string) (e *sessiontokenexchange.Exchanger, err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.GetExchangerFromCode")
	defer otelx.End(span, &err)

	e = new(sessiontokenexchange.Exchanger)
	conn := p.GetConnection(ctx)
	if err = conn.Where(
		"nid = ? AND code = ? AND session_id IS NOT NULL AND code <> ''",
		p.NetworkID(ctx), code).First(e); err != nil {
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

func (p *Persister) CodeForFlow(ctx context.Context, flowID uuid.UUID) (code string, found bool, err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.CodeForFlow")
	defer otelx.End(span, &err)

	var e sessiontokenexchange.Exchanger
	switch err = sqlcon.HandleError(p.GetConnection(ctx).
		Where("flow_id = ? AND nid = ? AND code <> ''", flowID, p.NetworkID(ctx)).
		First(&e)); {
	case err == nil:
		return e.Code, true, nil
	case errors.Is(err, sqlcon.ErrNoRows):
		return "", false, nil
	default:
		return "", false, err
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
