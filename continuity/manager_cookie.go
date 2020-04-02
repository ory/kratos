package continuity

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"

	"github.com/ory/herodot"
	"github.com/ory/x/sqlcon"

	"github.com/ory/kratos/session"
	"github.com/ory/kratos/x"
)

var _ Manager = new(ManagerCookie)

const sessionKeyID = "id"

type (
	managerCookieDependencies interface {
		PersistenceProvider
		x.CookieProvider
		session.ManagementProvider
	}
	ManagerCookie struct {
		d managerCookieDependencies
	}
)

func NewManagerCookie(d managerCookieDependencies) *ManagerCookie {
	return &ManagerCookie{d: d}
}

func (m ManagerCookie) Pause(ctx context.Context, w http.ResponseWriter, r *http.Request, name string, opts ...ManagerOption) error {
	if len(name) == 0 {
		return errors.Errorf("continuity container name must be set")
	}

	o, err := newManagerOptions(opts)
	if err != nil {
		return err
	}
	c := NewContainer(name, *o)

	if err := x.SessionPersistValues(w, r, m.d.CookieManager(), name, map[string]interface{}{
		sessionKeyID: c.ID.String(),
	}); err != nil {
		return err
	}

	if err := m.d.ContinuityPersister().SaveContinuitySession(ctx, c); err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (m ManagerCookie) Continue(ctx context.Context, r *http.Request, name string, opts ...ManagerOption) (*Container, error) {
	var sid uuid.UUID

	if s, err := x.SessionGetString(r, m.d.CookieManager(), name, sessionKeyID); err != nil {
		return nil, errors.WithStack(herodot.ErrBadRequest.WithReasonf("No resumable session could be found in the HTTP Header.").WithDebugf("%+v", err))
	} else if sid = x.ParseUUID(s); sid == uuid.Nil {
		return nil, errors.WithStack(herodot.ErrBadRequest.WithReasonf("No resumable session could be found in the HTTP Header.").WithDebug("sid is not a valid uuid"))
	}

	container, err := m.d.ContinuityPersister().GetContinuitySession(ctx, sid)
	if errors.Is(err, sqlcon.ErrNoRows) {
		return nil, errors.WithStack(herodot.ErrBadRequest.WithReasonf("No resumable session could be found in the HTTP Header.").WithDebugf("Resumable ID from cookie could not be found in the datastore: %+v", err))
	} else if err != nil {
		return nil, err
	}

	o, err := newManagerOptions(opts)
	if err != nil {
		return nil, err
	}

	if err := container.Valid(o.iid); err != nil {
		return nil, err
	}

	if o.payloadRaw != nil && container.Payload != nil {
		if err := json.NewDecoder(bytes.NewBuffer(container.Payload)).Decode(o.payloadRaw); err != nil {
			return nil, errors.WithStack(err)
		}
	}

	if err := m.d.ContinuityPersister().DeleteContinuitySession(ctx, container.ID); err != nil {
		return nil, err
	}

	return container, nil
}
