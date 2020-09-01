package x

import (
	"github.com/gorilla/sessions"

	"github.com/ory/herodot"
	"github.com/ory/x/logrusx"
)

type LoggingProvider interface {
	Logger() *logrusx.Logger
	Audit() *logrusx.Logger
}

type WriterProvider interface {
	Writer() herodot.Writer
}

type CookieProvider interface {
	CookieManager() sessions.Store
	ContinuityCookieManager() sessions.Store
}
