package daemon

import (
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/urfave/negroni"

	"github.com/ory/x/logrusx"

	"github.com/ory/x/healthx"
	"github.com/ory/x/reqlog"
)

func NewNegroniLoggerMiddleware(l *logrusx.Logger, name string) *reqlog.Middleware {
	n := reqlog.NewMiddlewareFromLogger(l, name).ExcludePaths(healthx.AliveCheckPath, healthx.ReadyCheckPath)
	n.Before = func(entry *logrusx.Logger, req *http.Request, remoteAddr string) *logrusx.Logger {
		return entry.WithFields(logrus.Fields{
			"name":    name,
			"request": req.RequestURI,
			"method":  req.Method,
			"remote":  remoteAddr,
		})
	}

	n.After = func(entry *logrusx.Logger, req *http.Request, res negroni.ResponseWriter, latency time.Duration, name string) *logrusx.Logger {
		return entry.WithFields(logrus.Fields{
			"name":        name,
			"status":      res.Status(),
			"text_status": http.StatusText(res.Status()),
			"took":        latency,
		})
	}
	return n
}
