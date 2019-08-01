package daemon

import (
	"net/http"
	"time"

	negronilogrus "github.com/meatballhat/negroni-logrus"
	"github.com/sirupsen/logrus"
	"github.com/urfave/negroni"
)

func NewNegroniLoggerMiddleware(l logrus.FieldLogger, name string) *negronilogrus.Middleware {
	n := negronilogrus.NewMiddlewareFromLogger(l.(*logrus.Logger), name)
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
