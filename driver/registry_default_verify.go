package driver

import (
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/flow/verify"
)

func (m *RegistryDefault) VerificationPersister() verify.Persister {
	return m.persister
}

func (m *RegistryDefault) VerificationRequestErrorHandler() *verify.ErrorHandler {
	if m.selfserviceVerifyErrorHandler == nil {
		m.selfserviceVerifyErrorHandler = verify.NewErrorHandler(m, m.c)
	}

	return m.selfserviceVerifyErrorHandler
}

func (m *RegistryDefault) VerificationManager() *identity.Manager {
	if m.selfserviceVerifyManager == nil {
		m.selfserviceVerifyManager = identity.NewManager(m, m.c)
	}

	return m.selfserviceVerifyManager
}

func (m *RegistryDefault) VerificationHandler() *verify.Handler {
	if m.selfserviceVerifyHandler == nil {
		m.selfserviceVerifyHandler = verify.NewHandler(m, m.c)
	}

	return m.selfserviceVerifyHandler
}

func (m *RegistryDefault) VerificationSender() *verify.Sender {
	if m.selfserviceVerifySender == nil {
		m.selfserviceVerifySender = verify.NewSender(m, m.c)
	}

	return m.selfserviceVerifySender
}
