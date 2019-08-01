package oidc

import (
	"net/url"
)

type ProviderGoogle struct {
	*ProviderGenericOIDC
}

func NewProviderGoogle(
	config *Configuration,
	public *url.URL,
) *ProviderGoogle {
	config.IssuerURL = "https://accounts.google.com"
	return &ProviderGoogle{
		ProviderGenericOIDC: &ProviderGenericOIDC{
			config: config,
			public: public,
		},
	}
}
