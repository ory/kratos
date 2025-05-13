// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package hook

import (
	"encoding/json"
	"net/http"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/flow/settings"

	"github.com/gofrs/uuid"
	"github.com/tidwall/gjson"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/selfservice/flow/registration"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/ui/node"
	"github.com/ory/kratos/x"
	"github.com/ory/x/otelx"
)

var (
	_ registration.PostHookPostPersistExecutor = new(ShowVerificationUIHook)
	_ login.PostHookExecutor                   = new(ShowVerificationUIHook)
	_ settings.PostHookPostPersistExecutor     = new(ShowVerificationUIHook)
)

type (
	showVerificationUIDependencies interface {
		x.WriterProvider
		config.Provider
		x.TracingProvider
		x.LoggingProvider
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
func (e *ShowVerificationUIHook) ExecutePostRegistrationPostPersistHook(_ http.ResponseWriter, r *http.Request, f *registration.Flow, _ *session.Session) error {
	return e.execute(r, f)
}

// ExecuteLoginPostHook adds redirect headers and status code if the request is a browser request.
// If the request is not a browser request, this hook does nothing.
func (e *ShowVerificationUIHook) ExecuteLoginPostHook(_ http.ResponseWriter, r *http.Request, _ node.UiNodeGroup, f *login.Flow, _ *session.Session) error {
	return e.execute(r, f)
}

func (e *ShowVerificationUIHook) ExecuteSettingsPostPersistHook(w http.ResponseWriter, r *http.Request, f *settings.Flow, id *identity.Identity, s *session.Session) error {
	return e.execute(r, f)
}

type verificationUIFlow interface {
	flow.InternalContexter
	flow.FlowWithContinueWith
}

type loginOrRegistrationFlow interface {
	SetReturnToVerification(string)
}

func (e *ShowVerificationUIHook) execute(r *http.Request, f verificationUIFlow) (err error) {
	ctx, span := e.d.Tracer(r.Context()).Tracer().Start(r.Context(), "selfservice.hook.ShowVerificationUIHook.Do")
	defer otelx.End(span, &err)

	var cw flow.ContinueWithVerificationUIFlow

	verificationFlow := gjson.GetBytes(f.GetInternalContext(), InternalContextRegistrationVerificationFlow).Raw
	if verificationFlow == "" {
		// TODO This is a fallback for flows that do not have the appropriate internalContext yet.
		// Remove this once we have released this change.
		for _, c := range f.ContinueWith() {
			if item, ok := c.(*flow.ContinueWithVerificationUI); ok {
				cw = item.Flow
			}
		}
	} else {
		if err := json.Unmarshal([]byte(verificationFlow), &cw); err != nil {
			return err
		}
	}

	if cw.ID == uuid.Nil {
		e.d.Logger().WithRequest(r).Warn("Ignoring hook `show_verification_ui` because no verification flow ID was found in the registration flow. Cannot show verification UI. This is likely a configuration issue or a bug.")
		return nil
	}

	vf := flow.NewContinueWithVerificationUI(cw.ID, cw.VerifiableAddress, cw.URL)
	f.AddContinueWith(vf)

	if returnToVerification, ok := f.(loginOrRegistrationFlow); ok && x.IsBrowserRequest(r) {
		verificationUI := e.d.Config().SelfServiceFlowVerificationUI(ctx)
		returnToVerification.SetReturnToVerification(vf.AppendTo(verificationUI).String())
	}

	return nil
}
