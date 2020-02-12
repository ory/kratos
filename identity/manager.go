package identity

import (
	"context"
	"reflect"
	"time"

	"github.com/gofrs/uuid"
	"github.com/mohae/deepcopy"
	"github.com/ory/herodot"
	"github.com/pkg/errors"

	"github.com/ory/kratos/courier"
	"github.com/ory/kratos/driver/configuration"
)

var errProtectedFieldModified = herodot.ErrInternalServerError.
	WithReasonf(`A field was modified that updates one or more credentials-related settings. This action was blocked because a unprivileged DBAL method was used to execute the update. This is either a configuration issue, or a bug.`)

type (
	managerDependencies interface {
		PoolProvider
		courier.Provider
		ValidationProvider
	}
	ManagementProvider interface {
		IdentityManager() *Manager
	}
	Manager struct {
		r managerDependencies
		c configuration.Provider
	}
)

func NewManager(r managerDependencies, c configuration.Provider) *Manager {
	return &Manager{r: r, c: c}
}

func (m *Manager) Create(ctx context.Context, i *Identity) error {
	if err := m.r.IdentityValidator().Validate(i); err != nil {
		return err
	}

	return m.r.IdentityPool().(PrivilegedPool).CreateIdentity(ctx, i)
}

func (m *Manager) Update(ctx context.Context, i *Identity) error {
	if err := m.r.IdentityValidator().Validate(i); err != nil {
		return err
	}

	return m.r.IdentityPool().(PrivilegedPool).UpdateIdentity(ctx, i)
}

func (m *Manager) UpdateUnprotectedTraits(ctx context.Context, id uuid.UUID, traits Traits) error {
	identity, err := m.r.IdentityPool().(PrivilegedPool).GetIdentityConfidential(ctx, id)
	if err != nil {
		return err
	}

	original := deepcopy.Copy(identity).(*Identity)
	identity.Traits = traits
	if err := m.r.IdentityValidator().Validate(identity); err != nil {
		return err
	}

	if CredentialsEqual(identity.Credentials, original.Credentials) {
		return errors.WithStack(errProtectedFieldModified)
	}

	if !reflect.DeepEqual(original.Addresses, identity.Addresses) {
		return errors.WithStack(errProtectedFieldModified)
	}

	return m.r.IdentityPool().(PrivilegedPool).UpdateIdentity(ctx, identity)
}

func (m *Manager) RefreshVerifyAddress(ctx context.Context, address *VerifiableAddress) error {
	code, err := NewVerifyCode()
	if err != nil {
		return err
	}

	address.Code = code
	address.ExpiresAt = time.Now().UTC().Add(m.c.SelfServiceVerificationLinkLifespan())
	return m.r.IdentityPool().(PrivilegedPool).UpdateVerifiableAddress(ctx, address)
}
