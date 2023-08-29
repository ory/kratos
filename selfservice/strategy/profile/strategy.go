// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package profile

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/ory/jsonschema/v3"

	"github.com/ory/kratos/text"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
	"github.com/tidwall/sjson"

	"github.com/ory/herodot"
	"github.com/ory/kratos/continuity"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/schema"
	"github.com/ory/kratos/selfservice/errorx"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/settings"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/ui/container"
	"github.com/ory/kratos/ui/node"
	"github.com/ory/kratos/x"
	"github.com/ory/x/decoderx"
)

var _ settings.Strategy = new(Strategy)

type (
	strategyDependencies interface {
		x.CSRFProvider
		x.CSRFTokenGeneratorProvider
		x.WriterProvider
		x.LoggingProvider

		config.Provider

		continuity.ManagementProvider

		session.HandlerProvider
		session.ManagementProvider

		identity.ValidationProvider
		identity.ManagementProvider
		identity.PrivilegedPoolProvider

		errorx.ManagementProvider

		settings.HookExecutorProvider
		settings.ErrorHandlerProvider
		settings.FlowPersistenceProvider
		settings.StrategyProvider
		settings.HooksProvider

		schema.IdentityTraitsProvider
	}
	Strategy struct {
		d  strategyDependencies
		dc *decoderx.HTTP
	}
)

func NewStrategy(d strategyDependencies) *Strategy {
	return &Strategy{d: d, dc: decoderx.NewHTTP()}
}

func (s *Strategy) SettingsStrategyID() string {
	return settings.StrategyProfile
}

func (s *Strategy) RegisterSettingsRoutes(public *x.RouterPublic) {}

func (s *Strategy) PopulateSettingsMethod(r *http.Request, id *identity.Identity, f *settings.Flow) error {
	schemas, err := s.d.Config().IdentityTraitsSchemas(r.Context())
	if err != nil {
		return err
	}

	traitsSchema, err := schemas.FindSchemaByID(id.SchemaID)
	if err != nil {
		return err
	}

	// use a schema compiler that disables identifiers
	schemaCompiler := jsonschema.NewCompiler()
	nodes, err := container.NodesFromJSONSchema(r.Context(), node.ProfileGroup, traitsSchema.URL, "", schemaCompiler)
	if err != nil {
		return err
	}

	for _, n := range nodes {
		f.UI.SetNode(n)
	}

	f.UI.SetCSRF(s.d.GenerateCSRFToken(r))
	f.UI.UpdateNodeValuesFromJSON(json.RawMessage(id.Traits), "traits", node.ProfileGroup)
	f.UI.Nodes.Append(node.NewInputField("method", "profile", node.ProfileGroup, node.InputAttributeTypeSubmit).WithMetaLabel(text.NewInfoNodeLabelSave()))

	return nil
}

func (s *Strategy) Settings(w http.ResponseWriter, r *http.Request, f *settings.Flow, ss *session.Session) (*settings.UpdateContext, error) {
	var p updateSettingsFlowWithProfileMethod
	ctxUpdate, err := settings.PrepareUpdate(s.d, w, r, f, ss, settings.ContinuityKey(s.SettingsStrategyID()), &p)
	if errors.Is(err, settings.ErrContinuePreviousAction) {
		return ctxUpdate, s.continueFlow(w, r, ctxUpdate, &p)
	} else if err != nil {
		return ctxUpdate, s.handleSettingsError(w, r, ctxUpdate, nil, &p, err)
	}

	if err := flow.MethodEnabledAndAllowedFromRequest(r, f.GetFlowName(), s.SettingsStrategyID(), s.d); err != nil {
		return ctxUpdate, err
	}

	option, err := s.newSettingsProfileDecoder(r.Context(), ctxUpdate.GetSessionIdentity())
	if err != nil {
		return ctxUpdate, s.handleSettingsError(w, r, ctxUpdate, nil, &p, err)
	}

	if err := s.dc.Decode(r, &p, option,
		decoderx.HTTPDecoderAllowedMethods("POST", "GET"),
		decoderx.HTTPDecoderSetValidatePayloads(true),
		decoderx.HTTPDecoderJSONFollowsFormFormat(),
	); err != nil {
		return ctxUpdate, s.handleSettingsError(w, r, ctxUpdate, nil, &p, err)
	}

	// Reset after decoding form
	p.SetFlowID(ctxUpdate.Flow.ID)

	if err := s.continueFlow(w, r, ctxUpdate, &p); err != nil {
		return ctxUpdate, s.handleSettingsError(w, r, ctxUpdate, nil, &p, err)
	}

	return ctxUpdate, nil
}

func (s *Strategy) continueFlow(w http.ResponseWriter, r *http.Request, ctxUpdate *settings.UpdateContext, p *updateSettingsFlowWithProfileMethod) error {
	if err := flow.MethodEnabledAndAllowed(r.Context(), flow.SettingsFlow, s.SettingsStrategyID(), p.Method, s.d); err != nil {
		return err
	}

	if err := flow.EnsureCSRF(s.d, r, ctxUpdate.Flow.Type, s.d.Config().DisableAPIFlowEnforcement(r.Context()), s.d.GenerateCSRFToken, p.CSRFToken); err != nil {
		return err
	}

	if len(p.Traits) == 0 {
		return errors.WithStack(herodot.ErrBadRequest.WithReasonf("Did not receive any value changes."))
	}

	if err := s.hydrateForm(r, ctxUpdate.Flow, ctxUpdate.Session, p.Traits); err != nil {
		return err
	}

	options := []identity.ManagerOption{identity.ManagerExposeValidationErrorsForInternalTypeAssertion}
	ttl := s.d.Config().SelfServiceFlowSettingsPrivilegedSessionMaxAge(r.Context())
	if ctxUpdate.Session.AuthenticatedAt.Add(ttl).After(time.Now()) {
		options = append(options, identity.ManagerAllowWriteProtectedTraits)
	}

	update, err := s.d.IdentityManager().SetTraits(r.Context(), ctxUpdate.GetSessionIdentity().ID, identity.Traits(p.Traits), options...)
	if err != nil {
		if errors.Is(err, identity.ErrProtectedFieldModified) {
			return settings.NewFlowNeedsReAuth()
		}
		return err
	}

	ctxUpdate.UpdateIdentity(update)
	return nil
}

// Update Settings Flow with Profile Method
//
// swagger:model updateSettingsFlowWithProfileMethod
//
//nolint:deadcode,unused
//lint:ignore U1000 Used to generate Swagger and OpenAPI definitions
type updateSettingsFlowWithProfileMethod struct {
	// Traits
	//
	// The identity's traits.
	//
	// required: true
	Traits json.RawMessage `json:"traits"`

	// Method
	//
	// Should be set to profile when trying to update a profile.
	//
	// required: true
	Method string `json:"method"`

	// FlowIDRequestID is the flow ID.
	//
	// swagger:ignore
	FlowID string `json:"flow"`

	// The Anti-CSRF Token
	//
	// This token is only required when performing browser flows.
	CSRFToken string `json:"csrf_token"`
}

func (p *updateSettingsFlowWithProfileMethod) GetFlowID() uuid.UUID {
	return x.ParseUUID(p.FlowID)
}

func (p *updateSettingsFlowWithProfileMethod) SetFlowID(rid uuid.UUID) {
	p.FlowID = rid.String()
}

func (s *Strategy) hydrateForm(r *http.Request, ar *settings.Flow, ss *session.Session, traits json.RawMessage) error {
	if traits != nil {
		ar.UI.Nodes.ResetNodesWithPrefix("traits.")
		ar.UI.UpdateNodeValuesFromJSON(traits, "traits", node.ProfileGroup)
	}
	ar.UI.SetCSRF(s.d.GenerateCSRFToken(r))

	return nil
}

// handleSettingsError is a convenience function for handling all types of errors that may occur (e.g. validation error)
// during a settings request.
func (s *Strategy) handleSettingsError(w http.ResponseWriter, r *http.Request, puc *settings.UpdateContext, traits json.RawMessage, p *updateSettingsFlowWithProfileMethod, err error) error {
	if e := new(settings.FlowNeedsReAuth); errors.As(err, &e) {
		if err := s.d.ContinuityManager().Pause(r.Context(), w, r,
			settings.ContinuityKey(s.SettingsStrategyID()),
			settings.ContinuityOptions(p, puc.GetSessionIdentity())...); err != nil {
			return err
		}
	}

	if puc.Flow != nil {
		if traits == nil {
			if len(p.Traits) >= 0 {
				traits = p.Traits
			} else {
				traits = json.RawMessage(puc.GetSessionIdentity().Traits)
			}
		}

		if err := s.hydrateForm(r, puc.Flow, puc.Session, traits); err != nil {
			return err
		}
	}

	return err
}

// newSettingsProfileDecoder returns a decoderx.HTTPDecoderOption with a JSON Schema for type assertion and
// validation.
func (s *Strategy) newSettingsProfileDecoder(ctx context.Context, i *identity.Identity) (decoderx.HTTPDecoderOption, error) {
	schemas, err := s.d.IdentityTraitsSchemas(ctx)
	if err != nil {
		return nil, err
	}
	ss, err := schemas.GetByID(i.SchemaID)
	if err != nil {
		return nil, err
	}
	raw, err := sjson.SetBytes(settingsSchema,
		"properties.traits.$ref", ss.URL.String()+"#/properties/traits")
	if err != nil {
		return nil, errors.WithStack(err)
	}

	o, err := decoderx.HTTPRawJSONSchemaCompiler(raw)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return o, nil
}

func (s *Strategy) NodeGroup() node.UiNodeGroup {
	return node.ProfileGroup
}
