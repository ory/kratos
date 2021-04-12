package sql

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/gobuffalo/pop/v5"
	"github.com/gofrs/uuid"
	"github.com/pkg/errors"

	"github.com/ory/kratos/corp"

	"github.com/ory/x/sqlcon"

	"github.com/ory/kratos/courier"
)

var _ courier.Persister = new(Persister)

func (p *Persister) AddMessage(ctx context.Context, m *courier.Message) error {
	m.NID = corp.ContextualizeNID(ctx, p.nid)
	m.Status = courier.MessageStatusQueued
	return sqlcon.HandleError(p.GetConnection(ctx).Create(m)) // do not create eager to avoid identity injection.
}

func (p *Persister) NextMessages(ctx context.Context, limit uint8) ([]courier.Message, error) {
	var m []courier.Message
	if err := p.Transaction(ctx, func(ctx context.Context, tx *pop.Connection) error {
		if err := tx.
			Where("nid = ? AND status = ?", corp.ContextualizeNID(ctx, p.nid), courier.MessageStatusQueued).
			Order("created_at ASC").Limit(int(limit)).All(&m); err != nil {
			return err
		}
		for i := range m {
			m[i].Status = courier.MessageStatusProcessing
		}
		return tx.UpdateColumns(&m, "status")
	}); err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, errors.WithStack(courier.ErrQueueEmpty)
		}
		return nil, sqlcon.HandleError(err)
	}

	if len(m) == 0 {
		return nil, errors.WithStack(courier.ErrQueueEmpty)
	}

	return m, nil
}

func (p *Persister) LatestQueuedMessage(ctx context.Context) (*courier.Message, error) {
	var m courier.Message
	if err := p.GetConnection(ctx).
		Where("nid = ? AND status = ?", corp.ContextualizeNID(ctx, p.nid), courier.MessageStatusQueued).
		Order("created_at DESC").First(&m); err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, errors.WithStack(courier.ErrQueueEmpty)
		}
		return nil, sqlcon.HandleError(err)
	}

	return &m, nil
}

func (p *Persister) SetMessageStatus(ctx context.Context, id uuid.UUID, ms courier.MessageStatus) error {
	count, err := p.GetConnection(ctx).RawQuery(
		// #nosec G201
		fmt.Sprintf(
			"UPDATE %s SET status = ? WHERE id = ? AND nid = ?",
			corp.ContextualizeTableName(ctx, "courier_messages"),
		), ms, id, corp.ContextualizeNID(ctx, p.nid)).ExecWithCount()
	if err != nil {
		return sqlcon.HandleError(err)
	}

	if count == 0 {
		return errors.WithStack(sqlcon.ErrNoRows)
	}

	return nil
}
