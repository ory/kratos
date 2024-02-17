// Copyright © 2024 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oid2

import (
	"github.com/ory/herodot"
	"github.com/pkg/errors"
	"golang.org/x/exp/maps"
)

type Configuration struct {
	// ID is the provider's ID
	ID string `json:"id"`

	// Provider is either "generic" for a generic OpenID 2.0 Provider or one of:
	// - generic
	// - steam
	Provider string `json:"provider"`

	// Label represents an optional label which can be used in the UI generation.
	Label string `json:"label"`

	// DiscoveryUrl is the URL of the Open ID 2.0 discovery document, typically something like:
	// https://example.org/openid. Should only be used and when `provider` is set to `generic`.
	DiscoveryUrl string `json:"discovery_url"`
}

type ConfigurationCollection struct {
	BaseRedirectURI string          `json:"base_redirect_uri"`
	Providers       []Configuration `json:"providers"`
}

var supportedProviders = map[string]func(config *Configuration, reg Dependencies) Provider{
	"generic": NewProviderGenericOid2,
	"steam":   NewProviderSteam,
}

func (c ConfigurationCollection) Provider(id string, reg Dependencies) (Provider, error) {
	for k := range c.Providers {
		p := c.Providers[k]
		if p.ID == id {
			if f, ok := supportedProviders[p.Provider]; ok {
				return f(&p, reg), nil
			}

			return nil, errors.Errorf("provider type %s is not supported, supported are: %v", p.Provider, maps.Keys(supportedProviders))
		}
	}
	return nil, errors.WithStack(herodot.ErrNotFound.WithReasonf(`OpenID 2.0 Provider "%s" is unknown or has not been configured`, id))
}
