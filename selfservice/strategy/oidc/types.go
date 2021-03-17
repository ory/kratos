package oidc

import (
	"bytes"
	"encoding/json"

	"github.com/ory/kratos/ui/container"

	"github.com/ory/kratos/ui/node"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"

	"github.com/ory/kratos/identity"

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
	*container.Container
}

func AddProviders(c *container.Container, providers []Configuration) {
	for _, p := range providers {
		c.GetNodes().Append(node.NewInputField(identity.CredentialsTypeOIDC.String()+".provider", p.ID, node.OpenIDConnectGroup, node.InputAttributeTypeSubmit))
	}
}

func NewFlowMethod(f *container.Container) *FlowMethod {
	return &FlowMethod{Container: f}
}

type ider interface {
	GetID() uuid.UUID
}
