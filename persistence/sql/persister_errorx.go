// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package sql

import (
	"bytes"
	"context"
	"encoding/json"
	"time"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"

	"github.com/ory/jsonschema/v3"

	"github.com/ory/herodot"
	"github.com/ory/x/otelx"
	"github.com/ory/x/sqlcon"

	"github.com/ory/kratos/selfservice/errorx"
)

var _ errorx.Persister = new(Persister)

func (p *Persister) CreateErrorContainer(ctx context.Context, csrfToken string, errs error) (containerID uuid.UUID, err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.CreateErrorContainer")
	defer otelx.End(span, &err)

	message, err := p.encodeSelfServiceErrors(ctx, errs)
	if err != nil {
		return uuid.Nil, err
	}

	c := &errorx.ErrorContainer{
		ID:        uuid.Nil,
		NID:       p.NetworkID(ctx),
		CSRFToken: csrfToken,
		Errors:    message,
		WasSeen:   false,
	}

	if err := p.GetConnection(ctx).Create(c); err != nil {
		return uuid.Nil, sqlcon.HandleError(err)
	}

	return c.ID, nil
}

func (p *Persister) ReadErrorContainer(ctx context.Context, id uuid.UUID) (_ *errorx.ErrorContainer, err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.ReadErrorContainer")
	defer otelx.End(span, &err)

	var ec errorx.ErrorContainer
	if err := p.GetConnection(ctx).Where("id = ? AND nid = ?", id, p.NetworkID(ctx)).First(&ec); err != nil {
		return nil, sqlcon.HandleError(err)
	}

	if err := p.GetConnection(ctx).RawQuery(
		"UPDATE selfservice_errors SET was_seen = true, seen_at = ? WHERE id = ? AND nid = ?",
		time.Now().UTC(), id, p.NetworkID(ctx)).Exec(); err != nil {
		return nil, sqlcon.HandleError(err)
	}

	return &ec, nil
}

func (p *Persister) ClearErrorContainers(ctx context.Context, olderThan time.Duration, force bool) (err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.ClearErrorContainers")
	defer otelx.End(span, &err)

	if force {
		err = p.GetConnection(ctx).RawQuery(
			"DELETE FROM selfservice_errors WHERE nid = ? AND seen_at < ? AND seen_at IS NOT NULL",
			p.NetworkID(ctx), time.Now().UTC().Add(-olderThan)).Exec()
	} else {
		err = p.GetConnection(ctx).RawQuery(
			"DELETE FROM selfservice_errors WHERE nid = ? AND was_seen=true AND seen_at < ? AND seen_at IS NOT NULL",
			p.NetworkID(ctx), time.Now().UTC().Add(-olderThan)).Exec()
	}

	return sqlcon.HandleError(err)
}

func (p *Persister) encodeSelfServiceErrors(ctx context.Context, e error) ([]byte, error) {
	if e == nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithDebug("A nil error was passed to the error manager which is most likely a code bug."))
	}

	if c := new(herodot.DefaultError); errors.As(e, &c) {
		e = c
	} else if c := new(jsonschema.ValidationError); errors.As(e, &c) {
		e = c
	} else {
		e = herodot.ToDefaultError(e, "")
	}

	var b bytes.Buffer
	if err := json.NewEncoder(&b).Encode(e); err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReason("Unable to encode error messages.").WithDebug(err.Error()))
	}

	return b.Bytes(), nil
}
