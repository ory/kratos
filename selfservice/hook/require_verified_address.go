// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package hook

import (
	"net/http"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/verification"
	"github.com/ory/kratos/text"
	"github.com/ory/kratos/x"
	"github.com/ory/kratos/x/nosurfx"
	"github.com/ory/x/otelx"
	"github.com/ory/x/otelx/semconv"

	"github.com/pkg/errors"

	"github.com/ory/kratos/ui/node"

	"github.com/ory/herodot"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/session"
)

var _ login.PostHookExecutor = new(AddressVerifier)

type (
	addressVerifierDependencies interface {
		config.Provider
		nosurfx.CSRFTokenGeneratorProvider
		nosurfx.CSRFProvider
		verification.StrategyProvider
		verification.FlowPersistenceProvider
		identity.PrivilegedPoolProvider
		x.WriterProvider
		x.TracingProvider
	}
	AddressVerifier struct {
		r addressVerifierDependencies
	}
)

func NewAddressVerifier(r addressVerifierDependencies) *AddressVerifier {
	return &AddressVerifier{
		r: r,
	}
}

func (e *AddressVerifier) ExecuteLoginPostHook(w http.ResponseWriter, r *http.Request, _ node.UiNodeGroup, f *login.Flow, s *session.Session) (err error) {
	ctx, span := e.r.Tracer(r.Context()).Tracer().Start(r.Context(), "selfservice.hook.Verifier.do")
	r = r.WithContext(ctx)
	defer otelx.End(span, &err)

	// TODO remove once flag is removed.
	if e.r.Config().UseLegacyRequireVerifiedLoginError(ctx) {
		if f.Active != identity.CredentialsTypePassword {
			span.AddEvent(semconv.NewDeprecatedFeatureUsedEvent(ctx, "legacy_require_verified_login_error"))
			return nil
		}
	}
	// END TODO

	if len(s.Identity.VerifiableAddresses) == 0 {
		return errors.WithStack(herodot.ErrMisconfiguration.WithReason("A misconfiguration prevents login. Expected to find a verification address but this identity does not have one assigned."))
	}

	for _, va := range s.Identity.VerifiableAddresses {
		if va.Verified {
			return nil
		}
	}

	strategy, err := e.r.GetActiveVerificationStrategy(ctx)
	if err != nil {
		return err
	}

	// TODO remove once flag is removed.
	if e.r.Config().UseLegacyRequireVerifiedLoginError(ctx) {
		span.AddEvent(semconv.NewDeprecatedFeatureUsedEvent(ctx, "legacy_require_verified_login_error"))
		return login.ErrAddressNotVerified
	}
	// END TODO

	i := s.Identity
	for k := range i.VerifiableAddresses {
		address := &i.VerifiableAddresses[k]
		if address.Value == "" {
			continue
		}

		verificationFlow, err := verification.NewPostHookFlow(e.r.Config(),
			e.r.Config().SelfServiceFlowVerificationRequestLifespan(ctx),
			e.r.GenerateCSRFToken(r), r, strategy, f)
		if err != nil {
			return err
		}

		verificationFlow.State = flow.StateEmailSent
		if err := strategy.PopulateVerificationMethod(r, verificationFlow); err != nil {
			return err
		}

		verificationFlow.UI.Nodes.Append(
			node.NewInputField(address.Via, address.Value, node.CodeGroup, node.InputAttributeTypeSubmit).
				WithMetaLabel(text.NewInfoNodeResendOTP()),
		)

		if err := e.r.VerificationFlowPersister().CreateVerificationFlow(ctx, verificationFlow); err != nil {
			return err
		}

		if err := strategy.SendVerificationCode(ctx, verificationFlow, i, address); err != nil {
			return err
		}

		flowURL := verificationFlow.AppendTo(e.r.Config().SelfServiceFlowVerificationUI(ctx)).String()
		continueWith := flow.NewContinueWithVerificationUI(verificationFlow.ID, address.Value, flowURL)
		f.AddContinueWith(continueWith)

		if x.IsJSONRequest(r) {
			e.r.Writer().WriteErrorCode(w, r, http.StatusForbidden, flow.ErrorWithContinueWith(login.ErrAddressNotVerified, continueWith))
			return errors.WithStack(login.ErrHookAbortFlow)
		}

		if x.IsBrowserRequest(r) {
			http.Redirect(w, r, flowURL, http.StatusSeeOther)
			return errors.WithStack(login.ErrHookAbortFlow)
		}

		return errors.WithStack(login.ErrHookAbortFlow)
	}

	return login.ErrAddressNotVerified
}
