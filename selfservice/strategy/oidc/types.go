package oidc

import (
	"bytes"
	"encoding/json"

	"github.com/ory/kratos/text"
	"github.com/ory/x/stringsx"

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

func NewCredentials(accessToken, refreshToken, provider, subject string) (*identity.Credentials, error) {
	var b bytes.Buffer
	if err := json.NewEncoder(&b).Encode(CredentialsConfig{
		Providers: []ProviderCredentialsConfig{
			{
				Subject:               subject,
				Provider:              provider,
				EncryptedAccessToken:  accessToken,
				EncryptedRefreshToken: refreshToken,
			}},
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
	Subject               string `json:"subject"`
	Provider              string `json:"provider"`
	EncryptedAccessToken  string `json:"encrypted_access_token"`
	EncryptedRefreshToken string `json:"encrypted_refresh_token"`
}

type FlowMethod struct {
	*container.Container
}

func AddProviders(c *container.Container, providers []Configuration, message func(provider string) *text.Message) {
	for _, p := range providers {
		AddProvider(c, p.ID, message(
			stringsx.Coalesce(p.Label, p.ID)))
	}
}

func AddProvider(c *container.Container, providerID string, message *text.Message) {
	c.GetNodes().Append(
		node.NewInputField("provider", providerID, node.OpenIDConnectGroup, node.InputAttributeTypeSubmit).WithMetaLabel(message),
	)
}

func NewFlowMethod(f *container.Container) *FlowMethod {
	return &FlowMethod{Container: f}
}

type ider interface {
	GetID() uuid.UUID
}
