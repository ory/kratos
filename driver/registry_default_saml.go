package driver

import "github.com/ory/kratos/selfservice/strategy/saml"

func (m *RegistryDefault) SAMLHandler() *saml.Handler {
	if m.selfserviceSAMLHandler == nil {
		m.selfserviceSAMLHandler = saml.NewHandler(m)
	}

	return m.selfserviceSAMLHandler
}
