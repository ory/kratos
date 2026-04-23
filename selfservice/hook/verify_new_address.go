// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package hook

import (
	"context"
	"encoding/json"
	"net/http"
	"slices"

	"github.com/pkg/errors"
	"github.com/tidwall/sjson"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/settings"
	"github.com/ory/kratos/selfservice/flow/verification"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/text"
	"github.com/ory/kratos/ui/node"
	"github.com/ory/kratos/x"
	"github.com/ory/kratos/x/nosurfx"
	"github.com/ory/x/httpx"
	"github.com/ory/x/otelx"
)

var _ settings.PostHookPrePersistExecutor = new(VerifyNewAddress)

type (
	verifyNewAddressDependencies interface {
		config.Provider
		nosurfx.CSRFTokenGeneratorProvider
		nosurfx.CSRFProvider
		verification.StrategyProvider
		verification.FlowPersistenceProvider
		identity.PrivilegedPoolProvider
		identity.PendingTraitsChangePersistenceProvider
		settings.FlowPersistenceProvider
		httpx.WriterProvider
		otelx.Provider
		session.ManagementProvider
	}

	VerifyNewAddressProvider interface {
		HookVerifyNewAddress() *VerifyNewAddress
	}

	VerifyNewAddress struct {
		r verifyNewAddressDependencies
	}
)

func NewVerifyNewAddress(r verifyNewAddressDependencies) *VerifyNewAddress {
	return &VerifyNewAddress{r: r}
}

func (e *VerifyNewAddress) ExecuteSettingsPrePersistHook(
	w http.ResponseWriter, r *http.Request,
	f *settings.Flow, i *identity.Identity, s *session.Session,
) error {
	return otelx.WithSpan(r.Context(), "selfservice.hook.VerifyNewAddress.ExecuteSettingsPrePersistHook", func(ctx context.Context) error {
		return e.execute(ctx, w, r, f, i, s)
	})
}

func (e *VerifyNewAddress) execute(
	ctx context.Context, w http.ResponseWriter, r *http.Request,
	f *settings.Flow, proposed *identity.Identity, s *session.Session,
) error {
	original := s.Identity
	if original == nil {
		return nil
	}

	// Find verifiable addresses that changed between original and proposed.
	changed := findChangedAddresses(original.VerifiableAddresses, proposed.VerifiableAddresses)
	if len(changed) == 0 {
		return nil // No address changes — let the flow proceed normally.
	}

	// At this point, we know the user is trying to change an address. We need to check that the session is currently privileged.
	if !e.r.SessionManager().IsPrivileged(ctx, s) {
		return errors.WithStack(settings.NewFlowNeedsReAuth())
	}

	// Only one address change can be verified at a time. Reject if multiple
	// addresses changed to prevent unverified address changes from being applied.
	if len(changed) > 1 {
		f.UI.Messages.Clear()
		f.UI.Messages.Add(text.NewErrorValidationSettingsTooManyAddressChanges())
		if err := e.r.SettingsFlowPersister().UpdateSettingsFlow(ctx, f); err != nil {
			return err
		}
		if x.IsJSONRequest(r) {
			e.r.Writer().Write(w, r, f)
			return errors.WithStack(settings.ErrHookAbortFlow)
		}
		http.Redirect(w, r, f.AppendTo(e.r.Config().SelfServiceFlowSettingsUI(ctx)).String(), http.StatusSeeOther)
		return errors.WithStack(settings.ErrHookAbortFlow)
	}

	// Delete any existing pending changes for this identity.
	if err := e.r.PendingTraitsChangePersister().DeletePendingTraitsChangesByIdentity(ctx, original.ID); err != nil {
		return err
	}

	// Process the first changed address (we can only redirect to one verification flow).
	addr := changed[0]

	// Create the verification flow.
	strategies, primaryStrategy, err := e.r.GetActiveVerificationStrategies(ctx)
	if err != nil {
		return err
	}

	var csrf string
	if f.Type == flow.TypeBrowser {
		csrf = e.r.GenerateCSRFToken(r)
	}

	verificationFlow, err := verification.NewPostHookFlow(e.r.Config(),
		e.r.Config().SelfServiceFlowVerificationRequestLifespan(ctx),
		csrf, r, strategies, f)
	if err != nil {
		return err
	}

	verificationFlow.State = flow.StateEmailSent
	if err := primaryStrategy.PopulateVerificationMethod(r, verificationFlow); err != nil {
		return err
	}

	verificationFlow.UI.Nodes.Append(
		node.NewInputField(addr.Via, addr.Value, node.CodeGroup, node.InputAttributeTypeSubmit).
			WithMetaLabel(text.NewInfoNodeResendOTP()),
	)

	if err := e.r.VerificationFlowPersister().CreateVerificationFlow(ctx, verificationFlow); err != nil {
		return err
	}

	// Create PendingTraitsChange record.
	ptc := &identity.PendingTraitsChange{
		IdentityID:         original.ID,
		NewAddressValue:    addr.Value,
		NewAddressVia:      addr.Via,
		OriginalTraitsHash: identity.HashTraits(json.RawMessage(original.Traits)),
		ProposedTraits:     json.RawMessage(proposed.Traits),
		VerificationFlowID: verificationFlow.ID,
		Status:             identity.PendingTraitsChangeStatusPending,
	}
	if err := e.r.PendingTraitsChangePersister().CreatePendingTraitsChange(ctx, ptc); err != nil {
		return err
	}

	// Send verification code to the new address.
	if err := primaryStrategy.SendVerificationCode(ctx, verificationFlow, proposed, ptc); err != nil {
		return err
	}

	// Store verification flow reference in InternalContext (for show_verification_ui).
	flowURL := ""
	if verificationFlow.Type == flow.TypeBrowser {
		flowURL = verificationFlow.AppendTo(e.r.Config().SelfServiceFlowVerificationUI(ctx)).String()
	}
	continueWith := flow.NewContinueWithVerificationUI(verificationFlow.ID, addr.Value, flowURL)
	internalContext, err := sjson.SetBytes(f.GetInternalContext(), InternalContextRegistrationVerificationFlow, continueWith.Flow)
	if err != nil {
		return err
	}
	f.SetInternalContext(internalContext)
	f.AddContinueWith(continueWith)

	// Persist the updated settings flow (with InternalContext/ContinueWith).
	if err := e.r.SettingsFlowPersister().UpdateSettingsFlow(ctx, f); err != nil {
		return err
	}

	// Write the HTTP response and abort the flow.
	if x.IsJSONRequest(r) || verificationFlow.Type == flow.TypeAPI {
		e.r.Writer().Write(w, r, f)
		return errors.WithStack(settings.ErrHookAbortFlow)
	}

	if flowURL != "" {
		http.Redirect(w, r, flowURL, http.StatusSeeOther)
		return errors.WithStack(settings.ErrHookAbortFlow)
	}

	// This should not happen: we only end up here if, the flow is an API flow, but

	return errors.WithStack(settings.ErrHookAbortFlow)
}

type changedAddress struct {
	Value string
	Via   string
}

// findChangedAddresses compares original and proposed verifiable addresses and
// returns addresses that are new or have a different value.
func findChangedAddresses(original, proposed []identity.VerifiableAddress) []changedAddress {
	origByVia := make(map[string][]string, len(original))
	for _, a := range original {
		origByVia[a.Via] = append(origByVia[a.Via], a.Value)
	}

	var changed []changedAddress
	for _, p := range proposed {
		if !slices.Contains(origByVia[p.Via], p.Value) {
			changed = append(changed, changedAddress{Value: p.Value, Via: p.Via})
		}
	}
	return changed
}
