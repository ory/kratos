// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package hook

import (
	"net/http"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/flow/registration"
	"github.com/ory/kratos/selfservice/strategy/code"
)

type (
	codeAddressDependencies interface {
		code.RegistrationCodePersistenceProvider
	}
	CodeAddressVerifier struct {
		r codeAddressDependencies
	}
)

var _ registration.PostHookPrePersistExecutor = new(CodeAddressVerifier)

func NewCodeAddressVerifier(r codeAddressDependencies) *CodeAddressVerifier {
	return &CodeAddressVerifier{r: r}
}

func (cv *CodeAddressVerifier) ExecutePostRegistrationPrePersistHook(w http.ResponseWriter, r *http.Request, a *registration.Flow, i *identity.Identity) error {
	if a.Active != identity.CredentialsTypeCodeAuth {
		return nil
	}

	recoveryCode, err := cv.r.RegistrationCodePersister().GetUsedRegistrationCode(r.Context(), a.GetID())
	if err != nil {
		return err
	}

	if recoveryCode == nil {
		return nil
	}

	for idx := range i.VerifiableAddresses {
		va := &i.VerifiableAddresses[idx]
		if !va.Verified && recoveryCode.Address == va.Value {
			va.Verified = true
			va.Status = identity.VerifiableAddressStatusCompleted
			break
		}
	}

	return nil
}
