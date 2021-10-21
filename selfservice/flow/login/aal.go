package login

import (
	"github.com/pkg/errors"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/flow"
)

func CheckAAL(f *Flow, expected identity.AuthenticatorAssuranceLevel) error {
	if f.RequestedAAL != expected {
		return errors.WithStack(flow.ErrStrategyNotResponsible)
	}
	return nil
}
