// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package hook

import (
	"net/http"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/flow/registration"
	"github.com/ory/kratos/selfservice/flow/verification"
	"github.com/ory/kratos/selfservice/strategy/code"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/x"
)

type (
	codeAddressDependencies interface {
		config.Provider
		x.CSRFTokenGeneratorProvider
		x.CSRFProvider
		verification.StrategyProvider
		verification.FlowPersistenceProvider
		code.RegistrationCodePersistenceProvider
		identity.PrivilegedPoolProvider
		x.WriterProvider
	}
	CodeAddressVerifier struct {
		r codeAddressDependencies
	}
)

var (
	_ registration.PostHookPostPersistExecutor = new(CodeAddressVerifier)
)

func NewCodeAddressVerifier(r codeAddressDependencies) *CodeAddressVerifier {
	return &CodeAddressVerifier{r: r}
}

func (cv *CodeAddressVerifier) ExecutePostRegistrationPostPersistHook(w http.ResponseWriter, r *http.Request, a *registration.Flow, s *session.Session) error {
	if a.Active != identity.CredentialsTypeCodeAuth {
		return nil
	}

	recoveryCode, err := cv.r.RegistrationCodePersister().GetUsedRegistrationCode(r.Context(), a.GetID())
	if err != nil {
		return err
	}

	for idx := range s.Identity.VerifiableAddresses {
		va := s.Identity.VerifiableAddresses[idx]
		if !va.Verified && recoveryCode.Address == va.Value {
			va.Verified = true
			va.Status = identity.VerifiableAddressStatusCompleted
			if err := cv.r.PrivilegedIdentityPool().UpdateVerifiableAddress(r.Context(), &va); err != nil {
				return err
			}
			break
		}
	}

	return nil
}
