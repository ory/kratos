package x

import (
	"github.com/gorilla/sessions"
	"github.com/sirupsen/logrus"

	"github.com/ory/herodot"
)

type LoggingProvider interface {
	Logger() logrus.FieldLogger
}

type WriterProvider interface {
	Writer() herodot.Writer
}

type CookieProvider interface {
	CookieManager() sessions.Store
}
