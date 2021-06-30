package settings

import (
	"context"
	"net/http"
	"net/url"
	"time"

	"github.com/ory/kratos/ui/node"

	"github.com/pkg/errors"

	"github.com/ory/herodot"
	"github.com/ory/x/urlx"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/schema"
	"github.com/ory/kratos/selfservice/errorx"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/text"
	"github.com/ory/kratos/x"
)

var (
	ErrHookAbortRequest = errors.New("aborted settings hook execution")
)

type (
	errorHandlerDependencies interface {
		config.Provider
		errorx.ManagementProvider
		x.WriterProvider
		x.LoggingProvider

		HandlerProvider
		FlowPersistenceProvider
		IdentityTraitsSchemas(ctx context.Context) schema.Schemas
	}

	ErrorHandlerProvider interface{ SettingsFlowErrorHandler() *ErrorHandler }

	ErrorHandler struct {
		d errorHandlerDependencies
	}

	FlowExpiredError struct {
		*herodot.DefaultError
		ago time.Duration
	}

	FlowNeedsReAuth struct {
		*herodot.DefaultError
	}
)

func NewFlowNeedsReAuth() *FlowNeedsReAuth {
	return &FlowNeedsReAuth{DefaultError: herodot.ErrForbidden.
		WithReasonf("The login session is too old and thus not allowed to update these fields. Please re-authenticate.")}
}

func NewFlowExpiredError(at time.Time) *FlowExpiredError {
	ago := time.Since(at)
	return &FlowExpiredError{
		ago: ago,
		DefaultError: herodot.ErrBadRequest.
			WithError("settings flow expired").
			WithReasonf(`The settings flow has expired. Please restart the flow.`).
			WithReasonf("The settings flow expired %.2f minutes ago, please try again.", ago.Minutes()),
	}
}

func NewErrorHandler(d errorHandlerDependencies) *ErrorHandler {
	return &ErrorHandler{d: d}
}

func (s *ErrorHandler) reauthenticate(
	w http.ResponseWriter,
	r *http.Request,
	f *Flow,
	err error,
) {
	if f.Type == flow.TypeAPI || x.IsJSONRequest(r) {
		s.d.Writer().WriteError(w, r, err)
		return
	}

	returnTo := urlx.CopyWithQuery(urlx.AppendPaths(s.d.Config(r.Context()).SelfPublicURL(r), r.URL.Path), r.URL.Query())
	http.Redirect(w, r, urlx.AppendPaths(urlx.CopyWithQuery(s.d.Config(r.Context()).SelfPublicURL(r),
		url.Values{"refresh": {"true"}, "return_to": {returnTo.String()}}),
		login.RouteInitBrowserFlow).String(), http.StatusSeeOther)
}

func (s *ErrorHandler) WriteFlowError(
	w http.ResponseWriter,
	r *http.Request,
	group node.Group,
	f *Flow,
	id *identity.Identity,
	err error,
) {
	s.d.Audit().
		WithError(err).
		WithRequest(r).
		WithField("settings_flow", f).
		Info("Encountered self-service settings error.")

	if f == nil {
		s.forward(w, r, f, err)
		return
	}

	if e := new(FlowExpiredError); errors.As(err, &e) {
		if id == nil {
			s.forward(w, r, f, err)
			return
		}

		// create new flow because the old one is not valid
		a, err := s.d.SettingsHandler().NewFlow(w, r, id, f.Type)
		if err != nil {
			// failed to create a new session and redirect to it, handle that error as a new one
			s.WriteFlowError(w, r, group, f, id, err)
			return
		}

		a.UI.Messages.Add(text.NewErrorValidationSettingsFlowExpired(e.ago))
		if err := s.d.SettingsFlowPersister().UpdateSettingsFlow(r.Context(), a); err != nil {
			s.forward(w, r, a, err)
			return
		}

		if f.Type == flow.TypeAPI || x.IsJSONRequest(r) {
			http.Redirect(w, r, urlx.CopyWithQuery(urlx.AppendPaths(s.d.Config(r.Context()).SelfPublicURL(r),
				RouteGetFlow), url.Values{"id": {a.ID.String()}}).String(), http.StatusSeeOther)
		} else {
			http.Redirect(w, r, a.AppendTo(s.d.Config(r.Context()).SelfServiceFlowSettingsUI()).String(), http.StatusSeeOther)
		}
		return
	}

	if e := new(FlowNeedsReAuth); errors.As(err, &e) {
		s.reauthenticate(w, r, f, err)
		return
	}

	if err := f.UI.ParseError(group, err); err != nil {
		s.forward(w, r, f, err)
		return
	}

	// Lookup the schema from the loaded configuration. This local schema
	// URL is needed for sorting the UI nodes, instead of the public URL.
	schema, err := s.d.IdentityTraitsSchemas(r.Context()).GetByID(id.SchemaID)
	if err != nil {
		s.forward(w, r, f, err)
		return
	}

	if err := sortNodes(f.UI.Nodes, schema.RawURL); err != nil {
		s.forward(w, r, f, err)
		return
	}

	if err := s.d.SettingsFlowPersister().UpdateSettingsFlow(r.Context(), f); err != nil {
		s.forward(w, r, f, err)
		return
	}

	if f.Type == flow.TypeBrowser && !x.IsJSONRequest(r) {
		http.Redirect(w, r, f.AppendTo(s.d.Config(r.Context()).SelfServiceFlowSettingsUI()).String(), http.StatusSeeOther)
		return
	}

	updatedFlow, innerErr := s.d.SettingsFlowPersister().GetSettingsFlow(r.Context(), f.ID)
	if innerErr != nil {
		s.forward(w, r, updatedFlow, innerErr)
	}

	s.d.Writer().WriteCode(w, r, x.RecoverStatusCode(err, http.StatusBadRequest), updatedFlow)
}

func (s *ErrorHandler) forward(w http.ResponseWriter, r *http.Request, rr *Flow, err error) {
	if rr == nil {
		if x.IsJSONRequest(r) {
			s.d.Writer().WriteError(w, r, err)
			return
		}
		s.d.SelfServiceErrorManager().Forward(r.Context(), w, r, err)
		return
	}

	if rr.Type == flow.TypeAPI || x.IsJSONRequest(r) {
		s.d.Writer().WriteErrorCode(w, r, x.RecoverStatusCode(err, http.StatusBadRequest), err)
	} else {
		s.d.SelfServiceErrorManager().Forward(r.Context(), w, r, err)
	}
}
