// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package identity

import (
	"context"
	"reflect"

	"github.com/ory/kratos/x"
	"github.com/ory/x/otelx"

	"github.com/ory/kratos/driver/config"

	"github.com/gofrs/uuid"

	"github.com/mohae/deepcopy"
	"github.com/pkg/errors"

	"github.com/ory/herodot"
	"github.com/ory/jsonschema/v3"
	"github.com/ory/x/errorsx"

	"github.com/ory/kratos/courier"
)

var ErrProtectedFieldModified = herodot.ErrForbidden.
	WithReasonf(`A field was modified that updates one or more credentials-related settings. This action was blocked because an unprivileged method was used to execute the update. This is either a configuration issue or a bug and should be reported to the system administrator.`)

type (
	managerDependencies interface {
		config.Provider
		PoolProvider
		x.TracingProvider
		courier.Provider
		ValidationProvider
		ActiveCredentialsCounterStrategyProvider
	}
	ManagementProvider interface {
		IdentityManager() *Manager
	}
	Manager struct {
		r managerDependencies
	}

	ManagerOptions struct {
		ExposeValidationErrors    bool
		AllowWriteProtectedTraits bool
	}

	ManagerOption func(*ManagerOptions)
)

func NewManager(r managerDependencies) *Manager {
	return &Manager{r: r}
}

func ManagerExposeValidationErrorsForInternalTypeAssertion(options *ManagerOptions) {
	options.ExposeValidationErrors = true
}

func ManagerAllowWriteProtectedTraits(options *ManagerOptions) {
	options.AllowWriteProtectedTraits = true
}

func newManagerOptions(opts []ManagerOption) *ManagerOptions {
	var o ManagerOptions
	for _, f := range opts {
		f(&o)
	}
	return &o
}

func (m *Manager) Create(ctx context.Context, i *Identity, opts ...ManagerOption) (err error) {
	ctx, span := m.r.Tracer(ctx).Tracer().Start(ctx, "identity.Manager.Create")
	defer otelx.End(span, &err)

	if i.SchemaID == "" {
		i.SchemaID = m.r.Config().DefaultIdentityTraitsSchemaID(ctx)
	}

	o := newManagerOptions(opts)
	if err := m.ValidateIdentity(ctx, i, o); err != nil {
		return err
	}

	return m.r.IdentityPool().(PrivilegedPool).CreateIdentity(ctx, i)
}

func (m *Manager) requiresPrivilegedAccess(ctx context.Context, original, updated *Identity, o *ManagerOptions) (err error) {
	_, span := m.r.Tracer(ctx).Tracer().Start(ctx, "identity.Manager.requiresPrivilegedAccess")
	defer otelx.End(span, &err)

	if !o.AllowWriteProtectedTraits {
		if !CredentialsEqual(updated.Credentials, original.Credentials) {
			// reset the identity
			*updated = *original
			return errors.WithStack(ErrProtectedFieldModified)
		}

		if !reflect.DeepEqual(original.VerifiableAddresses, updated.VerifiableAddresses) &&
			/* prevent nil != []string{} */
			len(original.VerifiableAddresses)+len(updated.VerifiableAddresses) != 0 {
			// reset the identity
			*updated = *original
			return errors.WithStack(ErrProtectedFieldModified)
		}
	}
	return nil
}

func (m *Manager) Update(ctx context.Context, updated *Identity, opts ...ManagerOption) (err error) {
	ctx, span := m.r.Tracer(ctx).Tracer().Start(ctx, "identity.Manager.Update")
	defer otelx.End(span, &err)

	o := newManagerOptions(opts)
	if err := m.ValidateIdentity(ctx, updated, o); err != nil {
		return err
	}

	original, err := m.r.IdentityPool().(PrivilegedPool).GetIdentityConfidential(ctx, updated.ID)
	if err != nil {
		return err
	}

	if err := m.requiresPrivilegedAccess(ctx, original, updated, o); err != nil {
		return err
	}

	return m.r.IdentityPool().(PrivilegedPool).UpdateIdentity(ctx, updated)
}

func (m *Manager) UpdateSchemaID(ctx context.Context, id uuid.UUID, schemaID string, opts ...ManagerOption) (err error) {
	ctx, span := m.r.Tracer(ctx).Tracer().Start(ctx, "identity.Manager.UpdateSchemaID")
	defer otelx.End(span, &err)

	o := newManagerOptions(opts)
	original, err := m.r.IdentityPool().(PrivilegedPool).GetIdentityConfidential(ctx, id)
	if err != nil {
		return err
	}

	if !o.AllowWriteProtectedTraits && original.SchemaID != schemaID {
		return errors.WithStack(ErrProtectedFieldModified)
	}

	original.SchemaID = schemaID
	if err := m.ValidateIdentity(ctx, original, o); err != nil {
		return err
	}

	return m.r.IdentityPool().(PrivilegedPool).UpdateIdentity(ctx, original)
}

func (m *Manager) SetTraits(ctx context.Context, id uuid.UUID, traits Traits, opts ...ManagerOption) (_ *Identity, err error) {
	ctx, span := m.r.Tracer(ctx).Tracer().Start(ctx, "identity.Manager.SetTraits")
	defer otelx.End(span, &err)

	o := newManagerOptions(opts)
	original, err := m.r.IdentityPool().(PrivilegedPool).GetIdentityConfidential(ctx, id)
	if err != nil {
		return nil, err
	}

	// original is used to check whether protected traits were modified
	updated := deepcopy.Copy(original).(*Identity)
	updated.Traits = traits
	if err := m.ValidateIdentity(ctx, updated, o); err != nil {
		return nil, err
	}

	if err := m.requiresPrivilegedAccess(ctx, original, updated, o); err != nil {
		return nil, err
	}

	return updated, nil
}

func (m *Manager) UpdateTraits(ctx context.Context, id uuid.UUID, traits Traits, opts ...ManagerOption) (err error) {
	ctx, span := m.r.Tracer(ctx).Tracer().Start(ctx, "identity.Manager.UpdateTraits")
	defer otelx.End(span, &err)

	updated, err := m.SetTraits(ctx, id, traits, opts...)
	if err != nil {
		return err
	}

	return m.r.IdentityPool().(PrivilegedPool).UpdateIdentity(ctx, updated)
}

func (m *Manager) ValidateIdentity(ctx context.Context, i *Identity, o *ManagerOptions) (err error) {
	ctx, span := m.r.Tracer(ctx).Tracer().Start(ctx, "identity.Manager.validate")
	defer otelx.End(span, &err)

	if err := m.r.IdentityValidator().Validate(ctx, i); err != nil {
		if _, ok := errorsx.Cause(err).(*jsonschema.ValidationError); ok && !o.ExposeValidationErrors {
			return herodot.ErrBadRequest.WithReasonf("%s", err).WithWrap(err)
		}
		return err
	}

	return nil
}

func (m *Manager) CountActiveFirstFactorCredentials(ctx context.Context, i *Identity) (count int, err error) {
	ctx, span := m.r.Tracer(ctx).Tracer().Start(ctx, "identity.Manager.CountActiveFirstFactorCredentials")
	defer otelx.End(span, &err)

	for _, strategy := range m.r.ActiveCredentialsCounterStrategies(ctx) {
		current, err := strategy.CountActiveFirstFactorCredentials(i.Credentials)
		if err != nil {
			return 0, err
		}

		count += current
	}
	return count, nil
}

func (m *Manager) CountActiveMultiFactorCredentials(ctx context.Context, i *Identity) (count int, err error) {
	ctx, span := m.r.Tracer(ctx).Tracer().Start(ctx, "identity.Manager.CountActiveMultiFactorCredentials")
	defer otelx.End(span, &err)

	for _, strategy := range m.r.ActiveCredentialsCounterStrategies(ctx) {
		current, err := strategy.CountActiveMultiFactorCredentials(i.Credentials)
		if err != nil {
			return 0, err
		}

		count += current
	}
	return count, nil
}
