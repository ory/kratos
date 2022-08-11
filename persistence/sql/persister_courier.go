package sql

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
	"github.com/pkg/errors"

	"github.com/ory/x/sqlcon"

	"github.com/ory/kratos/courier"
)

var _ courier.Persister = new(Persister)

func (p *Persister) AddMessage(ctx context.Context, m *courier.Message) error {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.AddMessage")
	defer span.End()

	m.NID = p.NetworkID(ctx)
	m.Status = courier.MessageStatusQueued
	return sqlcon.HandleError(p.GetConnection(ctx).Create(m)) // do not create eager to avoid identity injection.
}

func (p *Persister) NextMessages(ctx context.Context, limit uint8) (messages []courier.Message, err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.NextMessages")
	defer span.End()

	if err := p.Transaction(ctx, func(ctx context.Context, tx *pop.Connection) error {
		var m []courier.Message
		if err := tx.
			Where("nid = ? AND status = ?",
				p.NetworkID(ctx),
				courier.MessageStatusQueued,
			).
			Order("created_at ASC").
			Limit(int(limit)).
			All(&m); err != nil {
			return err
		}

		if len(m) == 0 {
			return sql.ErrNoRows
		}

		for i := range m {
			message := &m[i]
			message.Status = courier.MessageStatusProcessing
			if err := p.update(ctx, message, "status"); err != nil {
				return err
			}
		}

		messages = m
		return nil
	}); err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, errors.WithStack(courier.ErrQueueEmpty)
		}
		return nil, sqlcon.HandleError(err)
	}

	if len(messages) == 0 {
		return nil, errors.WithStack(courier.ErrQueueEmpty)
	}

	return messages, nil
}

func (p *Persister) LatestQueuedMessage(ctx context.Context) (*courier.Message, error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.LatestQueuedMessage")
	defer span.End()

	var m courier.Message
	if err := p.GetConnection(ctx).
		Where("nid = ? AND status = ?",
			p.NetworkID(ctx),
			courier.MessageStatusQueued,
		).
		Order("created_at DESC").
		First(&m); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.WithStack(courier.ErrQueueEmpty)
		}
		return nil, sqlcon.HandleError(err)
	}

	return &m, nil
}

func (p *Persister) SetMessageStatus(ctx context.Context, id uuid.UUID, ms courier.MessageStatus) error {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.SetMessageStatus")
	defer span.End()

	count, err := p.GetConnection(ctx).RawQuery(
		// #nosec G201
		fmt.Sprintf(
			"UPDATE %s SET status = ? WHERE id = ? AND nid = ?",
			"courier_messages",
		),
		ms,
		id,
		p.NetworkID(ctx),
	).ExecWithCount()
	if err != nil {
		return sqlcon.HandleError(err)
	}

	if count == 0 {
		return errors.WithStack(sqlcon.ErrNoRows)
	}

	return nil
}

func (p *Persister) IncrementMessageSendCount(ctx context.Context, id uuid.UUID) error {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.SetMessageStatus")
	defer span.End()

	count, err := p.GetConnection(ctx).RawQuery(
		// #nosec G201
		fmt.Sprintf(
			"UPDATE %s SET send_count = send_count + 1 WHERE id = ? AND nid = ?",
			"courier_messages",
		),
		id,
		p.NetworkID(ctx),
	).ExecWithCount()

	if err != nil {
		return sqlcon.HandleError(err)
	}

	if count == 0 {
		return errors.WithStack(sqlcon.ErrNoRows)
	}

	return nil
}
