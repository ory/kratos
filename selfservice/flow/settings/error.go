// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package settings

import (
	"context"
	"net/http"
	"net/url"

	"go.opentelemetry.io/otel/trace"

	"github.com/ory/kratos/x/events"

	"github.com/ory/kratos/session"
	"github.com/ory/kratos/x/swagger"

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

var ErrHookAbortFlow = errors.New("aborted settings hook execution")

type (
	errorHandlerDependencies interface {
		config.Provider
		errorx.ManagementProvider
		x.WriterProvider
		x.LoggingProvider

		HandlerProvider
		FlowPersistenceProvider
		IdentityTraitsSchemas(ctx context.Context) (schema.Schemas, error)
	}

	ErrorHandlerProvider interface{ SettingsFlowErrorHandler() *ErrorHandler }

	ErrorHandler struct {
		d errorHandlerDependencies
	}
)

// Is sent when a privileged session is required to perform the settings update.
//
// swagger:model needsPrivilegedSessionError
//
//nolint:deadcode,unused
//lint:ignore U1000 Used to generate Swagger and OpenAPI definitions
type needsPrivilegedSessionError struct {
	Error swagger.GenericError `json:"error"`

	// Points to where to redirect the user to next.
	//
	// required: true
	RedirectBrowserTo string `json:"redirect_browser_to"`
}

// FlowNeedsReAuth is sent when a privileged session is required to perform the settings update.
type FlowNeedsReAuth struct {
	*herodot.DefaultError `json:"error"`

	// Points to where to redirect the user to next.
	RedirectBrowserTo string `json:"redirect_browser_to"`
}

func (e *FlowNeedsReAuth) EnhanceJSONError() interface{} {
	return e
}

func NewFlowNeedsReAuth() *FlowNeedsReAuth {
	return &FlowNeedsReAuth{
		DefaultError: herodot.ErrForbidden.WithID(text.ErrIDNeedsPrivilegedSession).
			WithReasonf("The login session is too old and thus not allowed to update these fields. Please re-authenticate."),
	}
}

func NewErrorHandler(d errorHandlerDependencies) *ErrorHandler {
	return &ErrorHandler{d: d}
}

func (s *ErrorHandler) reauthenticate(
	w http.ResponseWriter,
	r *http.Request,
	f *Flow,
	err *FlowNeedsReAuth,
) {
	returnTo := urlx.CopyWithQuery(urlx.AppendPaths(s.d.Config().SelfPublicURL(r.Context()), r.URL.Path), r.URL.Query())

	params := url.Values{}
	params.Set("refresh", "true")
	params.Set("return_to", returnTo.String())

	redirectTo := urlx.AppendPaths(urlx.CopyWithQuery(s.d.Config().SelfPublicURL(r.Context()), params), login.RouteInitBrowserFlow).String()
	err.RedirectBrowserTo = redirectTo
	if f.Type == flow.TypeAPI || x.IsJSONRequest(r) {
		s.d.Writer().WriteError(w, r, err)
		return
	}

	http.Redirect(w, r, redirectTo, http.StatusSeeOther)
}

func (s *ErrorHandler) PrepareReplacementForExpiredFlow(w http.ResponseWriter, r *http.Request, f *Flow, id *identity.Identity, err error) (*flow.ExpiredError, error) {
	e := new(flow.ExpiredError)
	if !errors.As(err, &e) {
		return nil, nil
	}

	// create new flow because the old one is not valid
	a, err := s.d.SettingsHandler().FromOldFlow(w, r, id, *f)
	if err != nil {
		return nil, err
	}

	a.UI.Messages.Add(text.NewErrorValidationSettingsFlowExpired(e.ExpiredAt))
	if err := s.d.SettingsFlowPersister().UpdateSettingsFlow(r.Context(), a); err != nil {
		return nil, err
	}

	return e.WithFlow(a), nil
}

func (s *ErrorHandler) WriteFlowError(
	w http.ResponseWriter,
	r *http.Request,
	group node.UiNodeGroup,
	f *Flow,
	id *identity.Identity,
	err error,
) {
	s.d.Audit().
		WithError(err).
		WithRequest(r).
		WithField("settings_flow", f).
		Info("Encountered self-service settings error.")

	shouldRespondWithJSON := x.IsJSONRequest(r)
	if f != nil && f.Type == flow.TypeAPI {
		shouldRespondWithJSON = true
	}

	if e := new(session.ErrNoActiveSessionFound); errors.As(err, &e) {
		if shouldRespondWithJSON {
			s.d.Writer().WriteError(w, r, err)
		} else {
			http.Redirect(w, r, urlx.AppendPaths(s.d.Config().SelfPublicURL(r.Context()), login.RouteInitBrowserFlow).String(), http.StatusSeeOther)
		}
		return
	}

	if aalErr := new(session.ErrAALNotSatisfied); errors.As(err, &aalErr) {
		if shouldRespondWithJSON {
			s.d.Writer().WriteError(w, r, aalErr)
		} else {
			http.Redirect(w, r, aalErr.RedirectTo, http.StatusSeeOther)
		}
		return
	}

	if f == nil {
		trace.SpanFromContext(r.Context()).AddEvent(events.NewSettingsFailed(r.Context(), "", ""))
		s.forward(w, r, nil, err)
		return
	}
	trace.SpanFromContext(r.Context()).AddEvent(events.NewSettingsFailed(r.Context(), string(f.Type), f.Active.String()))

	if expired, inner := s.PrepareReplacementForExpiredFlow(w, r, f, id, err); inner != nil {
		s.forward(w, r, f, err)
		return
	} else if expired != nil {
		if id == nil {
			s.forward(w, r, f, err)
			return
		}

		if f.Type == flow.TypeAPI || x.IsJSONRequest(r) {
			s.d.Writer().WriteError(w, r, expired)
		} else {
			http.Redirect(w, r, expired.GetFlow().AppendTo(s.d.Config().SelfServiceFlowSettingsUI(r.Context())).String(), http.StatusSeeOther)
		}
		return
	}

	if errors.Is(err, flow.ErrStrategyAsksToReturnToUI) {
		if shouldRespondWithJSON {
			s.d.Writer().Write(w, r, f)
		} else {
			http.Redirect(w, r, f.AppendTo(s.d.Config().SelfServiceFlowSettingsUI(r.Context())).String(), http.StatusSeeOther)
		}
		return
	}

	if e := new(FlowNeedsReAuth); errors.As(err, &e) {
		s.reauthenticate(w, r, f, e)
		return
	}

	if err := f.UI.ParseError(group, err); err != nil {
		s.forward(w, r, f, err)
		return
	}

	// Lookup the schema from the loaded configuration. This local schema
	// URL is needed for sorting the UI nodes, instead of the public URL.
	schemas, err := s.d.IdentityTraitsSchemas(r.Context())
	if err != nil {
		s.forward(w, r, f, err)
		return
	}

	schema, err := schemas.GetByID(id.SchemaID)
	if err != nil {
		s.forward(w, r, f, err)
		return
	}

	if err := sortNodes(r.Context(), f.UI.Nodes, schema.RawURL); err != nil {
		s.forward(w, r, f, err)
		return
	}

	if err := s.d.SettingsFlowPersister().UpdateSettingsFlow(r.Context(), f); err != nil {
		s.forward(w, r, f, err)
		return
	}

	if f.Type == flow.TypeBrowser && !x.IsJSONRequest(r) {
		http.Redirect(w, r, f.AppendTo(s.d.Config().SelfServiceFlowSettingsUI(r.Context())).String(), http.StatusSeeOther)
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
