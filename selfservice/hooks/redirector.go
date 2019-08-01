package hooks

import (
	"net/http"
	"net/url"

	"github.com/ory/herodot"

	"github.com/ory/hive/selfservice"
	"github.com/ory/hive/session"
	"github.com/ory/hive/x"
)

var _ selfservice.HookLoginPostExecutor = new(Redirector)
var _ selfservice.HookRegistrationPostExecutor = new(Redirector)

type Redirector struct {
	returnTo         func() *url.URL
	whitelist        func() []url.URL
	allowUserDefined func() bool
}

func NewRedirector(
	returnTo func() *url.URL,
	whitelist func() []url.URL,
	allowUserDefined func() bool,
) *Redirector {
	return &Redirector{
		returnTo:         returnTo,
		whitelist:        whitelist,
		allowUserDefined: allowUserDefined,
	}
}

func (e *Redirector) ExecuteRegistrationPostHook(w http.ResponseWriter, r *http.Request, sr *selfservice.RegistrationRequest, _ *session.Session) error {
	return e.do(w, r, sr.RequestURL)
}

func (e *Redirector) ExecuteLoginPostHook(w http.ResponseWriter, r *http.Request, sr *selfservice.LoginRequest, _ *session.Session) error {
	return e.do(w, r, sr.RequestURL)
}

func (e *Redirector) do(w http.ResponseWriter, r *http.Request, originalURL string) error {
	ou, err := url.ParseRequestURI(originalURL)
	if err != nil {
		return herodot.ErrInternalServerError.WithReasonf("The redirect hook was unable to parse the original request URL: %s", err)
	}

	returnTo := e.returnTo().String()
	if e.allowUserDefined() {
		var err error
		returnTo, err = x.DetermineReturnToURL(ou, e.returnTo(), e.whitelist())
		if err != nil {
			return err
		}
	}

	http.Redirect(w, r, returnTo, http.StatusFound)
	return nil
}
