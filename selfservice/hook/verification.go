// Copyright © 2022 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package hook

import (
	"context"
	"net/http"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

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
		verification.StrategyProvider
		verification.FlowPersistenceProvider
	}
	Verifier struct {
		r verifierDependencies
	}
)

func NewVerifier(r verifierDependencies) *Verifier {
	return &Verifier{r: r}
}

func (e *Verifier) ExecutePostRegistrationPostPersistHook(_ http.ResponseWriter, r *http.Request, f *registration.Flow, s *session.Session) error {
	return otelx.WithSpan(r.Context(), "selfservice.hook.ExecutePostRegistrationPostPersistHook", func(ctx context.Context) error {
		return e.do(r.WithContext(ctx), s.Identity, f)
	}, trace.WithAttributes(attribute.String("hook", "verifier")))
}

func (e *Verifier) ExecuteSettingsPostPersistHook(w http.ResponseWriter, r *http.Request, a *settings.Flow, i *identity.Identity) error {
	return otelx.WithSpan(r.Context(), "selfservice.hook.ExecuteSettingsPostPersistHook", func(ctx context.Context) error {
		return e.do(r.WithContext(ctx), i, a)
	}, trace.WithAttributes(attribute.String("hook", "verifier")))
}

func (e *Verifier) do(r *http.Request, i *identity.Identity, f flow.Flow) error {
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
		verificationFlow, err := verification.NewPostHookFlow(e.r.Config(),
			e.r.Config().SelfServiceFlowVerificationRequestLifespan(r.Context()),
			e.r.GenerateCSRFToken(r), r, strategy, f)
		if err != nil {
			return err
		}

		verificationFlow.State = verification.StateEmailSent

		if err := e.r.VerificationFlowPersister().CreateVerificationFlow(r.Context(), verificationFlow); err != nil {
			return err
		}

		if err := strategy.SendVerificationEmail(r.Context(), verificationFlow, i, address); err != nil {
			return err
		}

	}
	return nil
}
