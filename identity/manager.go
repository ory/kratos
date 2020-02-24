package identity

import (
	"context"
	"reflect"
	"time"

	"github.com/gofrs/uuid"
	"github.com/mohae/deepcopy"
	"github.com/pkg/errors"

	"github.com/ory/herodot"
	"github.com/ory/jsonschema/v3"
	"github.com/ory/x/errorsx"

	"github.com/ory/kratos/courier"
	"github.com/ory/kratos/driver/configuration"
)

var ErrProtectedFieldModified = herodot.ErrForbidden.
	WithReasonf(`A field was modified that updates one or more credentials-related settings. This action was blocked because an unprivileged method was used to execute the update. This is either a configuration issue or a bug and should be reported to the system administrator.`)

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

	managerOptions struct {
		ExposeValidationErrors bool
	}

	managerOption func(*managerOptions)
)

func NewManager(r managerDependencies, c configuration.Provider) *Manager {
	return &Manager{r: r, c: c}
}

func ManagerExposeValidationErrors(options *managerOptions) {
	options.ExposeValidationErrors = true
}

func newManagerOptions(opts []managerOption) *managerOptions {
	var o managerOptions
	for _, f := range opts {
		f(&o)
	}
	return &o
}

func (m *Manager) Create(ctx context.Context, i *Identity, opts ...managerOption) error {
	o := newManagerOptions(opts)
	if err := m.validate(i, o); err != nil {
		return err
	}

	return m.r.IdentityPool().(PrivilegedPool).CreateIdentity(ctx, i)
}

func (m *Manager) Update(ctx context.Context, i *Identity, opts ...managerOption) error {
	o := newManagerOptions(opts)
	if err := m.validate(i, o); err != nil {
		return err
	}

	return m.r.IdentityPool().(PrivilegedPool).UpdateIdentity(ctx, i)
}

func (m *Manager) UpdateTraits(ctx context.Context, id uuid.UUID, traits Traits, allowAllTraits bool, opts ...managerOption) error {
	o := newManagerOptions(opts)

	identity, err := m.r.IdentityPool().(PrivilegedPool).GetIdentityConfidential(ctx, id)
	if err != nil {
		return err
	}

	original := deepcopy.Copy(identity).(*Identity)
	identity.Traits = traits
	if err := m.validate(identity, o); err != nil {
		return err
	}

	if !allowAllTraits {
		if !CredentialsEqual(identity.Credentials, original.Credentials) {
			return errors.WithStack(ErrProtectedFieldModified)
		}

		if !reflect.DeepEqual(original.Addresses, identity.Addresses) &&
			/* prevent nil != []string{} */
			len(original.Addresses)+len(identity.Addresses) != 0 {
			return errors.WithStack(ErrProtectedFieldModified)
		}
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

func (m *Manager) validate(i *Identity, o *managerOptions) error {
	if err := m.r.IdentityValidator().Validate(i); err != nil {
		if _, ok := errorsx.Cause(err).(*jsonschema.ValidationError); ok && !o.ExposeValidationErrors {
			return errors.WithStack(herodot.ErrBadRequest.WithReasonf("%s", err))
		}
		return err
	}

	return nil
}
