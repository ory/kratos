package session

import (
	"context"
	"net/http"
	"net/url"

	"github.com/ory/kratos/text"

	"github.com/gofrs/uuid"

	"github.com/ory/herodot"
)

// ErrNoActiveSessionFound is returned when no active cookie session could be found in the request.
type ErrNoActiveSessionFound struct {
	*herodot.DefaultError `json:"error"`
}

// NewErrNoActiveSessionFound creates a new ErrNoActiveSessionFound
func NewErrNoActiveSessionFound() *ErrNoActiveSessionFound {
	return &ErrNoActiveSessionFound{
		DefaultError: herodot.ErrUnauthorized.WithID(text.ErrNoActiveSession).WithError("request does not have a valid authentication session").WithReason("No active session was found in this request."),
	}
}

func (e *ErrNoActiveSessionFound) EnhanceJSONError() interface{} {
	return e
}

// ErrAALNotSatisfied is returned when an active session was found but the requested AAL is not satisfied.
//
// swagger:model errorAuthenticatorAssuranceLevelNotSatisfied
type ErrAALNotSatisfied struct {
	*herodot.DefaultError `json:"error"`
	RedirectTo            string `json:"redirect_browser_to"`
}

func (e *ErrAALNotSatisfied) EnhanceJSONError() interface{} {
	return e
}

func (e *ErrAALNotSatisfied) PassReturnToParameter(requestURL string) error {
	req, err := url.Parse(requestURL)
	if err != nil {
		return err
	}
	returnTo := req.Query().Get("return_to")
	if len(returnTo) == 0 {
		return nil
	}

	u, err := url.Parse(e.RedirectTo)
	if err != nil {
		return err
	}
	q := u.Query()
	q.Set("return_to", returnTo)
	u.RawQuery = q.Encode()
	e.RedirectTo = u.String()

	return nil
}

// NewErrAALNotSatisfied creates a new ErrAALNotSatisfied.
func NewErrAALNotSatisfied(redirectTo string) *ErrAALNotSatisfied {
	return &ErrAALNotSatisfied{
		RedirectTo: redirectTo,
		DefaultError: &herodot.DefaultError{
			IDField:     text.ErrIDHigherAALRequired,
			StatusField: http.StatusText(http.StatusForbidden),
			ErrorField:  "Session does not fulfill the requested Authenticator Assurance Level",
			ReasonField: "An active session was found but it does not fulfill the requested Authenticator Assurance Level. Please verify yourself with a second factor to resolve this issue.",
			CodeField:   http.StatusForbidden,
			DetailsField: map[string]interface{}{
				"redirect_browser_to": redirectTo,
			},
		},
	}
}

// Manager handles identity sessions.
type Manager interface {
	// UpsertAndIssueCookie stores a session in the database and issues a cookie by calling IssueCookie.
	//
	// Also regenerates CSRF tokens due to assumed principal change.
	UpsertAndIssueCookie(context.Context, http.ResponseWriter, *http.Request, *Session) error

	// IssueCookie issues a cookie for the given session.
	//
	// Also regenerates CSRF tokens due to assumed principal change.
	IssueCookie(context.Context, http.ResponseWriter, *http.Request, *Session) error

	// Checks if the request uses an outdated cookie and re-issues the cookie if needed
	ReIssueRefreshedCookie(context.Context, http.ResponseWriter, *http.Request, *Session) error

	// FetchFromRequest creates an HTTP session using cookies.
	FetchFromRequest(context.Context, *http.Request) (*Session, error)

	// PurgeFromRequest removes an HTTP session.
	PurgeFromRequest(context.Context, http.ResponseWriter, *http.Request) error

	// DoesSessionSatisfy answers if a session is satisfying the AAL.
	DoesSessionSatisfy(r *http.Request, sess *Session, requestedAAL string) error

	// SessionAddAuthenticationMethods adds one or more authentication method to the session.
	SessionAddAuthenticationMethods(ctx context.Context, sid uuid.UUID, methods ...AuthenticationMethod) error
}

type ManagementProvider interface {
	SessionManager() Manager
}
