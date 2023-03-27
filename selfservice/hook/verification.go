// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package hook

import (
	"context"
	"net/http"

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

var _ registration.PostHookPostPersistExecutor = new(Verifier)
var _ settings.PostHookPostPersistExecutor = new(Verifier)

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
		return e.do(w, r.WithContext(ctx), s.Identity, f)
	})
}

func (e *Verifier) ExecuteSettingsPostPersistHook(w http.ResponseWriter, r *http.Request, a *settings.Flow, i *identity.Identity) error {
	return otelx.WithSpan(r.Context(), "selfservice.hook.Verifier.ExecuteSettingsPostPersistHook", func(ctx context.Context) error {
		return e.do(w, r.WithContext(ctx), i, a)
	})
}

func (e *Verifier) do(w http.ResponseWriter, r *http.Request, i *identity.Identity, f flow.FlowWithContinueWith) error {
	// This is called after the identity has been created so we can safely assume that all addresses are available
	// already.

	strategy, err := e.r.GetActiveVerificationStrategy(r.Context())
	if err != nil {
		return err
	}

	for k := range i.VerifiableAddresses {
		address := &i.VerifiableAddresses[k]
		if address.Status != identity.VerifiableAddressStatusPending {
			continue
		}
		csrf := ""
		// TODO: this is pretty ugly, we should probably have a better way to handle CSRF tokens here.
		if f.GetType() != flow.TypeBrowser {
		} else if _, ok := f.(*registration.Flow); ok {
			// If this hook is executed from a registration flow, we need to regenerate the CSRF token.
			csrf = e.r.CSRFHandler().RegenerateToken(w, r)
		} else {
			// If it came from a settings flow, there already is a CSRF token, so we can just use that.
			csrf = e.r.GenerateCSRFToken(r)
		}
		verificationFlow, err := verification.NewPostHookFlow(e.r.Config(),
			e.r.Config().SelfServiceFlowVerificationRequestLifespan(r.Context()),
			csrf, r, strategy, f)
		if err != nil {
			return err
		}

		verificationFlow.State = verification.StateEmailSent

		if err := strategy.PopulateVerificationMethod(r, verificationFlow); err != nil {
			return err
		}

		if err := e.r.VerificationFlowPersister().CreateVerificationFlow(r.Context(), verificationFlow); err != nil {
			return err
		}

		if err := strategy.SendVerificationEmail(r.Context(), verificationFlow, i, address); err != nil {
			return err
		}

		flowURL := ""
		if verificationFlow.Type == flow.TypeBrowser {
			flowURL = verificationFlow.AppendTo(e.r.Config().SelfServiceFlowVerificationUI(r.Context())).String()
		}

		f.AddContinueWith(flow.NewContinueWithVerificationUI(verificationFlow, address.Value, flowURL))
	}
	return nil
}
