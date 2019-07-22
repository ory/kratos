package selfservice

import (
	"context"
	"sync"

	"github.com/pkg/errors"

	"github.com/ory/herodot"

	"github.com/ory/hive-cloud/hive/identity"
)

var _ RegistrationRequestManager = new(RequestManagerMemory)
var _ LoginRequestManager = new(RequestManagerMemory)

type RequestManagerMemory struct {
	sync.RWMutex
	sir map[string]LoginRequest
	sur map[string]RegistrationRequest
}

func NewRequestManagerMemory() *RequestManagerMemory {
	return &RequestManagerMemory{
		sir: make(map[string]LoginRequest),
		sur: make(map[string]RegistrationRequest),
	}
}

func (m *RequestManagerMemory) cr(r interface{}) error {
	m.Lock()
	defer m.Unlock()
	switch t := r.(type) {
	case *LoginRequest:
		m.sir[t.ID] = *t
	case *RegistrationRequest:
		m.sur[t.ID] = *t
	default:
		panic("Unknown type")
	}
	return nil
}

func (m *RequestManagerMemory) CreateLoginRequest(ctx context.Context, r *LoginRequest) error {
	return m.cr(r)
}

func (m *RequestManagerMemory) CreateRegistrationRequest(ctx context.Context, r *RegistrationRequest) error {
	return m.cr(r)
}

func (m *RequestManagerMemory) GetLoginRequest(ctx context.Context, id string) (*LoginRequest, error) {
	m.RLock()
	defer m.RUnlock()
	if r, ok := m.sir[id]; ok {
		return &r, nil
	}

	return nil, errors.WithStack(herodot.ErrNotFound.WithReasonf("Unable to find request: %s", id))
}

func (m *RequestManagerMemory) GetRegistrationRequest(ctx context.Context, id string) (*RegistrationRequest, error) {
	m.RLock()
	defer m.RUnlock()
	if r, ok := m.sur[id]; ok {
		return &r, nil
	}

	return nil, errors.WithStack(herodot.ErrNotFound.WithReasonf("Unable to find request: %s", id))
}

func (m *RequestManagerMemory) UpdateRegistrationRequest(ctx context.Context, id string, t identity.CredentialsType, c interface{}) error {
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

	me.Config = c
	r.Methods[t] = me
	m.sur[id] = *r

	return nil
}

func (m *RequestManagerMemory) UpdateLoginRequest(ctx context.Context, id string, t identity.CredentialsType, c interface{}) error {
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

	me.Config = c
	r.Methods[t] = me
	m.sir[id] = *r

	return nil
}
