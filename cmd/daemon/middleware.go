package daemon

import (
	"net/http"
	"time"

	"github.com/ory/x/healthx"
	"github.com/ory/x/reqlog"
	"github.com/sirupsen/logrus"
	"github.com/urfave/negroni"
)

func NewNegroniLoggerMiddleware(l logrus.FieldLogger, name string) *reqlog.Middleware {
	n := reqlog.NewMiddlewareFromLogger(l.(*logrus.Logger), name).ExcludePaths(healthx.AliveCheckPath, healthx.ReadyCheckPath)
	n.Before = func(entry *logrus.Entry, req *http.Request, remoteAddr string) *logrus.Entry {
		return entry.WithFields(logrus.Fields{
			"name":    name,
			"request": req.RequestURI,
			"method":  req.Method,
			"remote":  remoteAddr,
		})
	}

	n.After = func(entry *logrus.Entry, res negroni.ResponseWriter, latency time.Duration, name string) *logrus.Entry {
		return entry.WithFields(logrus.Fields{
			"name":        name,
			"status":      res.Status(),
			"text_status": http.StatusText(res.Status()),
			"took":        latency,
		})
	}
	return n
}
