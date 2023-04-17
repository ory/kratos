// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package hook

import (
	"context"
	"net/http"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/registration"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/x"
	"github.com/ory/x/otelx"
)

var (
	_ registration.PostHookPostPersistExecutor = new(SessionIssuer)
)

type (
	showVerificationUIDependencies interface {
		x.WriterProvider
		config.Provider
	}

	ShowVerfificationUIProvider interface {
		HookShowVerificationUI() *ShowVerificationUIHook
	}

	// ShowVerificationUIHook is a post registration hook that redirects browser clients to the verification UI.
	ShowVerificationUIHook struct {
		d showVerificationUIDependencies
	}
)

func NewShowVerificationUIHook(d showVerificationUIDependencies) *ShowVerificationUIHook {
	return &ShowVerificationUIHook{d: d}
}

// ExecutePostRegistrationPostPersistHook adds redirect headers and status code if the request is a browser request.
// If the request is not a browser request, this hook does nothing.
func (e *ShowVerificationUIHook) ExecutePostRegistrationPostPersistHook(w http.ResponseWriter, r *http.Request, f *registration.Flow, s *session.Session) error {
	return otelx.WithSpan(r.Context(), "selfservice.hook.SessionIssuer.ExecutePostRegistrationPostPersistHook", func(ctx context.Context) error {
		return e.execute(w, r.WithContext(ctx), f, s)
	})
}

func (e *ShowVerificationUIHook) execute(w http.ResponseWriter, r *http.Request, f *registration.Flow, s *session.Session) error {
	if !x.IsBrowserRequest(r) {
		// this hook is only intended to be used by browsers, as it redirects to the verification ui
		// JSON API clients should use the `continue_with` field to continue the flow
		return nil
	}

	var vf *flow.ContinueWithVerificationUI
	for _, c := range f.ContinueWithItems {
		if item, ok := c.(*flow.ContinueWithVerificationUI); ok {
			vf = item
		}
	}

	if vf != nil {
		redir := e.d.Config().SelfServiceFlowVerificationUI(r.Context())
		f.ReturnToVerification = vf.AppendTo(redir).String()
	}

	return nil
}
