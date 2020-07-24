package driver

import (
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/flow/verification"
)

func (m *RegistryDefault) VerificationPersister() verification.Persister {
	return m.persister
}

func (m *RegistryDefault) VerificationRequestErrorHandler() *verification.ErrorHandler {
	if m.selfserviceVerifyErrorHandler == nil {
		m.selfserviceVerifyErrorHandler = verification.NewErrorHandler(m, m.c)
	}

	return m.selfserviceVerifyErrorHandler
}

func (m *RegistryDefault) VerificationManager() *identity.Manager {
	if m.selfserviceVerifyManager == nil {
		m.selfserviceVerifyManager = identity.NewManager(m, m.c)
	}

	return m.selfserviceVerifyManager
}

func (m *RegistryDefault) VerificationHandler() *verification.Handler {
	if m.selfserviceVerifyHandler == nil {
		m.selfserviceVerifyHandler = verification.NewHandler(m, m.c)
	}

	return m.selfserviceVerifyHandler
}

func (m *RegistryDefault) VerificationSender() *verification.Sender {
	if m.selfserviceVerifySender == nil {
		m.selfserviceVerifySender = verification.NewSender(m, m.c)
	}

	return m.selfserviceVerifySender
}
