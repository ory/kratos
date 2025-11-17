// Copyright © 2024 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package idfirst

import (
	"net/http"

	"go.opentelemetry.io/otel/attribute"

	"github.com/ory/x/otelx"

	"github.com/ory/kratos/schema"

	"github.com/pkg/errors"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/text"
	"github.com/ory/kratos/ui/node"
	"github.com/ory/kratos/x"
	"github.com/ory/x/decoderx"
	"github.com/ory/x/sqlcon"
)

var (
	_                     login.AAL1FormHydrator = new(Strategy)
	_                     login.Strategy         = new(Strategy)
	ErrNoCredentialsFound                        = errors.New("no credentials found")
)

func (s *Strategy) handleLoginError(r *http.Request, f *login.Flow, payload UpdateLoginFlowWithIdentifierFirstMethod, err error) error {
	if f != nil {
		f.UI.Nodes.SetValueAttribute("identifier", payload.Identifier)
		if f.Type == flow.TypeBrowser {
			f.UI.SetCSRF(s.d.GenerateCSRFToken(r))
		}
	}

	return err
}

func (s *Strategy) Login(w http.ResponseWriter, r *http.Request, f *login.Flow, sess *session.Session) (_ *identity.Identity, err error) {
	ctx, span := s.d.Tracer(r.Context()).Tracer().Start(r.Context(), "selfservice.strategy.idfirst.Strategy.Login")
	defer otelx.End(span, &err)

	if !s.d.Config().SelfServiceLoginFlowIdentifierFirstEnabled(ctx) {
		span.SetAttributes(attribute.String("not_responsible_reason", "strategy is not enabled"))
		return nil, errors.WithStack(flow.ErrStrategyNotResponsible)
	}

	if err := login.CheckAAL(f, identity.AuthenticatorAssuranceLevel1); err != nil {
		span.SetAttributes(attribute.String("not_responsible_reason", "requested AAL is not sufficient"))
		return nil, err
	}

	var p UpdateLoginFlowWithIdentifierFirstMethod
	if err := s.hd.Decode(r, &p,
		decoderx.HTTPDecoderSetValidatePayloads(true),
		decoderx.MustHTTPRawJSONSchemaCompiler(LoginSchema),
		decoderx.HTTPDecoderJSONFollowsFormFormat()); err != nil {
		return nil, s.handleLoginError(r, f, p, err)
	}
	f.TransientPayload = p.TransientPayload

	if err := flow.EnsureCSRF(s.d, r, f.Type, s.d.Config().DisableAPIFlowEnforcement(ctx), s.d.GenerateCSRFToken, p.CSRFToken); err != nil {
		return nil, s.handleLoginError(r, f, p, err)
	}

	var opts []login.FormHydratorModifier

	// Look up the user by the identifier.
	identityHint, err := s.d.PrivilegedIdentityPool().FindIdentityByCredentialIdentifier(ctx, p.Identifier,
		// We are dealing with user input -> lookup should be case-insensitive.
		false,
	)
	if errors.Is(err, sqlcon.ErrNoRows) {
		// If the user is not found, we still want to potentially show the UI for some method. That's why we don't exit here.
		// We have to mitigate account enumeration. So we continue without setting the identity hint.
		//
		// This will later be handled by `didPopulate`.
	} else if err != nil {
		// An error happened during lookup
		return nil, s.handleLoginError(r, f, p, err)
	} else if !s.d.Config().SecurityAccountEnumerationMitigate(ctx) {
		// Hydrate credentials
		if err := s.d.PrivilegedIdentityPool().HydrateIdentityAssociations(ctx, identityHint, identity.ExpandCredentials); err != nil {
			return nil, s.handleLoginError(r, f, p, err)
		}
	}

	f.UI.ResetMessages()
	f.UI.Nodes.SetValueAttribute("identifier", p.Identifier)

	// Add identity hint
	opts = append(opts, login.WithIdentityHint(identityHint))
	opts = append(opts, login.WithIdentifier(p.Identifier))

	didPopulate := false
	for _, ls := range s.d.LoginStrategies(ctx, login.PrepareOrganizations(r, f, sess)...) {
		populator, ok := ls.(login.AAL1FormHydrator)
		if !ok {
			continue
		}

		if err := populator.PopulateLoginMethodIdentifierFirstCredentials(r, f, opts...); errors.Is(err, login.ErrBreakLoginPopulate) {
			didPopulate = true
			break
		} else if errors.Is(err, ErrNoCredentialsFound) {
			// This strategy is not responsible for this flow. We do not set didPopulate to true if that happens.
		} else if err != nil {
			return nil, s.handleLoginError(r, f, p, err)
		} else {
			didPopulate = true

			// If the method supports FastLogin, we can attempt a fast path.
			if strategy, ok := ls.(login.FastLoginStrategy); ok {
				err = strategy.FastLogin1FA(w, r, f, sess)
				if err == nil {
					return nil, flow.ErrCompletedByStrategy
				} else if !errors.Is(err, flow.ErrStrategyNotResponsible) {
					return nil, err
				}
			}
		}
	}

	// If no strategy populated, it means that the account (very likely) does not exist. We show a user not found error,
	// but only if account enumeration mitigation is disabled. Otherwise, we proceed to render the rest of the form.
	if !didPopulate && !s.d.Config().SecurityAccountEnumerationMitigate(ctx) {
		return nil, s.handleLoginError(r, f, p, errors.WithStack(schema.NewAccountNotFoundError()))
	}

	// We found credentials - hide the identifier.
	f.UI.GetNodes().RemoveMatching(node.NewInputField("method", s.ID(), s.NodeGroup(), node.InputAttributeTypeSubmit))

	// We set the identifier to hidden, so it's still available in the form but not visible to the user.
	for k, n := range f.UI.Nodes {
		if n.ID() != "identifier" {
			continue
		}

		attrs, ok := f.UI.Nodes[k].Attributes.(*node.InputAttributes)
		if !ok {
			continue
		}
		attrs.Autocomplete = node.InputAttributeAutocompleteUsernameWebauthn

		attrs.Type = node.InputAttributeTypeHidden
		f.UI.Nodes[k].Attributes = attrs
	}

	f.Active = s.ID()
	if err = s.d.LoginFlowPersister().UpdateLoginFlow(ctx, f); err != nil {
		return nil, s.handleLoginError(r, f, p, err)
	}

	if x.IsJSONRequest(r) {
		s.d.Writer().WriteCode(w, r, http.StatusBadRequest, f)
	} else {
		http.Redirect(w, r, f.AppendTo(s.d.Config().SelfServiceFlowLoginUI(ctx)).String(), http.StatusSeeOther)
	}

	return nil, flow.ErrCompletedByStrategy
}

func (s *Strategy) PopulateLoginMethodFirstFactorRefresh(r *http.Request, sr *login.Flow, _ *session.Session) error {
	return nil
}

func (s *Strategy) PopulateLoginMethodFirstFactor(r *http.Request, sr *login.Flow) error {
	return nil
}

func (s *Strategy) PopulateLoginMethodIdentifierFirstIdentification(r *http.Request, f *login.Flow) error {
	f.UI.SetCSRF(s.d.GenerateCSRFToken(r))

	ds, err := f.IdentitySchema.URL(r.Context(), s.d.Config())
	if err != nil {
		return err
	}

	identifierLabel, err := login.GetIdentifierLabelFromSchema(r.Context(), ds.String())
	if err != nil {
		return err
	}

	f.UI.SetNode(node.NewInputField("identifier", "", s.NodeGroup(), node.InputAttributeTypeText, node.WithInputAttributes(func(a *node.InputAttributes) {
		a.Autocomplete = node.InputAttributeAutocompleteUsernameWebauthn
		a.Required = true
	})).WithMetaLabel(identifierLabel))
	f.UI.GetNodes().Append(node.NewInputField("method", s.ID(), s.NodeGroup(), node.InputAttributeTypeSubmit).WithMetaLabel(text.NewInfoNodeLabelContinue()))
	return nil
}

func (s *Strategy) PopulateLoginMethodIdentifierFirstCredentials(_ *http.Request, f *login.Flow, opts ...login.FormHydratorModifier) error {
	return ErrNoCredentialsFound
}

func (s *Strategy) RegisterLoginRoutes(_ *x.RouterPublic) {}
