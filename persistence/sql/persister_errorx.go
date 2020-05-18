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
	"github.com/ory/x/sqlcon"

	"github.com/ory/kratos/selfservice/errorx"
)

var _ errorx.Persister = new(Persister)

func (p *Persister) Add(ctx context.Context, csrfToken string, errs ...error) (uuid.UUID, error) {
	buf, err := p.encodeSelfServiceErrors(errs)
	if err != nil {
		return uuid.Nil, err
	}

	c := &errorx.ErrorContainer{
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
	var ec errorx.ErrorContainer
	if err := p.GetConnection(ctx).Find(&ec, id); err != nil {
		return nil, sqlcon.HandleError(err)
	}

	if err := p.GetConnection(ctx).RawQuery("UPDATE selfservice_errors SET was_seen = true, seen_at = ? WHERE id = ?", time.Now().UTC(), id).Exec(); err != nil {
		return nil, sqlcon.HandleError(err)
	}

	return &ec, nil
}

func (p *Persister) Clear(ctx context.Context, olderThan time.Duration, force bool) (err error) {
	if force {
		err = p.GetConnection(ctx).RawQuery("DELETE FROM selfservice_errors WHERE seen_at < ? AND seen_at IS NOT NULL", olderThan).Exec()
	} else {
		err = p.GetConnection(ctx).RawQuery("DELETE FROM selfservice_errors WHERE was_seen=true AND seen_at < ? AND seen_at IS NOT NULL", time.Now().UTC().Add(-olderThan)).Exec()
	}

	return sqlcon.HandleError(err)
}

func (p *Persister) encodeSelfServiceErrors(errs []error) (*bytes.Buffer, error) {
	es := make([]interface{}, len(errs))
	for k, e := range errs {
		if e == nil {
			return nil, errors.WithStack(herodot.ErrInternalServerError.WithDebug("A nil error was passed to the error manager which is most likely a code bug."))
		}

		if c := new(herodot.DefaultError); errors.As(e, &c) {
			es[k] = c
		} else if c := new(jsonschema.ValidationError); errors.As(e, &c) {
			es[k] = c
		} else {
			es[k] = herodot.ToDefaultError(e, "")
		}
	}

	var b bytes.Buffer
	if err := json.NewEncoder(&b).Encode(es); err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReason("Unable to encode error messages.").WithDebug(err.Error()))
	}

	return &b, nil
}
