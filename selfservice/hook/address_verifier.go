// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package hook

import (
	"net/http"

	"github.com/ory/kratos/x"

	"github.com/pkg/errors"

	"github.com/ory/kratos/ui/node"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/session"
)

var _ login.PostHookExecutor = new(AddressVerifier)

type AddressVerifier struct{}

func NewAddressVerifier() *AddressVerifier {
	return &AddressVerifier{}
}

func (e *AddressVerifier) ExecuteLoginPostHook(_ http.ResponseWriter, _ *http.Request, _ node.UiNodeGroup, f *login.Flow, s *session.Session) error {
	// if the login happens using the password method, there must be at least one verified address
	if f.Active != identity.CredentialsTypePassword {
		return nil
	}

	if len(s.Identity.VerifiableAddresses) == 0 {
		return errors.WithStack(x.ErrMisconfiguration.WithReason("Expected to find a verification address but this identity does not have one assigned. This indicates an error with the identity schema."))
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
