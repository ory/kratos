// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package identity

import (
	"context"

	"github.com/pkg/errors"

	"github.com/ory/herodot"

	"github.com/tidwall/sjson"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/schema"
	"github.com/ory/x/otelx"
)

type (
	validatorDependencies interface {
		schema.IdentitySchemaProvider
		config.Provider
	}
	Validator struct {
		v *schema.Validator
		d validatorDependencies
	}
	ValidationProvider interface {
		IdentityValidator() *Validator
	}
)

func NewValidator(d validatorDependencies) *Validator {
	return &Validator{v: schema.NewValidator(), d: d}
}

func (v *Validator) ValidateWithRunner(ctx context.Context, i *Identity, runners ...schema.ValidateExtension) error {
	runner, err := schema.NewExtensionRunner(ctx, schema.WithValidateRunners(runners...))
	if err != nil {
		return err
	}

	ss, err := v.d.IdentityTraitsSchemas(ctx)
	if err != nil {
		return err
	}

	s, err := ss.GetByID(i.SchemaID)
	if err != nil {
		return err
	}

	if len(i.Traits) == 0 {
		i.Traits = []byte(`{}`)
	}

	traits, err := sjson.SetRawBytes([]byte(`{}`), "traits", i.Traits)
	if err != nil {
		return errors.WithStack(herodot.ErrBadRequest.WithError(err.Error()))
	}

	return v.v.Validate(ctx, s.URL.String(), traits, schema.WithExtensionRunner(runner))
}

func (v *Validator) Validate(ctx context.Context, i *Identity) error {
	return otelx.WithSpan(ctx, "identity.Validator.Validate", func(ctx context.Context) error {
		return v.ValidateWithRunner(ctx, i,
			NewSchemaExtensionCredentials(i),
			NewSchemaExtensionVerification(i, v.d.Config().SelfServiceFlowVerificationRequestLifespan(ctx)),
			NewSchemaExtensionRecovery(i),
		)
	})
}
