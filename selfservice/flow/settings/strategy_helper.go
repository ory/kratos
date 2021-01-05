package settings

import (
	"fmt"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/gofrs/uuid"
	"github.com/julienschmidt/httprouter"
	"github.com/pkg/errors"

	"github.com/ory/x/sqlcon"

	"github.com/ory/herodot"

	"github.com/ory/kratos/continuity"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/x"
)

var (
	ErrContinuePreviousAction = errors.New("found prior action")
)

type UpdatePayload interface {
	GetFlowID() uuid.UUID
	SetFlowID(id uuid.UUID)
}

type UpdateContext struct {
	Session *session.Session
	Flow    *Flow
}

func (c UpdateContext) GetSessionIdentity() *identity.Identity {
	if c.Session == nil {
		return nil
	}
	return c.Session.Identity
}

func PrepareUpdate(d interface {
	x.LoggingProvider
	continuity.ManagementProvider
	session.ManagementProvider
	FlowPersistenceProvider
}, w http.ResponseWriter, r *http.Request, name string, payload UpdatePayload) (*UpdateContext, error) {
	ss, err := d.SessionManager().FetchFromRequest(r.Context(), r)
	if err != nil {
		return new(UpdateContext), err
	}

	rid, err := GetFlowID(r)
	if err != nil {
		return new(UpdateContext), err
	}

	payload.SetFlowID(rid)
	req, err := d.SettingsFlowPersister().GetSettingsFlow(r.Context(), rid)
	if errors.Is(err, sqlcon.ErrNoRows) {
		return new(UpdateContext), errors.WithStack(herodot.ErrNotFound.WithReasonf("The settings request could not be found. Please restart the flow."))
	} else if err != nil {
		return new(UpdateContext), err
	}

	if err := req.Valid(ss); err != nil {
		return new(UpdateContext), err
	}

	c := &UpdateContext{Session: ss, Flow: req}
	if req.Type == flow.TypeAPI {
		return c, nil
	}

	if _, err := d.ContinuityManager().Continue(r.Context(), w, r, name, ContinuityOptions(payload, ss.Identity)...); err == nil {
		if payload.GetFlowID() == rid {
			return c, ErrContinuePreviousAction
		}
		d.Logger().
			WithField("package", pkgName).
			WithField("stack_trace", fmt.Sprintf("%s", debug.Stack())).
			WithField("expected_request_id", payload.GetFlowID()).
			WithField("actual_request_id", rid).
			Debug("Flow ID from continuity manager does not match Flow ID from request.")
		return c, nil
	} else if !errors.Is(err, &continuity.ErrNotResumable) {
		return new(UpdateContext), err
	}

	return c, nil
}

func ContinuityOptions(p interface{}, i *identity.Identity) []continuity.ManagerOption {
	return []continuity.ManagerOption{
		continuity.WithPayload(p),
		continuity.WithIdentity(i),
		continuity.WithLifespan(time.Minute * 15),
	}
}

func GetFlowID(r *http.Request) (uuid.UUID, error) {
	rid := x.ParseUUID(r.URL.Query().Get("flow"))
	if rid == uuid.Nil {
		return rid, errors.WithStack(herodot.ErrBadRequest.WithReasonf("The request query parameter is missing or malformed."))
	}
	return rid, nil
}

func OnUnauthenticated(reg interface {
	config.Providers
	x.WriterProvider
}) func(http.ResponseWriter, *http.Request, httprouter.Params) {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		handler := session.RedirectOnUnauthenticated(reg.Configuration(r.Context()).SelfServiceFlowLoginUI().String())
		if x.IsJSONRequest(r) {
			handler = session.RespondWithJSONErrorOnAuthenticated(reg.Writer(), herodot.ErrUnauthorized.WithReasonf("A valid ORY Session Cookie or ORY Session Token is missing."))
		}

		handler(w, r, ps)
	}
}
