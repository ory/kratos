package saml

import (
	"github.com/ory/herodot"
	"github.com/pkg/errors"
)

type Configuration struct {
	// ID is the provider's ID
	ID string `json:"id"`

	// Provider is "generic" for SAML provider
	Provider string `json:"provider"`

	// Label represents an optional label which can be used in the UI generation.
	Label string `json:"label"`

	// Represent the path of the certificate of your application
	PublicCertPath string `json:"public_cert_path"`

	// Represent the path of the private key of your application
	PrivateKeyPath string `json:"private_key_path"`

	// It is a map where you have to name the attributes contained in the SAML response to associate them with their value
	AttributesMap map[string]string `json:"attributes_map"`

	// Information about the IDP like the sso url, slo url, entiy ID, metadata url
	IDPInformation map[string]string `json:"idp_information"`

	// Mapper specifies the JSONNet code snippet
	// It can be either a URL (file://, http(s)://, base64://) or an inline JSONNet code snippet.
	Mapper string `json:"mapper_url"`
}

type ConfigurationCollection struct {
	SAMLProviders []Configuration `json:"providers"`
}

func (c ConfigurationCollection) Provider(id string, reg registrationStrategyDependencies) (Provider, error) {
	for k := range c.SAMLProviders {
		p := c.SAMLProviders[k]
		if p.ID == id {
			var providerNames []string
			var addProviderName = func(pn string) string {
				providerNames = append(providerNames, pn)
				return pn
			}

			switch p.Provider {
			case addProviderName("generic"):
				return NewProviderSAML(&p, reg), nil
			}
			return nil, errors.Errorf("provider type %s is not supported, supported are: %v", p.Provider, providerNames)
		}
	}
	return nil, errors.WithStack(herodot.ErrNotFound.WithReasonf(`SAML Provider "%s" is unknown or has not been configured`, id))
}

func (c ConfigurationCollection) ProviderConfig(id string) (*Configuration, error) {
	for k := range c.SAMLProviders {
		p := c.SAMLProviders[k]
		if p.ID == id {
			return &p, nil
		}
	}
	return nil, errors.WithStack(herodot.ErrNotFound.WithReasonf(`SAML Provider "%s" is unknown or has not been configured`, id))
}
