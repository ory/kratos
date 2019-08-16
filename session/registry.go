package session

import (
	"time"

	"github.com/gorilla/sessions"

	"github.com/ory/hive/identity"
)

type Registry interface {
	CookieManager() sessions.Store
	SessionManager() Manager
	identity.PoolProvider
}

type Configuration interface {
	SessionLifespan() time.Duration
	SessionSecrets() [][]byte
}
