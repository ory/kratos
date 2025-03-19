// Copyright © 2025 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package registration

import (
	"encoding/json"
	"net/http"
)

type UnifiedFormHydrator interface {
	PopulateRegistrationMethod(r *http.Request, sr *Flow) error
}

type FormHydratorOptions struct {
	WithTraits json.RawMessage
}

type FormHydratorModifier func(o *FormHydratorOptions)

func WithTraits(traits json.RawMessage) FormHydratorModifier {
	return func(o *FormHydratorOptions) {
		o.WithTraits = traits
	}
}

type FormHydrator interface {
	UnifiedFormHydrator
	PopulateRegistrationMethodCredentials(r *http.Request, sr *Flow, options ...FormHydratorModifier) error
	PopulateRegistrationMethodProfile(r *http.Request, sr *Flow, options ...FormHydratorModifier) error
}
