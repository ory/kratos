// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package hook

import (
	"github.com/ory/kratos/driver/config"
	"net/http"

	"github.com/pkg/errors"

	"github.com/ory/kratos/ui/node"

	"github.com/ory/herodot"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/session"
)

var _ login.PostHookExecutor = new(AddressVerifier)

type addressVerifierDependencies interface {
	config.Provider
}

type AddressVerifier struct {
	d addressVerifierDependencies
}

func NewAddressVerifier() *AddressVerifier {
	return &AddressVerifier{}
}

func (e *AddressVerifier) ExecuteLoginPostHook(w http.ResponseWriter, r *http.Request, g node.UiNodeGroup, f *login.Flow, s *session.Session) error {
	if e.d.Config().LegacyRequireVerifiedAddressError(r.Context()) {
		return e.legacyExecuteLoginPostHook(w, r, g, f, s)
	}

	return e.executeLoginPostHook(w, r, g, f, s)
}

func (e *AddressVerifier) legacyExecuteLoginPostHook(_ http.ResponseWriter, _ *http.Request, _ node.UiNodeGroup, f *login.Flow, s *session.Session) error {
	// if the login happens using the password method, there must be at least one verified address
	if f.Active != identity.CredentialsTypePassword {
		return nil
	}

	// TODO: can this happen at all?
	if len(s.Identity.VerifiableAddresses) == 0 {
		return errors.WithStack(herodot.ErrMisconfiguration.WithReason("A misconfiguration prevents login. Expected to find a verification address but this identity does not have one assigned."))
	}

	addressVerified := false
	for _, va := range s.Identity.VerifiableAddresses {
		if va.Verified {
			addressVerified = true
			break
		}
	}

	if !addressVerified {
		return login.ErrAddressNotVerified
	}

	return nil
}

func (e *AddressVerifier) executeLoginPostHook(_ http.ResponseWriter, r *http.Request, _ node.UiNodeGroup, f *login.Flow, s *session.Session) error {
	// The verification hook does not trigger for the code method, as the code method handles verification itself.
	if f.Active == identity.CredentialsTypeCodeAuth {
		return nil
	}

	if len(s.Identity.VerifiableAddresses) == 0 {
		return errors.WithStack(herodot.ErrMisconfiguration.WithReason("Expected to find a verification address but this identity does not have one assigned."))
	}

	// We require at least one verified address.
	for _, va := range s.Identity.VerifiableAddresses {
		if va.Verified {
			return nil
		}
	}

	// No address was found, create a verification flow and add it to the continue with.

	return nil
}
