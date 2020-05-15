package settings

import (
	"fmt"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"

	"github.com/ory/x/sqlcon"

	"github.com/ory/herodot"

	"github.com/ory/kratos/continuity"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/x"
)

var (
	ErrContinuePreviousAction = errors.New("found prior action")
)

type UpdatePayload interface {
	GetRequestID() uuid.UUID
	SetRequestID(id uuid.UUID)
}

type UpdateContext struct {
	Valid   bool
	Session *session.Session
	Request *Request
}

func PrepareUpdate(d interface {
	x.LoggingProvider
	continuity.ManagementProvider
	session.ManagementProvider
	RequestPersistenceProvider
}, w http.ResponseWriter, r *http.Request, name string, payload UpdatePayload) (*UpdateContext, error) {
	ss, err := d.SessionManager().FetchFromRequest(r.Context(), r)
	if err != nil {
		return new(UpdateContext), err
	}

	rid, err := GetRequestID(r)
	if err != nil {
		return new(UpdateContext), err
	}

	payload.SetRequestID(rid)

	req, err := d.SettingsRequestPersister().GetSettingsRequest(r.Context(), rid)
	if errors.Is(err, sqlcon.ErrNoRows) {
		return new(UpdateContext), errors.WithStack(herodot.ErrNotFound.WithReasonf("The settings request could not be found. Please restart the flow."))
	} else if err != nil {
		return new(UpdateContext), err
	}

	if err := req.Valid(ss); err != nil {
		return new(UpdateContext), err
	}

	c := &UpdateContext{Session: ss, Request: req, Valid: true}
	if _, err := d.ContinuityManager().Continue(
		r.Context(), w, r, name,
		ContinuityOptions(payload, ss.Identity)...); err == nil {
		c.Valid = true
		if payload.GetRequestID() == rid {
			return c, ErrContinuePreviousAction
		}
		d.Logger().
			WithField("package", pkgName).
			WithField("stack_trace", fmt.Sprintf("%s", debug.Stack())).
			WithField("expected_request_id", payload.GetRequestID()).
			WithField("actual_request_id", rid).
			Debug("Request ID from continuity manager does not match Request ID from request.")
		return c, nil
	} else if !errors.Is(err, &continuity.ErrNotResumable) {
		return new(UpdateContext), err
	}

	if err := r.ParseForm(); err != nil {
		return new(UpdateContext), errors.WithStack(err)
	}

	c.Valid = true
	return c, nil
}

func ContinuityOptions(p interface{}, i *identity.Identity) []continuity.ManagerOption {
	return []continuity.ManagerOption{
		continuity.WithPayload(p),
		continuity.WithIdentity(i),
		continuity.WithLifespan(time.Minute * 15),
	}
}

func GetRequestID(r *http.Request) (uuid.UUID, error) {
	rid := x.ParseUUID(r.URL.Query().Get("request"))
	if rid == uuid.Nil {
		return rid, errors.WithStack(herodot.ErrBadRequest.WithReasonf("The request query parameter is missing or malformed."))
	}
	return rid, nil
}
