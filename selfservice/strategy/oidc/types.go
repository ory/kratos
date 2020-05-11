package oidc

import (
	"github.com/gofrs/uuid"

	"github.com/ory/kratos/selfservice/form"
)

// swagger:model oidcStrategyCredentialsConfig
type CredentialsConfig struct {
	Providers []ProviderCredentialsConfig `json:"providers"`
}

type ProviderCredentialsConfig struct {
	Subject  string `json:"subject"`
	Provider string `json:"provider"`
}

// swagger:model oidcRequestMethodConfig
type RequestMethod struct {
	*form.HTMLForm
}

func (r *RequestMethod) AddProviders(providers []Configuration) *RequestMethod {
	for _, p := range providers {
		r.Fields = append(r.Fields, form.Field{Name: "provider", Type: "submit", Value: p.ID})
	}
	return r
}

func NewRequestMethodConfig(f *form.HTMLForm) *RequestMethod {
	return &RequestMethod{HTMLForm: f}
}

type request interface {
	GetID() uuid.UUID
	IsForced() bool
}
