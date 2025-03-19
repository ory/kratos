// Copyright © 2025 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package registration

import (
	"net/http"
)

type UnifiedFormHydrator interface {
	PopulateRegistrationMethod(r *http.Request, sr *Flow) error
}

type FormHydratorOptions struct {
}

type FormHydratorModifier func(o *FormHydratorOptions)

type FormHydrator interface {
	UnifiedFormHydrator
	PopulateRegistrationMethodCredentials(r *http.Request, sr *Flow, options ...FormHydratorModifier) error
	PopulateRegistrationMethodProfile(r *http.Request, sr *Flow) error
}
