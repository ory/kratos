// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package identity

import (
	"context"
	"encoding/json"
	"reflect"
	"sort"

	"go.opentelemetry.io/otel/trace"

	"github.com/ory/kratos/schema"
	"github.com/ory/kratos/x/events"
	"github.com/ory/x/sqlcon"

	"github.com/ory/x/otelx"

	"github.com/ory/kratos/x"

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
		PrivilegedPoolProvider
		x.TracingProvider
		courier.Provider
		ValidationProvider
		ActiveCredentialsCounterStrategyProvider
		x.LoggingProvider
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

	if err := i.SetAvailableAAL(ctx, m); err != nil {
		return err
	}

	if err := m.r.PrivilegedIdentityPool().CreateIdentity(ctx, i); err != nil {
		if errors.Is(err, sqlcon.ErrUniqueViolation) {
			return m.findExistingAuthMethod(ctx, err, i)
		}
		return err
	}

	trace.SpanFromContext(ctx).AddEvent(events.NewIdentityCreated(ctx, i.ID))
	return nil
}

func (m *Manager) ConflictingIdentity(ctx context.Context, i *Identity) (found *Identity, foundConflictAddress string, err error) {
	for ct, cred := range i.Credentials {
		for _, id := range cred.Identifiers {
			found, _, err = m.r.PrivilegedIdentityPool().FindByCredentialsIdentifier(ctx, ct, id)
			if err != nil {
				continue
			}

			// FindByCredentialsIdentifier does not expand identity credentials.
			if err = m.r.PrivilegedIdentityPool().HydrateIdentityAssociations(ctx, found, ExpandCredentials); err != nil {
				return nil, "", err
			}

			return found, id, nil
		}
	}

	// If the conflict is not in the identifiers table, it is coming from the verifiable or recovery address.
	for _, va := range i.VerifiableAddresses {
		conflictingAddress, err := m.r.PrivilegedIdentityPool().FindVerifiableAddressByValue(ctx, va.Via, va.Value)
		if errors.Is(err, sqlcon.ErrNoRows) {
			continue
		} else if err != nil {
			return nil, "", err
		}

		foundConflictAddress = conflictingAddress.Value
		found, err = m.r.PrivilegedIdentityPool().GetIdentity(ctx, conflictingAddress.IdentityID, ExpandCredentials)
		if err != nil {
			return nil, "", err
		}

		return found, foundConflictAddress, nil
	}

	// Last option: check the recovery address
	for _, va := range i.RecoveryAddresses {
		conflictingAddress, err := m.r.PrivilegedIdentityPool().FindRecoveryAddressByValue(ctx, va.Via, va.Value)
		if errors.Is(err, sqlcon.ErrNoRows) {
			continue
		} else if err != nil {
			return nil, "", err
		}

		foundConflictAddress = conflictingAddress.Value
		found, err = m.r.PrivilegedIdentityPool().GetIdentity(ctx, conflictingAddress.IdentityID, ExpandCredentials)
		if err != nil {
			return nil, "", err
		}

		return found, foundConflictAddress, nil
	}

	return nil, "", sqlcon.ErrNoRows
}

func (m *Manager) findExistingAuthMethod(ctx context.Context, e error, i *Identity) (err error) {
	if !m.r.Config().SelfServiceFlowRegistrationLoginHints(ctx) {
		return &ErrDuplicateCredentials{error: e}
	}

	found, foundConflictAddress, err := m.ConflictingIdentity(ctx, i)
	if err != nil {
		if errors.Is(err, sqlcon.ErrNoRows) {
			return &ErrDuplicateCredentials{error: e}
		}
		return err
	}

	// We need to sort the credentials for the error message to be deterministic.
	var creds []Credentials
	for _, cred := range found.Credentials {
		creds = append(creds, cred)
	}
	sort.Slice(creds, func(i, j int) bool {
		return creds[i].Type < creds[j].Type
	})

	for _, cred := range creds {
		if cred.Config == nil {
			continue
		}

		// Basically, we only have password, oidc, and webauthn as first factor credentials.
		// We don't care about second factor, because they don't help the user understand how to sign
		// in to the first factor (obviously).
		switch cred.Type {
		case CredentialsTypePassword:
			identifierHint := foundConflictAddress
			if len(cred.Identifiers) > 0 {
				identifierHint = cred.Identifiers[0]
			}
			return &ErrDuplicateCredentials{
				error:                e,
				availableCredentials: []CredentialsType{cred.Type},
				identifierHint:       identifierHint,
			}
		case CredentialsTypeOIDC:
			var cfg CredentialsOIDC
			if err := json.Unmarshal(cred.Config, &cfg); err != nil {
				return errors.WithStack(herodot.ErrInternalServerError.WithReasonf("Unable to JSON decode identity credentials %s for identity %s.", cred.Type, found.ID))
			}

			available := make([]string, 0, len(cfg.Providers))
			for _, provider := range cfg.Providers {
				available = append(available, provider.Provider)
			}

			return &ErrDuplicateCredentials{
				error:                  e,
				availableCredentials:   []CredentialsType{cred.Type},
				availableOIDCProviders: available,
				identifierHint:         foundConflictAddress,
			}
		case CredentialsTypeWebAuthn:
			var cfg CredentialsWebAuthnConfig
			if err := json.Unmarshal(cred.Config, &cfg); err != nil {
				return errors.WithStack(herodot.ErrInternalServerError.WithReasonf("Unable to JSON decode identity credentials %s for identity %s.", cred.Type, found.ID))
			}

			identifierHint := foundConflictAddress
			if len(cred.Identifiers) > 0 {
				identifierHint = cred.Identifiers[0]
			}

			for _, webauthn := range cfg.Credentials {
				if webauthn.IsPasswordless {
					return &ErrDuplicateCredentials{
						error:                e,
						availableCredentials: []CredentialsType{cred.Type},
						identifierHint:       identifierHint,
					}
				}
			}
		}
	}

	// Still not found? Return generic error.
	return &ErrDuplicateCredentials{error: e}
}

type ErrDuplicateCredentials struct {
	error

	availableCredentials   []CredentialsType
	availableOIDCProviders []string
	identifierHint         string
}

var _ schema.DuplicateCredentialsHinter = (*ErrDuplicateCredentials)(nil)

func (e *ErrDuplicateCredentials) Unwrap() error {
	return e.error
}

func (e *ErrDuplicateCredentials) AvailableCredentials() []string {
	res := make([]string, len(e.availableCredentials))
	for k, v := range e.availableCredentials {
		res[k] = string(v)
	}
	return res
}

func (e *ErrDuplicateCredentials) AvailableOIDCProviders() []string {
	return e.availableOIDCProviders
}

func (e *ErrDuplicateCredentials) IdentifierHint() string {
	return e.identifierHint
}
func (e *ErrDuplicateCredentials) HasHints() bool {
	return len(e.availableCredentials) > 0 || len(e.availableOIDCProviders) > 0 || len(e.identifierHint) > 0
}

func (m *Manager) CreateIdentities(ctx context.Context, identities []*Identity, opts ...ManagerOption) (err error) {
	ctx, span := m.r.Tracer(ctx).Tracer().Start(ctx, "identity.Manager.CreateIdentities")
	defer otelx.End(span, &err)

	for _, i := range identities {
		if i.SchemaID == "" {
			i.SchemaID = m.r.Config().DefaultIdentityTraitsSchemaID(ctx)
		}

		if err := i.SetAvailableAAL(ctx, m); err != nil {
			return err
		}

		o := newManagerOptions(opts)
		if err := m.ValidateIdentity(ctx, i, o); err != nil {
			return err
		}
	}

	if err := m.r.PrivilegedIdentityPool().CreateIdentities(ctx, identities...); err != nil {
		return err
	}

	for _, i := range identities {
		trace.SpanFromContext(ctx).AddEvent(events.NewIdentityCreated(ctx, i.ID))
	}

	return nil
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

	original, err := m.r.PrivilegedIdentityPool().GetIdentityConfidential(ctx, updated.ID)
	if err != nil {
		return err
	}

	if err := m.requiresPrivilegedAccess(ctx, original, updated, o); err != nil {
		return err
	}

	if err := updated.SetAvailableAAL(ctx, m); err != nil {
		return err
	}

	return m.r.PrivilegedIdentityPool().UpdateIdentity(ctx, updated)
}

func (m *Manager) UpdateSchemaID(ctx context.Context, id uuid.UUID, schemaID string, opts ...ManagerOption) (err error) {
	ctx, span := m.r.Tracer(ctx).Tracer().Start(ctx, "identity.Manager.UpdateSchemaID")
	defer otelx.End(span, &err)

	o := newManagerOptions(opts)
	original, err := m.r.PrivilegedIdentityPool().GetIdentityConfidential(ctx, id)
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

	trace.SpanFromContext(ctx).AddEvent(events.NewIdentityUpdated(ctx, id))
	return m.r.PrivilegedIdentityPool().UpdateIdentity(ctx, original)
}

func (m *Manager) SetTraits(ctx context.Context, id uuid.UUID, traits Traits, opts ...ManagerOption) (_ *Identity, err error) {
	ctx, span := m.r.Tracer(ctx).Tracer().Start(ctx, "identity.Manager.SetTraits")
	defer otelx.End(span, &err)

	o := newManagerOptions(opts)
	original, err := m.r.PrivilegedIdentityPool().GetIdentityConfidential(ctx, id)
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

	trace.SpanFromContext(ctx).AddEvent(events.NewIdentityUpdated(ctx, id))
	return m.r.PrivilegedIdentityPool().UpdateIdentity(ctx, updated)
}

func (m *Manager) ValidateIdentity(ctx context.Context, i *Identity, o *ManagerOptions) (err error) {
	// This trace is more noisy than it's worth in diagnostic power.
	// ctx, span := m.r.Tracer(ctx).Tracer().Start(ctx, "identity.Manager.validate")
	// defer otelx.End(span, &err)

	if err := m.r.IdentityValidator().Validate(ctx, i); err != nil {
		if _, ok := errorsx.Cause(err).(*jsonschema.ValidationError); ok && !o.ExposeValidationErrors {
			return herodot.ErrBadRequest.WithReasonf("%s", err).WithWrap(err)
		}
		return err
	}

	return nil
}

func (m *Manager) CountActiveFirstFactorCredentials(ctx context.Context, i *Identity) (count int, err error) {
	// This trace is more noisy than it's worth in diagnostic power.
	// ctx, span := m.r.Tracer(ctx).Tracer().Start(ctx, "identity.Manager.CountActiveFirstFactorCredentials")
	// defer otelx.End(span, &err)

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
	// This trace is more noisy than it's worth in diagnostic power.
	// ctx, span := m.r.Tracer(ctx).Tracer().Start(ctx, "identity.Manager.CountActiveMultiFactorCredentials")
	// defer otelx.End(span, &err)

	for _, strategy := range m.r.ActiveCredentialsCounterStrategies(ctx) {
		current, err := strategy.CountActiveMultiFactorCredentials(i.Credentials)
		if err != nil {
			return 0, err
		}

		count += current
	}
	return count, nil
}
