package oidc

import (
	"github.com/gofrs/uuid"

	"github.com/ory/kratos/selfservice/form"
)

// swagger:model oidcStrategyCredentialsConfig
type CredentialsConfig struct {
	Subject  string `json:"subject"`
	Provider string `json:"provider"`
}

// swagger:model oidcRequestMethodConfig
type RequestMethod struct {
	*form.HTMLForm
	Providers []form.Field `json:"providers"`
}

func (r *RequestMethod) AddProviders(providers []Configuration) *RequestMethod {
	for _, p := range providers {
		r.Providers = append(r.Providers, form.Field{Name: "provider", Type: "submit", Value: p.ID})
	}
	return r
}

func NewRequestMethodConfig(f *form.HTMLForm) *RequestMethod {
	return &RequestMethod{HTMLForm: f}
}

type request interface {
	GetID() uuid.UUID
}
