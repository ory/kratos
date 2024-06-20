// Copyright © 2024 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package login

import (
	"net/http"

	"github.com/pkg/errors"

	"github.com/ory/kratos/identity"
)

type OneStepFormHydrator interface {
	PopulateLoginMethod(r *http.Request, requestedAAL identity.AuthenticatorAssuranceLevel, sr *Flow) error
}

type FormHydrator interface {
	PopulateLoginMethodFirstFactorRefresh(r *http.Request, sr *Flow) error
	PopulateLoginMethodFirstFactor(r *http.Request, sr *Flow) error
	PopulateLoginMethodSecondFactor(r *http.Request, sr *Flow) error
	PopulateLoginMethodSecondFactorRefresh(r *http.Request, sr *Flow) error
	PopulateLoginMethodIdentifierFirstCredentials(r *http.Request, sr *Flow, options ...FormHydratorModifier) error
	PopulateLoginMethodIdentifierFirstIdentification(r *http.Request, sr *Flow) error
}

var ErrBreakLoginPopulate = errors.New("skip rest of login form population")

type FormHydratorOptions struct {
	IdentityHint *identity.Identity
	Identifier   string
}

type FormHydratorModifier func(o *FormHydratorOptions)

func WithIdentityHint(i *identity.Identity) FormHydratorModifier {
	return func(o *FormHydratorOptions) {
		o.IdentityHint = i
	}
}

func WithIdentifier(i string) FormHydratorModifier {
	return func(o *FormHydratorOptions) {
		o.Identifier = i
	}
}

func NewFormHydratorOptions(modifiers []FormHydratorModifier) *FormHydratorOptions {
	o := new(FormHydratorOptions)
	for _, m := range modifiers {
		m(o)
	}
	return o
}
