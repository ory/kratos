// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oidc

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestClaimsValidate(t *testing.T) {
	require.Error(t, new(Claims).Validate())
	require.Error(t, (&Claims{Issuer: "not-empty"}).Validate())
	require.Error(t, (&Claims{Issuer: "not-empty"}).Validate())
	require.Error(t, (&Claims{Subject: "not-empty"}).Validate())
	require.Error(t, (&Claims{Subject: "not-empty"}).Validate())
	require.NoError(t, (&Claims{Issuer: "not-empty", Subject: "not-empty"}).Validate())
}

type TestProvider struct {
	*ProviderGenericOIDC
}

func NewTestProvider(c *Configuration, reg Dependencies) Provider {
	return &TestProvider{
		ProviderGenericOIDC: NewProviderGenericOIDC(c, reg).(*ProviderGenericOIDC),
	}
}

func RegisterTestProvider(id string) func() {
	supportedProviders[id] = func(c *Configuration, reg Dependencies) Provider {
		return NewTestProvider(c, reg)
	}
	return func() {
		delete(supportedProviders, id)
	}
}

var _ IDTokenVerifier = new(TestProvider)

func (t *TestProvider) Verify(ctx context.Context, token string) (*Claims, error) {
	if token == "error" {
		return nil, fmt.Errorf("stub error")
	}
	c := Claims{}
	if err := json.Unmarshal([]byte(token), &c); err != nil {
		return nil, err
	}
	return &c, nil
}
