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
	"github.com/ory/kratos/selfservice/flow/registration"
	"github.com/ory/kratos/selfservice/flow/settings"
	"github.com/ory/kratos/selfservice/flow/verification"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/x"
	"github.com/ory/x/otelx"
)

var (
	_ registration.PostHookPostPersistExecutor = new(Verifier)
	_ settings.PostHookPostPersistExecutor     = new(Verifier)
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
	isRegistrationFlow := f.GetFlowName() == flow.RegistrationFlow

	for k := range i.VerifiableAddresses {
		address := &i.VerifiableAddresses[k]
		if address.Status != identity.VerifiableAddressStatusPending {
			continue
		}

		var csrf string

		// TODO: this is pretty ugly, we should probably have a better way to handle CSRF tokens here.
		if isBrowserFlow {
			if isRegistrationFlow {
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
