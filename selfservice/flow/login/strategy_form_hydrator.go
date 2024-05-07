package login

import (
	"github.com/ory/kratos/identity"
	"net/http"
)

type LegacyFormHydrator interface {
	PopulateLoginMethod(r *http.Request, requestedAAL identity.AuthenticatorAssuranceLevel, sr *Flow) error
}

type FormHydrator interface {
	PopulateLoginMethodRefresh(r *http.Request, sr *Flow) error
	PopulateLoginMethodFirstFactor(r *http.Request, sr *Flow) error
	PopulateLoginMethodSecondFactor(r *http.Request, sr *Flow) error
	PopulateLoginMethodMultiStepSelection(r *http.Request, sr *Flow, options ...FormHydratorModifier) error
	PopulateLoginMethodMultiStepIdentification(r *http.Request, sr *Flow) error
}

type FormHydratorOptions struct {
	IdentityHint *identity.Identity
}

type FormHydratorModifier func(o *FormHydratorOptions)

func WithIdentityHint(i *identity.Identity) FormHydratorModifier {
	return func(o *FormHydratorOptions) {
		o.IdentityHint = i
	}
}

func NewFormHydratorOptions(modifiers []FormHydratorModifier) *FormHydratorOptions {
	o := new(FormHydratorOptions)
	for _, m := range modifiers {
		m(o)
	}
	return o
}
