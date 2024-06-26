// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package continuity

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"

	"github.com/ory/herodot"
	"github.com/ory/x/otelx"
	"github.com/ory/x/sqlcon"

	"github.com/ory/kratos/session"
	"github.com/ory/kratos/x"
)

var (
	_               Manager = new(ManagerCookie)
	ErrNotResumable         = *herodot.ErrBadRequest.WithError("no resumable session found").WithReasonf("The browser does not contain the necessary cookie to resume the session. This is a security violation and was blocked. Please clear your browser's cookies and cache and try again!")
)

const CookieName = "ory_kratos_continuity"

type (
	managerCookieDependencies interface {
		PersistenceProvider
		x.CookieProvider
		session.ManagementProvider
		x.TracingProvider
	}
	ManagerCookie struct {
		d managerCookieDependencies
	}
)

func NewManagerCookie(d managerCookieDependencies) *ManagerCookie {
	return &ManagerCookie{d: d}
}

func (m *ManagerCookie) Pause(ctx context.Context, w http.ResponseWriter, r *http.Request, name string, opts ...ManagerOption) (err error) {
	ctx, span := m.d.Tracer(ctx).Tracer().Start(ctx, "continuity.ManagerCookie.Pause")
	defer otelx.End(span, &err)
	if len(name) == 0 {
		return errors.Errorf("continuity container name must be set")
	}

	o, err := newManagerOptions(opts)
	if err != nil {
		return err
	}
	c := NewContainer(name, *o)

	if err := m.d.ContinuityPersister().SaveContinuitySession(ctx, c); err != nil {
		return errors.WithStack(err)
	}

	if err := x.SessionPersistValues(w, r, m.d.ContinuityCookieManager(ctx), CookieName, map[string]interface{}{
		name: c.ID.String(),
	}); err != nil {
		return err
	}

	return nil
}

func (m *ManagerCookie) Continue(ctx context.Context, w http.ResponseWriter, r *http.Request, name string, opts ...ManagerOption) (container *Container, err error) {
	ctx, span := m.d.Tracer(ctx).Tracer().Start(ctx, "continuity.ManagerCookie.Continue")
	defer otelx.End(span, &err)

	container, err = m.container(ctx, w, r, name)
	if err != nil {
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

	if o.setExpiresIn > 0 {
		if err := m.d.ContinuityPersister().SetContinuitySessionExpiry(
			ctx,
			container.ID,
			time.Now().UTC().Add(o.setExpiresIn).Truncate(time.Second),
		); err != nil && !errors.Is(err, sqlcon.ErrNoRows) {
			return nil, err
		}
	} else {
		if err := x.SessionUnsetKey(w, r, m.d.ContinuityCookieManager(ctx), CookieName, name); err != nil {
			return nil, err
		}

		if err := m.d.ContinuityPersister().DeleteContinuitySession(ctx, container.ID); err != nil && !errors.Is(err, sqlcon.ErrNoRows) {
			return nil, err
		}
	}

	return container, nil
}

func (m *ManagerCookie) sessionID(ctx context.Context, w http.ResponseWriter, r *http.Request, name string) (uuid.UUID, error) {
	s, err := x.SessionGetString(r, m.d.ContinuityCookieManager(ctx), CookieName, name)
	if err != nil {
		_ = x.SessionUnsetKey(w, r, m.d.ContinuityCookieManager(ctx), CookieName, name)
		return uuid.Nil, errors.WithStack(ErrNotResumable.WithDebugf("%+v", err))
	}

	sid, err := uuid.FromString(s)
	if err != nil {
		_ = x.SessionUnsetKey(w, r, m.d.ContinuityCookieManager(ctx), CookieName, name)
		return uuid.Nil, errors.WithStack(ErrNotResumable.WithDebug("session id is not a valid uuid"))
	}

	return sid, nil
}

func (m *ManagerCookie) container(ctx context.Context, w http.ResponseWriter, r *http.Request, name string) (*Container, error) {
	sid, err := m.sessionID(ctx, w, r, name)
	if err != nil {
		return nil, err
	}

	container, err := m.d.ContinuityPersister().GetContinuitySession(ctx, sid)
	// If an error happens, we need to clean up the cookie.
	if err != nil {
		_ = x.SessionUnsetKey(w, r, m.d.ContinuityCookieManager(ctx), CookieName, name)
	}

	if errors.Is(err, sqlcon.ErrNoRows) {
		return nil, errors.WithStack(ErrNotResumable.WithDebugf("Resumable ID from cookie could not be found in the datastore: %+v", err))
	} else if err != nil {
		return nil, err
	} else if container.ExpiresAt.Before(time.Now()) {
		_ = x.SessionUnsetKey(w, r, m.d.ContinuityCookieManager(ctx), CookieName, name)
		return nil, errors.WithStack(ErrNotResumable.WithDebugf("Resumable session has expired"))
	}

	return container, err
}

func (m ManagerCookie) Abort(ctx context.Context, w http.ResponseWriter, r *http.Request, name string) (err error) {
	ctx, span := m.d.Tracer(ctx).Tracer().Start(ctx, "continuity.ManagerCookie.Abort")
	defer otelx.End(span, &err)
	sid, err := m.sessionID(ctx, w, r, name)
	if errors.Is(err, &ErrNotResumable) {
		// We do not care about an error here
		return nil
	} else if err != nil {
		return err
	}

	if err := x.SessionUnsetKey(w, r, m.d.ContinuityCookieManager(ctx), CookieName, name); err != nil {
		return err
	}

	if err := m.d.ContinuityPersister().DeleteContinuitySession(ctx, sid); err != nil && !errors.Is(err, sqlcon.ErrNoRows) {
		return errors.WithStack(err)
	}

	return nil
}
