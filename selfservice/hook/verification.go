// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package hook

import (
	"context"
	"net/http"

	"github.com/gofrs/uuid"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/selfservice/flow/registration"
	"github.com/ory/kratos/selfservice/flow/settings"
	"github.com/ory/kratos/selfservice/flow/verification"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/text"
	"github.com/ory/kratos/ui/node"
	"github.com/ory/kratos/x"
	"github.com/ory/x/otelx"
)

var (
	_ registration.PostHookPostPersistExecutor = new(Verifier)
	_ settings.PostHookPostPersistExecutor     = new(Verifier)
	_ login.PostHookExecutor                   = new(Verifier)
)

type (
	verifierDependencies interface {
		config.Provider
		x.CSRFTokenGeneratorProvider
		x.CSRFProvider
		verification.StrategyProvider
		verification.FlowPersistenceProvider
		identity.PrivilegedPoolProvider
		x.WriterProvider
		x.TracingProvider
	}
	Verifier struct {
		r verifierDependencies
	}
)

func NewVerifier(r verifierDependencies) *Verifier {
	return &Verifier{r: r}
}

func (e *Verifier) ExecutePostRegistrationPostPersistHook(w http.ResponseWriter, r *http.Request, f *registration.Flow, s *session.Session) error {
	return otelx.WithSpan(r.Context(), "selfservice.hook.Verifier.ExecutePostRegistrationPostPersistHook", func(ctx context.Context) error {
		return e.do(w, r.WithContext(ctx), s.Identity, f, func(v *verification.Flow) {
			v.OAuth2LoginChallenge = f.OAuth2LoginChallenge
			v.SessionID = uuid.NullUUID{UUID: s.ID, Valid: true}
			v.IdentityID = uuid.NullUUID{UUID: s.Identity.ID, Valid: true}
			v.AMR = s.AMR
		})
	})
}

func (e *Verifier) ExecuteSettingsPostPersistHook(w http.ResponseWriter, r *http.Request, f *settings.Flow, i *identity.Identity, _ *session.Session) error {
	return otelx.WithSpan(r.Context(), "selfservice.hook.Verifier.ExecuteSettingsPostPersistHook", func(ctx context.Context) error {
		return e.do(w, r.WithContext(ctx), i, f, nil)
	})
}

func (e *Verifier) ExecuteLoginPostHook(w http.ResponseWriter, r *http.Request, g node.UiNodeGroup, f *login.Flow, s *session.Session) (err error) {
	ctx, span := e.r.Tracer(r.Context()).Tracer().Start(r.Context(), "selfservice.hook.Verifier.ExecuteLoginPostHook")
	r = r.WithContext(ctx)
	defer otelx.End(span, &err)
	if f.RequestedAAL != identity.AuthenticatorAssuranceLevel1 {
		span.AddEvent("Skipping verification hook because AAL is not 1")
		return nil
	}

	return e.do(w, r.WithContext(ctx), s.Identity, f, nil)
}

func (e *Verifier) do(
	w http.ResponseWriter,
	r *http.Request,
	i *identity.Identity,
	f flow.FlowWithContinueWith,
	flowCallback func(*verification.Flow),
) error {
	// This is called after the identity has been created so we can safely assume that all addresses are available
	// already.
	ctx := r.Context()

	strategy, err := e.r.GetActiveVerificationStrategy(ctx)
	if err != nil {
		return err
	}

	isBrowserFlow := f.GetType() == flow.TypeBrowser
	isRegistrationOrLoginFlow := f.GetFlowName() == flow.RegistrationFlow || f.GetFlowName() == flow.LoginFlow

	for k := range i.VerifiableAddresses {
		address := &i.VerifiableAddresses[k]
		if isRegistrationOrLoginFlow && address.Verified {
			continue
		} else if !isRegistrationOrLoginFlow && address.Status != identity.VerifiableAddressStatusPending {
			// In case of the settings flow, we only want to create a new verification flow if there is no pending
			// verification flow for the address. Otherwise, we would create a new verification flow for each setting,
			// even if the address did not change.
			continue
		}

		var csrf string

		// TODO: this is pretty ugly, we should probably have a better way to handle CSRF tokens here.
		if isBrowserFlow {
			if isRegistrationOrLoginFlow {
				// If this hook is executed from a registration flow, we need to regenerate the CSRF token.
				csrf = e.r.CSRFHandler().RegenerateToken(w, r)
			} else {
				// If it came from a settings flow, there already is a CSRF token, so we can just use that.
				csrf = e.r.GenerateCSRFToken(r)
			}
		}

		verificationFlow, err := verification.NewPostHookFlow(e.r.Config(),
			e.r.Config().SelfServiceFlowVerificationRequestLifespan(ctx),
			csrf, r, strategy, f)
		if err != nil {
			return err
		}

		if flowCallback != nil {
			flowCallback(verificationFlow)
		}

		verificationFlow.State = flow.StateEmailSent

		if err := strategy.PopulateVerificationMethod(r, verificationFlow); err != nil {
			return err
		}

		if address.Value != "" && address.Via == identity.VerifiableAddressTypeEmail {
			verificationFlow.UI.Nodes.Append(
				node.NewInputField(address.Via, address.Value, node.CodeGroup, node.InputAttributeTypeSubmit).
					WithMetaLabel(text.NewInfoNodeResendOTP()),
			)
		}

		if err := e.r.VerificationFlowPersister().CreateVerificationFlow(ctx, verificationFlow); err != nil {
			return err
		}

		if err := strategy.SendVerificationEmail(ctx, verificationFlow, i, address); err != nil {
			return err
		}

		flowURL := ""
		if verificationFlow.Type == flow.TypeBrowser {
			flowURL = verificationFlow.AppendTo(e.r.Config().SelfServiceFlowVerificationUI(ctx)).String()
		}

		f.AddContinueWith(flow.NewContinueWithVerificationUI(verificationFlow, address.Value, flowURL))
	}
	return nil
}
