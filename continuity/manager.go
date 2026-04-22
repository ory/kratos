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
	"github.com/ory/kratos/identity"
	"github.com/ory/x/otelx"
	"github.com/ory/x/sqlcon"
)

func ErrNotResumable() *herodot.DefaultError {
	return herodot.ErrBadRequest().WithError("no resumable session found").WithReasonf("The browser does not contain the necessary cookie to resume the session. This is a security violation and was blocked. Please clear your browser's cookies and cache and try again!")
}

const CookieName = "ory_kratos_continuity"

type (
	ManagementProvider interface {
		ContinuityManager() *Manager
	}
	managerDependencies interface {
		PersistenceProvider
		otelx.Provider
	}
	Manager struct {
		d managerDependencies
	}
)

func NewManager(d managerDependencies) *Manager {
	return &Manager{d: d}
}

type managerOptions struct {
	iid          uuid.UUID
	ttl          time.Duration
	setExpiresIn time.Duration
	payload      json.RawMessage
	payloadRaw   any
}

type ManagerOption func(*managerOptions) error

func newManagerOptions(opts []ManagerOption) (*managerOptions, error) {
	var o = &managerOptions{
		ttl: time.Minute * 10,
	}
	for _, opt := range opts {
		if err := opt(o); err != nil {
			return nil, err
		}
	}
	return o, nil
}

func WithIdentity(i *identity.Identity) ManagerOption {
	return func(o *managerOptions) error {
		if i != nil {
			o.iid = i.ID
		}
		return nil
	}
}

func WithLifespan(ttl time.Duration) ManagerOption {
	return func(o *managerOptions) error {
		o.ttl = ttl
		return nil
	}
}

func WithPayload(payload any) ManagerOption {
	return func(o *managerOptions) error {
		var b bytes.Buffer
		if err := json.NewEncoder(&b).Encode(payload); err != nil {
			return errors.WithStack(err)
		}
		o.payload = b.Bytes()
		o.payloadRaw = payload
		return nil
	}
}

func WithExpireInsteadOfDelete(duration time.Duration) ManagerOption {
	return func(o *managerOptions) error {
		o.setExpiresIn = duration
		return nil
	}
}

func (m *Manager) Pause(ctx context.Context, w http.ResponseWriter, r *http.Request, name string, store ContainerReferenceStore, opts ...ManagerOption) (containerID uuid.UUID, err error) {
	ctx, span := m.d.Tracer(ctx).Tracer().Start(ctx, "continuity.ManagerDefault.Pause")
	defer otelx.End(span, &err)
	if len(name) == 0 {
		return uuid.Nil, errors.Errorf("continuity container name must be set")
	}

	o, err := newManagerOptions(opts)
	if err != nil {
		return uuid.Nil, err
	}
	c := NewContainer(name, *o)

	if err := m.d.ContinuityPersister().SaveContinuitySession(ctx, c); err != nil {
		return uuid.Nil, errors.WithStack(err)
	}

	if err := store.Store(ctx, w, r, name, c.ID); err != nil {
		return uuid.Nil, err
	}

	return c.ID, nil
}

func (m *Manager) Continue(ctx context.Context, w http.ResponseWriter, r *http.Request, name string, store ContainerReferenceStore, opts ...ManagerOption) (container *Container, err error) {
	ctx, span := m.d.Tracer(ctx).Tracer().Start(ctx, "continuity.ManagerDefault.Continue")
	defer otelx.End(span, &err)

	o, err := newManagerOptions(opts)
	if err != nil {
		return nil, err
	}

	sid, err := store.Retrieve(ctx, w, r, name)
	if err != nil {
		return nil, err
	}

	container, err = m.d.ContinuityPersister().GetContinuitySession(ctx, sid)
	if errors.Is(err, sqlcon.ErrNoRows()) {
		_ = store.Clear(ctx, w, r, name)
		return nil, errors.WithStack(ErrNotResumable().WithDebugf("Resumable ID could not be found in the datastore: %+v", err))
	} else if err != nil {
		_ = store.Clear(ctx, w, r, name)
		return nil, err
	} else if container.ExpiresAt.Before(time.Now()) {
		_ = store.Clear(ctx, w, r, name)
		return nil, errors.WithStack(ErrNotResumable().WithDebugf("Resumable session has expired"))
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
			ctx, container.ID,
			time.Now().UTC().Add(o.setExpiresIn).Truncate(time.Second),
		); err != nil && !errors.Is(err, sqlcon.ErrNoRows()) {
			return nil, err
		}
	} else {
		_ = store.Clear(ctx, w, r, name)
		if err := m.d.ContinuityPersister().DeleteContinuitySession(ctx, container.ID); err != nil && !errors.Is(err, sqlcon.ErrNoRows()) {
			return nil, err
		}
	}

	return container, nil
}

func (m Manager) Abort(ctx context.Context, w http.ResponseWriter, r *http.Request, name string, store ContainerReferenceStore) (err error) {
	ctx, span := m.d.Tracer(ctx).Tracer().Start(ctx, "continuity.ManagerDefault.Abort")
	defer otelx.End(span, &err)

	sid, err := store.Retrieve(ctx, w, r, name)
	if errors.Is(err, ErrNotResumable()) {
		return nil
	} else if err != nil {
		return err
	}

	_ = store.Clear(ctx, w, r, name)

	if err := m.d.ContinuityPersister().DeleteContinuitySession(ctx, sid); err != nil && !errors.Is(err, sqlcon.ErrNoRows()) {
		return errors.WithStack(err)
	}

	return nil
}
