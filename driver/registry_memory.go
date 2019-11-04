package driver

import (
	"github.com/ory/x/dbal"

	"github.com/ory/hive/errorx"
	"github.com/ory/hive/identity"
	"github.com/ory/hive/selfservice"
	"github.com/ory/hive/session"
)

var _ Registry = new(RegistryMemory)

func init() {
	dbal.RegisterDriver(NewRegistryMemory())
}

type RegistryMemory struct {
	*RegistryAbstract

	identityPool              identity.Pool
	sessionManager            session.Manager
	selfserviceRequestManager selfservice.RequestManager
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
func (m *RegistryMemory) RegistrationRequestManager() selfservice.RegistrationRequestManager {
	if m.selfserviceRequestManager == nil {
		m.selfserviceRequestManager = selfservice.NewRequestManagerMemory()
	}
	return m.selfserviceRequestManager
}
func (m *RegistryMemory) LoginRequestManager() selfservice.LoginRequestManager {
	if m.selfserviceRequestManager == nil {
		m.selfserviceRequestManager = selfservice.NewRequestManagerMemory()
	}
	return m.selfserviceRequestManager
}

func (m *RegistryMemory) ErrorManager() errorx.Manager {
	if m.errorManager == nil {
		m.errorManager = errorx.NewManagerMemory(m, m.c)
	}
	return m.errorManager
}
