package driver

import (
	"github.com/ory/x/dbal"

	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/selfservice/flow/profile"
	"github.com/ory/kratos/selfservice/flow/registration"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/errorx"
	"github.com/ory/kratos/selfservice/persistence"
	"github.com/ory/kratos/session"
)

var _ Registry = new(RegistryMemory)

func init() {
	dbal.RegisterDriver(NewRegistryMemory())
}

type RegistryMemory struct {
	*RegistryAbstract

	errorManager              errorx.Manager
	identityPool              identity.Pool
	sessionManager            session.Manager
	selfserviceRequestManager persistence.RequestPersister
}

func NewRegistryMemory() *RegistryMemory {
	r := &RegistryMemory{RegistryAbstract: new(RegistryAbstract)}
	r.RegistryAbstract.with(r)
	return r
}

func (m *RegistryMemory) Init() error {
	return nil
}

func (m *RegistryMemory) IdentityPool() identity.Pool {
	if m.identityPool == nil {
		m.identityPool = identity.NewPoolMemory(m.c, m)
	}
	return m.identityPool
}

func (m *RegistryMemory) CanHandle(dsn string) bool {
	return dsn == "memory"
}

func (m *RegistryMemory) Ping() error {
	return nil
}

func (m *RegistryMemory) SessionManager() session.Manager {
	if m.sessionManager == nil {
		m.sessionManager = session.NewManagerMemory(m.c, m)
	}
	return m.sessionManager
}

func (m *RegistryMemory) getSelfserviceRequestManager() persistence.RequestPersister {
	if m.selfserviceRequestManager == nil {
		m.selfserviceRequestManager = persistence.NewRequestManagerMemory()
	}
	return m.selfserviceRequestManager
}

func (m *RegistryMemory) RegistrationRequestPersister() registration.RequestPersister {
	return m.getSelfserviceRequestManager()
}

func (m *RegistryMemory) LoginRequestPersister() login.RequestPersister {
	return m.getSelfserviceRequestManager()
}

func (m *RegistryMemory) ProfileRequestPersister() profile.RequestPersister {
	return m.getSelfserviceRequestManager()
}

func (m *RegistryMemory) ErrorManager() errorx.Manager {
	if m.errorManager == nil {
		m.errorManager = errorx.NewManagerMemory(m, m.c)
	}
	return m.errorManager
}
