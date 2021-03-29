package daemon

import (
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/urfave/negroni"

	"github.com/ory/x/healthx"
	"github.com/ory/x/logrusx"
	"github.com/ory/x/reqlog"

	"github.com/ory/kratos/driver/config"
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

func NewStripPrefixMiddleware(prefixLoader func(r *http.Request) string) negroni.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		prefix := strings.TrimRight(prefixLoader(r), "/")
		if len(prefix) == 0 {
			next.ServeHTTP(w, r)
			return
		}
		http.StripPrefix(prefix, next).ServeHTTP(w, r)
	}
}

func extractPrefixFromBaseURL(u *url.URL) string {
	if u == nil {
		return ""
	}
	return strings.TrimSuffix(u.Path, "/")
}

func publicURLPrefixExtractor(provider config.Provider) func(r *http.Request) string {
	return func(r *http.Request) string {
		return extractPrefixFromBaseURL(provider.Config(r.Context()).SelfPublicURL(r))
	}
}

func adminURLPrefixExtractor(provider config.Provider) func(r *http.Request) string {
	return func(r *http.Request) string {
		return extractPrefixFromBaseURL(provider.Config(r.Context()).SelfAdminURL())
	}
}
