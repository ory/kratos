package multistep

import (
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/schema"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/text"
	"github.com/ory/kratos/ui/node"
	"github.com/ory/kratos/x"
	"github.com/ory/x/decoderx"
	"github.com/ory/x/sqlcon"
	"github.com/pkg/errors"
	"net/http"
)

var _ login.FormHydrator = new(Strategy)
var _ login.Strategy = new(Strategy)

func (s *Strategy) handleLoginError(w http.ResponseWriter, r *http.Request, f *login.Flow, payload *updateLoginFlowWithMultiStepMethod, err error) error {
	if f != nil {
		f.UI.Nodes.SetValueAttribute("identifier", payload.Identifier)
		if f.Type == flow.TypeBrowser {
			f.UI.SetCSRF(s.d.GenerateCSRFToken(r))
		}
	}

	return err
}

func (s *Strategy) Login(w http.ResponseWriter, r *http.Request, f *login.Flow, _ *session.Session) (_ *identity.Identity, err error) {
	if !s.d.Config().SelfServiceLoginFlowTwoStepEnabled(r.Context()) {
		return nil, errors.WithStack(flow.ErrStrategyNotResponsible)
	}

	if err := login.CheckAAL(f, identity.AuthenticatorAssuranceLevel1); err != nil {
		return nil, err
	}

	var p updateLoginFlowWithMultiStepMethod
	if err := s.hd.Decode(r, &p,
		decoderx.HTTPDecoderSetValidatePayloads(true),
		decoderx.MustHTTPRawJSONSchemaCompiler(loginSchema),
		decoderx.HTTPDecoderJSONFollowsFormFormat()); err != nil {
		return nil, s.handleLoginError(w, r, f, &p, err)
	}
	f.TransientPayload = p.TransientPayload

	if err := flow.EnsureCSRF(s.d, r, f.Type, s.d.Config().DisableAPIFlowEnforcement(r.Context()), s.d.GenerateCSRFToken, p.CSRFToken); err != nil {
		return nil, s.handleLoginError(w, r, f, &p, err)
	}

	var opts []login.FormHydratorModifier

	// Look up the user by the identifier.
	identityHint, err := s.d.PrivilegedIdentityPool().FindIdentityByCredentialIdentifier(r.Context(), p.Identifier,
		// We are dealing with user input -> lookup should be case-insensitive.
		false,
	)
	if errors.Is(err, sqlcon.ErrNoRows) {
		// User not found
		if !s.d.Config().SecurityAccountEnumerationMitigate(r.Context()) {
			// We don't have to mitigate account enumeration and show the user that the account doesn't exist
			return nil, s.handleLoginError(w, r, f, &p, errors.WithStack(schema.NewAccountNotFoundError()))
		}

		// We have to mitigate account enumeration. So we continue without setting the identity hint.
	} else if err != nil {
		// An error happened during lookup
		return nil, s.handleLoginError(w, r, f, &p, err)
	} else if !s.d.Config().SecurityAccountEnumerationMitigate(r.Context()) {
		// Hydrate credentials
		if err := s.d.PrivilegedIdentityPool().HydrateIdentityAssociations(r.Context(), identityHint, identity.ExpandCredentials); err != nil {
			return nil, s.handleLoginError(w, r, f, &p, err)
		}
	}

	f.UI.ResetMessages()
	f.UI.Nodes.SetValueAttribute("identifier", p.Identifier)

	// Add identity hint
	opts = append(opts, login.WithIdentityHint(identityHint))

	for _, ls := range s.d.LoginStrategies(r.Context()) {
		populator, ok := ls.(login.FormHydrator)
		if !ok {
			continue
		}

		if err := populator.PopulateLoginMethodMultiStepSelection(r, f, opts...); err != nil {
			return nil, s.handleLoginError(w, r, f, &p, err)
		}
	}

	f.Active = identity.TwoStep
	if err = s.d.LoginFlowPersister().UpdateLoginFlow(r.Context(), f); err != nil {
		return nil, s.handleLoginError(w, r, f, &p, err)
	}

	if x.IsJSONRequest(r) {
		s.d.Writer().WriteCode(w, r, http.StatusBadRequest, f)
	} else {
		http.Redirect(w, r, f.AppendTo(s.d.Config().SelfServiceFlowLoginUI(r.Context())).String(), http.StatusSeeOther)
	}

	return nil, flow.ErrCompletedByStrategy
}

func (s *Strategy) PopulateLoginMethodRefresh(r *http.Request, sr *login.Flow) error {
	return nil
}

func (s *Strategy) PopulateLoginMethodFirstFactor(r *http.Request, sr *login.Flow) error {
	return nil
}

func (s *Strategy) PopulateLoginMethodSecondFactor(r *http.Request, sr *login.Flow) error {
	return nil
}

func (s *Strategy) PopulateLoginMethodMultiStepIdentification(r *http.Request, f *login.Flow) error {
	f.UI.SetCSRF(s.d.GenerateCSRFToken(r))

	ds, err := s.d.Config().DefaultIdentityTraitsSchemaURL(r.Context())
	if err != nil {
		return err
	}

	identifierLabel, err := login.GetIdentifierLabelFromSchema(r.Context(), ds.String())
	if err != nil {
		return err
	}

	f.UI.SetNode(node.NewInputField("identifier", "", s.NodeGroup(), node.InputAttributeTypeText, node.WithRequiredInputAttribute).WithMetaLabel(identifierLabel))
	f.UI.GetNodes().Append(node.NewInputField("method", s.ID(), s.NodeGroup(), node.InputAttributeTypeSubmit).WithMetaLabel(text.NewInfoNodeLabelContinue()))
	return nil
}

func (s *Strategy) PopulateLoginMethodMultiStepSelection(_ *http.Request, f *login.Flow, _ ...login.FormHydratorModifier) error {
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

		attrs.Type = node.InputAttributeTypeHidden
		f.UI.Nodes[k].Attributes = attrs
	}

	return nil
}

func (s *Strategy) RegisterLoginRoutes(_ *x.RouterPublic) {}
