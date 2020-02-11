package driver

import (
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

func (m *RegistryDefault) VerificationManager() *verify.Manager {
	if m.selfserviceVerifyManager == nil {
		m.selfserviceVerifyManager = verify.NewManager(m, m.c)
	}

	return m.selfserviceVerifyManager
}

func (m *RegistryDefault) VerificationHandler() *verify.Handler {
	if m.selfserviceVerifyHandler == nil {
		m.selfserviceVerifyHandler = verify.NewHandler(m, m.c)
	}

	return m.selfserviceVerifyHandler
}
