// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package sql

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ory/kratos/x"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"

	"github.com/ory/jsonschema/v3"

	"github.com/ory/herodot"
	"github.com/ory/x/sqlcon"

	"github.com/ory/kratos/selfservice/errorx"
)

var _ errorx.Persister = new(Persister)

func (p *Persister) Add(ctx context.Context, csrfToken string, errs error) (uuid.UUID, error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.Add")
	defer span.End()

	buf, err := p.encodeSelfServiceErrors(ctx, errs)
	if err != nil {
		return uuid.Nil, err
	}

	c := &errorx.ErrorContainer{
		ID:        x.NewUUID(),
		NID:       p.NetworkID(ctx),
		CSRFToken: csrfToken,
		Errors:    buf.Bytes(),
		WasSeen:   false,
	}

	if err := p.GetConnection(ctx).Create(c); err != nil {
		return uuid.Nil, sqlcon.HandleError(err)
	}

	return c.ID, nil
}

func (p *Persister) Read(ctx context.Context, id uuid.UUID) (*errorx.ErrorContainer, error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.Read")
	defer span.End()

	var ec errorx.ErrorContainer
	if err := p.GetConnection(ctx).Where("id = ? AND nid = ?", id, p.NetworkID(ctx)).First(&ec); err != nil {
		return nil, sqlcon.HandleError(err)
	}

	// #nosec G201
	if err := p.GetConnection(ctx).RawQuery(
		fmt.Sprintf("UPDATE %s SET was_seen = true, seen_at = ? WHERE id = ? AND nid = ?", "selfservice_errors"),
		time.Now().UTC(), id, p.NetworkID(ctx)).Exec(); err != nil {
		return nil, sqlcon.HandleError(err)
	}

	return &ec, nil
}

func (p *Persister) Clear(ctx context.Context, olderThan time.Duration, force bool) (err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.Clear")
	defer span.End()

	if force {
		// #nosec G201
		err = p.GetConnection(ctx).RawQuery(
			fmt.Sprintf("DELETE FROM %s WHERE nid = ? AND seen_at < ? AND seen_at IS NOT NULL", "selfservice_errors"),
			p.NetworkID(ctx), time.Now().UTC().Add(-olderThan)).Exec()
	} else {
		// #nosec G201
		err = p.GetConnection(ctx).RawQuery(
			fmt.Sprintf("DELETE FROM %s WHERE nid = ? AND was_seen=true AND seen_at < ? AND seen_at IS NOT NULL", "selfservice_errors"),
			p.NetworkID(ctx), time.Now().UTC().Add(-olderThan)).Exec()
	}

	return sqlcon.HandleError(err)
}

func (p *Persister) encodeSelfServiceErrors(ctx context.Context, e error) (*bytes.Buffer, error) {
	_, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.encodeSelfServiceErrors")
	defer span.End()

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

	return &b, nil
}
