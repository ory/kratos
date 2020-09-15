package oidc

import (
	"bytes"
	"encoding/json"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/form"
	"github.com/ory/kratos/x"
)

type CredentialsConfig struct {
	Providers []ProviderCredentialsConfig `json:"providers"`
}

func NewCredentials(provider, subject string) (*identity.Credentials, error) {
	var b bytes.Buffer
	if err := json.NewEncoder(&b).Encode(CredentialsConfig{
		Providers: []ProviderCredentialsConfig{{Subject: subject, Provider: provider}},
	}); err != nil {
		return nil, errors.WithStack(x.PseudoPanic.
			WithDebugf("Unable to encode password options to JSON: %s", err))
	}

	return &identity.Credentials{
		Type:        identity.CredentialsTypeOIDC,
		Identifiers: []string{uid(provider, subject)},
		Config:      b.Bytes(),
	}, nil
}

type ProviderCredentialsConfig struct {
	Subject  string `json:"subject"`
	Provider string `json:"provider"`
}

type FlowMethod struct {
	*form.HTMLForm
}

func (r *FlowMethod) AddProviders(providers []Configuration) *FlowMethod {
	for _, p := range providers {
		r.Fields = append(r.Fields, form.Field{Name: "provider", Type: "submit", Value: p.ID})
	}
	return r
}

func NewFlowMethod(f *form.HTMLForm) *FlowMethod {
	return &FlowMethod{HTMLForm: f}
}

type ider interface {
	GetID() uuid.UUID
}
