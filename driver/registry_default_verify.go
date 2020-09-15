package driver

import (
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/flow/verification"
	"github.com/ory/kratos/selfservice/strategy/link"
)

func (m *RegistryDefault) VerificationFlowPersister() verification.FlowPersister {
	return m.persister
}

func (m *RegistryDefault) VerificationFlowErrorHandler() *verification.ErrorHandler {
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

func (m *RegistryDefault) LinkSender() *link.Sender {
	if m.selfserviceLinkSender == nil {
		m.selfserviceLinkSender = link.NewSender(m, m.c)
	}

	return m.selfserviceLinkSender
}
