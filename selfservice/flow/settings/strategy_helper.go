package settings

import (
	"net/http"
	"time"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"

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
	Session *session.Session
	Request *Request
}

func PrepareUpdate(d interface {
	continuity.ManagementProvider
	session.ManagementProvider
	RequestPersistenceProvider
}, r *http.Request, prefix string, payload UpdatePayload) (*UpdateContext, error) {
	ss, err := d.SessionManager().FetchFromRequest(r.Context(), r)
	if err != nil {
		return nil, err
	}

	rid, err := GetRequestID(r)
	if err != nil {
		return nil, err
	}

	req, err := d.SettingsRequestPersister().GetSettingsRequest(r.Context(), rid)
	if err != nil {
		return nil, err
	}

	if err := req.Valid(ss); err != nil {
		return nil, err
	}

	c := &UpdateContext{Session: ss, Request: req}
	if _, err := d.ContinuityManager().Continue(
		r.Context(), r,
		prefix+"."+rid.String(),
		ContinuityOptions(payload, ss.Identity)...); err == nil {
		if payload.GetRequestID() == rid {
			return c, ErrContinuePreviousAction
		}
		return c, nil
	} else if !errors.Is(err, &continuity.ErrNotResumable) {
		return nil, err
	}

	if err := r.ParseForm(); err != nil {
		return nil, errors.WithStack(err)
	}

	payload.SetRequestID(rid)
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
		return rid, errors.WithStack(herodot.ErrBadRequest.WithReasonf("The request query parameter is missing."))
	}
	return rid, nil
}
