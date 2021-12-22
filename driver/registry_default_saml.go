package driver

import (
	"github.com/ory/kratos/selfservice/flow/saml"
)

func (m *RegistryDefault) SamlHandler() *saml.Handler {
	if m.selfserviceSamlHandler == nil {
		m.selfserviceSamlHandler = saml.NewHandler(m)
	}

	return m.selfserviceSamlHandler
}
