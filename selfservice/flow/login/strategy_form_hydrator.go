// Copyright © 2024 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package login

import (
	"net/http"

	"github.com/pkg/errors"

	"github.com/ory/kratos/identity"
)

type UnifiedFormHydrator interface {
	PopulateLoginMethod(r *http.Request, requestedAAL identity.AuthenticatorAssuranceLevel, sr *Flow) error
}

type FormHydrator interface {
	PopulateLoginMethodFirstFactorRefresh(r *http.Request, sr *Flow) error
	PopulateLoginMethodFirstFactor(r *http.Request, sr *Flow) error
	PopulateLoginMethodSecondFactor(r *http.Request, sr *Flow) error
	PopulateLoginMethodSecondFactorRefresh(r *http.Request, sr *Flow) error

	// PopulateLoginMethodIdentifierFirstCredentials populates the login form with the first factor credentials.
	// This method is called when the login flow is set to identifier first. The method will receive information
	// about the identity that is being used to log in and the identifier that was used to find the identity.
	//
	// The method should populate the login form with the credentials of the identity.
	//
	// If the method can not find any credentials (because the identity does not exist) idfirst.ErrNoCredentialsFound
	// must be returned. When returning  idfirst.ErrNoCredentialsFound the strategy will appropriately deal with
	// account enumeration mitigation.
	//
	// This method does however need to take appropriate steps to show/hide certain fields depending on the account
	// enumeration configuration.
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
