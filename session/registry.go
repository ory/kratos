package session

import (
	"time"

	"github.com/gorilla/sessions"
)

type Registry interface {
	CookieManager() sessions.Store
	SessionManager() Manager
}

type Configuration interface {
	SessionLifespan() time.Duration
	SessionSecrets() [][]byte
}
