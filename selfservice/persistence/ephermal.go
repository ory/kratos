package persistence

import (
	"context"
	"sync"

	"github.com/pkg/errors"

	"github.com/ory/herodot"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/selfservice/flow/profile"
	"github.com/ory/kratos/selfservice/flow/registration"
)

var _ registration.RequestPersister = new(RequestManagerMemory)
var _ login.RequestPersister = new(RequestManagerMemory)
var _ profile.RequestPersister = new(RequestManagerMemory)

type RequestManagerMemory struct {
	sync.RWMutex
	sir map[string]login.Request
	sur map[string]registration.Request
	pr  map[string]profile.Request
}

func NewRequestManagerMemory() *RequestManagerMemory {
	return &RequestManagerMemory{
		sir: make(map[string]login.Request),
		sur: make(map[string]registration.Request),
		pr:  make(map[string]profile.Request),
	}
}

func (m *RequestManagerMemory) cr(r interface{}) error {
	m.Lock()
	defer m.Unlock()
	switch t := r.(type) {
	case *login.Request:
		m.sir[t.ID] = *t
	case *registration.Request:
		m.sur[t.ID] = *t
	case *profile.Request:
		m.pr[t.ID] = *t
	default:
		panic("Unknown type")
	}
	return nil
}

func (m *RequestManagerMemory) CreateLoginRequest(ctx context.Context, r *login.Request) error {
	return m.cr(r)
}

func (m *RequestManagerMemory) CreateRegistrationRequest(ctx context.Context, r *registration.Request) error {
	return m.cr(r)
}

func (m *RequestManagerMemory) GetLoginRequest(ctx context.Context, id string) (*login.Request, error) {
	m.RLock()
	defer m.RUnlock()
	if r, ok := m.sir[id]; ok {
		return &r, nil
	}

	return nil, errors.WithStack(herodot.ErrNotFound.WithReasonf("Unable to find request: %s", id))
}

func (m *RequestManagerMemory) GetRegistrationRequest(ctx context.Context, id string) (*registration.Request, error) {
	m.RLock()
	defer m.RUnlock()
	if r, ok := m.sur[id]; ok {
		return &r, nil
	}

	return nil, errors.WithStack(herodot.ErrNotFound.WithReasonf("Unable to find request: %s", id))
}

func (m *RequestManagerMemory) UpdateRegistrationRequest(ctx context.Context, id string, t identity.CredentialsType, c *registration.RequestMethod) error {
	r, err := m.GetRegistrationRequest(ctx, id)
	if err != nil {
		return err
	}

	m.Lock()
	defer m.Unlock()

	me, ok := r.Methods[t]
	if !ok {
		return errors.WithStack(herodot.ErrInternalServerError.WithReasonf(`Expected registration request "%s" to have credentials type "%s", indicating an internal error.`, id, t))
	}

	me.Config = c.Config
	r.Active = t
	r.Methods[t] = me
	m.sur[id] = *r

	return nil
}

func (m *RequestManagerMemory) UpdateLoginRequest(ctx context.Context, id string, t identity.CredentialsType, c *login.RequestMethod) error {
	r, err := m.GetLoginRequest(ctx, id)
	if err != nil {
		return err
	}

	m.Lock()
	defer m.Unlock()

	me, ok := r.Methods[t]
	if !ok {
		return errors.WithStack(herodot.ErrInternalServerError.WithReasonf(`Expected login request "%s" to have credentials type "%s", indicating an internal error.`, id, t))
	}

	me.Config = c.Config
	r.Active = t
	r.Methods[t] = me
	m.sir[id] = *r

	return nil
}

func (m *RequestManagerMemory) CreateProfileRequest(ctx context.Context, r *profile.Request) error {
	return m.cr(r)
}

func (m *RequestManagerMemory) GetProfileRequest(ctx context.Context, id string) (*profile.Request, error) {
	m.RLock()
	defer m.RUnlock()
	if r, ok := m.pr[id]; ok {
		return &r, nil
	}

	return nil, errors.WithStack(herodot.ErrNotFound.WithReasonf("Unable to find request: %s", id))
}

func (m *RequestManagerMemory) UpdateProfileRequest(ctx context.Context, id string, request *profile.Request) error {
	m.Lock()
	defer m.Unlock()
	r, ok := m.pr[id]
	if !ok {
		return errors.WithStack(herodot.ErrNotFound.WithReasonf("Unable to find request: %s", id))
	}

	*r.Form = *request.Form
	m.pr[id] = r
	return nil
}
