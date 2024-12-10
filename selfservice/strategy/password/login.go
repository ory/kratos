// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package password

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"

	"go.opentelemetry.io/otel/attribute"

	"github.com/ory/x/otelx"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"

	"github.com/ory/herodot"
	"github.com/ory/kratos/hash"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/schema"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/selfservice/flowhelpers"
	"github.com/ory/kratos/selfservice/hook"
	"github.com/ory/kratos/selfservice/strategy/idfirst"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/text"
	"github.com/ory/kratos/ui/node"
	"github.com/ory/kratos/x"
	"github.com/ory/x/decoderx"
	"github.com/ory/x/stringsx"
)

var _ login.FormHydrator = new(Strategy)

func (s *Strategy) RegisterLoginRoutes(r *x.RouterPublic) {
}

func (s *Strategy) handleLoginError(r *http.Request, f *login.Flow, payload updateLoginFlowWithPasswordMethod, err error) error {
	if f != nil {
		f.UI.Nodes.ResetNodes("password")
		f.UI.Nodes.SetValueAttribute("identifier", stringsx.Coalesce(payload.Identifier, payload.LegacyIdentifier))
		if f.Type == flow.TypeBrowser {
			f.UI.SetCSRF(s.d.GenerateCSRFToken(r))
		}
	}

	return err
}

func (s *Strategy) Login(w http.ResponseWriter, r *http.Request, f *login.Flow, _ *session.Session) (i *identity.Identity, err error) {
	ctx, span := s.d.Tracer(r.Context()).Tracer().Start(r.Context(), "selfservice.strategy.password.Strategy.Login")
	defer otelx.End(span, &err)

	if err := login.CheckAAL(f, identity.AuthenticatorAssuranceLevel1); err != nil {
		span.SetAttributes(attribute.String("not_responsible_reason", "requested AAL is not AAL1"))
		return nil, err
	}

	if err := flow.MethodEnabledAndAllowedFromRequest(r, f.GetFlowName(), s.ID().String(), s.d); err != nil {
		return nil, err
	}

	var p updateLoginFlowWithPasswordMethod
	if err := s.hd.Decode(r, &p,
		decoderx.HTTPDecoderSetValidatePayloads(true),
		decoderx.MustHTTPRawJSONSchemaCompiler(loginSchema),
		decoderx.HTTPDecoderJSONFollowsFormFormat()); err != nil {
		return nil, s.handleLoginError(r, f, p, err)
	}
	f.TransientPayload = p.TransientPayload

	if err := flow.EnsureCSRF(s.d, r, f.Type, s.d.Config().DisableAPIFlowEnforcement(ctx), s.d.GenerateCSRFToken, p.CSRFToken); err != nil {
		return nil, s.handleLoginError(r, f, p, err)
	}

	identifier := stringsx.Coalesce(p.Identifier, p.LegacyIdentifier)
	i, c, err := s.d.PrivilegedIdentityPool().FindByCredentialsIdentifier(ctx, s.ID(), identifier)
	if err != nil {
		time.Sleep(x.RandomDelay(s.d.Config().HasherArgon2(ctx).ExpectedDuration, s.d.Config().HasherArgon2(ctx).ExpectedDeviation))
		return nil, s.handleLoginError(r, f, p, errors.WithStack(schema.NewInvalidCredentialsError()))
	}

	var o identity.CredentialsPassword
	d := json.NewDecoder(bytes.NewBuffer(c.Config))
	if err := d.Decode(&o); err != nil {
		return nil, herodot.ErrInternalServerError.WithReason("The password credentials could not be decoded properly").WithDebug(err.Error()).WithWrap(err)
	}

	if o.ShouldUsePasswordMigrationHook() {
		pwHook := s.d.Config().PasswordMigrationHook(ctx)
		if !pwHook.Enabled {
			return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("Password migration hook is not enabled but password migration is requested."))
		}

		migrationHook := hook.NewPasswordMigrationHook(s.d, pwHook.Config)
		err = migrationHook.Execute(ctx, &hook.PasswordMigrationRequest{Identifier: identifier, Password: p.Password})
		if err != nil {
			return nil, s.handleLoginError(r, f, p, err)
		}

		if err := s.migratePasswordHash(ctx, i.ID, []byte(p.Password)); err != nil {
			return nil, s.handleLoginError(r, f, p, err)
		}
	} else {
		if err := hash.Compare(ctx, []byte(p.Password), []byte(o.HashedPassword)); err != nil {
			return nil, s.handleLoginError(r, f, p, errors.WithStack(schema.NewInvalidCredentialsError()))
		}

		if !s.d.Hasher(ctx).Understands([]byte(o.HashedPassword)) {
			if err := s.migratePasswordHash(ctx, i.ID, []byte(p.Password)); err != nil {
				s.d.Logger().Warnf("Unable to migrate password hash for identity %s: %s Keeping existing password hash and continuing.", i.ID, err)
			}
		}
	}

	f.Active = s.ID()
	if err = s.d.LoginFlowPersister().UpdateLoginFlow(ctx, f); err != nil {
		return nil, s.handleLoginError(r, f, p, errors.WithStack(herodot.ErrInternalServerError.WithReason("Could not update flow").WithDebug(err.Error())))
	}

	return i, nil
}

func (s *Strategy) migratePasswordHash(ctx context.Context, identifier uuid.UUID, password []byte) (err error) {
	ctx, span := s.d.Tracer(ctx).Tracer().Start(ctx, "selfservice.strategy.password.Strategy.migratePasswordHash")
	defer otelx.End(span, &err)

	hpw, err := s.d.Hasher(ctx).Generate(ctx, password)
	if err != nil {
		return err
	}
	co, err := json.Marshal(&identity.CredentialsPassword{HashedPassword: string(hpw)})
	if err != nil {
		return errors.Wrap(err, "unable to encode password configuration to JSON")
	}

	i, err := s.d.PrivilegedIdentityPool().GetIdentityConfidential(ctx, identifier)
	if err != nil {
		return err
	}

	c, ok := i.GetCredentials(s.ID())
	if !ok {
		return errors.New("expected to find password credential but could not")
	}

	c.Config = co
	i.SetCredentials(s.ID(), *c)

	return s.d.IdentityManager().Update(ctx, i, identity.ManagerAllowWriteProtectedTraits)
}

func (s *Strategy) PopulateLoginMethodFirstFactorRefresh(r *http.Request, sr *login.Flow) (err error) {
	ctx := r.Context()
	ctx, span := s.d.Tracer(ctx).Tracer().Start(ctx, "selfservice.strategy.password.Strategy.PopulateLoginMethodFirstFactorRefresh")
	defer otelx.End(span, &err)

	identifier, id, _ := flowhelpers.GuessForcedLoginIdentifier(r, s.d, sr, s.ID())
	if identifier == "" {
		return nil
	}

	// If we don't have a password set, do not show the password field.
	count, err := s.CountActiveFirstFactorCredentials(ctx, id.Credentials)
	if err != nil {
		return err
	} else if count == 0 {
		return nil
	}

	sr.UI.SetCSRF(s.d.GenerateCSRFToken(r))
	sr.UI.SetNode(node.NewInputField("identifier", identifier, node.DefaultGroup, node.InputAttributeTypeHidden))
	sr.UI.SetNode(NewPasswordNode("password", node.InputAttributeAutocompleteCurrentPassword))
	sr.UI.GetNodes().Append(node.NewInputField("method", "password", node.PasswordGroup, node.InputAttributeTypeSubmit).WithMetaLabel(text.NewInfoLogin()))
	return nil
}

func (s *Strategy) PopulateLoginMethodSecondFactor(r *http.Request, sr *login.Flow) error {
	return nil
}

func (s *Strategy) PopulateLoginMethodSecondFactorRefresh(r *http.Request, sr *login.Flow) error {
	return nil
}

func (s *Strategy) addIdentifierNode(r *http.Request, sr *login.Flow) error {
	ds, err := s.d.Config().DefaultIdentityTraitsSchemaURL(r.Context())
	if err != nil {
		return err
	}

	identifierLabel, err := login.GetIdentifierLabelFromSchema(r.Context(), ds.String())
	if err != nil {
		return err
	}

	sr.UI.SetNode(node.NewInputField("identifier", "", node.DefaultGroup, node.InputAttributeTypeText, node.WithRequiredInputAttribute).WithMetaLabel(identifierLabel))
	return nil
}

func (s *Strategy) PopulateLoginMethodFirstFactor(r *http.Request, sr *login.Flow) error {
	if err := s.addIdentifierNode(r, sr); err != nil {
		return err
	}

	sr.UI.SetCSRF(s.d.GenerateCSRFToken(r))
	sr.UI.SetNode(NewPasswordNode("password", node.InputAttributeAutocompleteCurrentPassword))
	sr.UI.GetNodes().Append(node.NewInputField("method", "password", node.PasswordGroup, node.InputAttributeTypeSubmit).WithMetaLabel(text.NewInfoLoginPassword()))
	return nil
}

func (s *Strategy) PopulateLoginMethodIdentifierFirstCredentials(r *http.Request, sr *login.Flow, opts ...login.FormHydratorModifier) (err error) {
	ctx, span := s.d.Tracer(r.Context()).Tracer().Start(r.Context(), "selfservice.strategy.password.Strategy.PopulateLoginMethodIdentifierFirstCredentials")
	defer otelx.End(span, &err)

	o := login.NewFormHydratorOptions(opts)

	var count int
	if o.IdentityHint != nil {
		var err error
		// If we have an identity hint we can perform identity credentials discovery and
		// hide this credential if it should not be included.
		if count, err = s.CountActiveFirstFactorCredentials(ctx, o.IdentityHint.Credentials); err != nil {
			return err
		}
	}

	if count > 0 || s.d.Config().SecurityAccountEnumerationMitigate(ctx) {
		sr.UI.SetCSRF(s.d.GenerateCSRFToken(r))
		sr.UI.SetNode(NewPasswordNode("password", node.InputAttributeAutocompleteCurrentPassword))
		sr.UI.GetNodes().Append(node.NewInputField("method", "password", node.PasswordGroup, node.InputAttributeTypeSubmit).WithMetaLabel(text.NewInfoLoginPassword()))
	}

	if count == 0 {
		return errors.WithStack(idfirst.ErrNoCredentialsFound)
	}

	return nil
}

func (s *Strategy) PopulateLoginMethodIdentifierFirstIdentification(r *http.Request, sr *login.Flow) error {
	return nil
}
