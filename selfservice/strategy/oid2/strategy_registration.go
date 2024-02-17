// Copyright © 2024 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oid2

import (
	"encoding/json"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/registration"
	"github.com/ory/kratos/text"
	"github.com/ory/kratos/x"
	"github.com/ory/x/decoderx"
	"github.com/ory/x/otelx"
	"github.com/pkg/errors"
	"github.com/tidwall/sjson"
	"github.com/yohcop/openid-go"
	"net/http"
	"strings"
)

var _ registration.Strategy = new(Strategy)

func (s *Strategy) RegisterRegistrationRoutes(public *x.RouterPublic) {
	s.setRoutes(public)
}

func (s *Strategy) PopulateRegistrationMethod(r *http.Request, f *registration.Flow) error {
	return s.populateMethod(r, f, text.NewInfoRegistrationWith)
}

// Update Registration Flow with OpenID 2.0 Method
//
// swagger:model updateRegistrationFlowWithOid2Method
type UpdateRegistrationFlowWithOid2Method struct {
	// The provider to register with
	//
	// required: true
	Provider string `json:"provider"`

	// The CSRF Token
	CSRFToken string `json:"csrf_token"`

	// The identity traits
	Traits json.RawMessage `json:"traits"`

	// Method to use
	//
	// This field must be set to `oid2` when using the oid2 method.
	//
	// required: true
	Method string `json:"method"`
}

func (s *Strategy) Register(w http.ResponseWriter, r *http.Request, f *registration.Flow, i *identity.Identity) (err error) {
	ctx, span := s.d.Tracer(r.Context()).Tracer().Start(r.Context(), "selfservice.strategy.oid2.strategy.Register")
	defer otelx.End(span, &err)

	var p UpdateRegistrationFlowWithOid2Method
	if err := s.newLinkDecoder(&p, r); err != nil {
		return err
	}

	pid := p.Provider // this can come from both url query and post body
	if pid == "" {
		return errors.WithStack(flow.ErrStrategyNotResponsible)
	}

	if !strings.EqualFold(strings.ToLower(p.Method), s.ID().String()) && p.Method != "" {
		// the user is sending a method that is not oid2, but the payload includes a provider
		s.d.Audit().
			WithRequest(r).
			WithField("provider", p.Provider).
			WithField("method", p.Method).
			Warn("The payload includes a `provider` field but is using a method other than `oid2`. Therefore, Open ID 2.0 sign in will not be executed.")
		return errors.WithStack(flow.ErrStrategyNotResponsible)
	}

	provider, err := s.provider(ctx, pid)
	if err != nil {
		return err
	}

	redirectUrl, err := openid.RedirectURL(provider.Config().DiscoveryUrl, provider.GetRedirectUrl(ctx), s.d.Config().Oid2RedirectURIBase(ctx).String())
	if err != nil {
		return err
	}

	if x.IsJSONRequest(r) {
		s.d.Writer().WriteError(w, r, flow.NewBrowserLocationChangeRequiredError(redirectUrl))
	} else {
		http.Redirect(w, r, redirectUrl, http.StatusSeeOther)
	}

	return errors.WithStack(flow.ErrCompletedByStrategy)
}

// TODO #3631 ...what?
func (s *Strategy) newLinkDecoder(p interface{}, r *http.Request) error {
	ds, err := s.d.Config().DefaultIdentityTraitsSchemaURL(r.Context())
	if err != nil {
		return err
	}

	raw, err := sjson.SetBytes(linkSchema, "properties.traits.$ref", ds.String()+"#/properties/traits")
	if err != nil {
		return errors.WithStack(err)
	}

	compiler, err := decoderx.HTTPRawJSONSchemaCompiler(raw)
	if err != nil {
		return errors.WithStack(err)
	}

	if err := s.dec.Decode(r, &p, compiler,
		decoderx.HTTPKeepRequestBody(true),
		decoderx.HTTPDecoderSetValidatePayloads(false),
		decoderx.HTTPDecoderUseQueryAndBody(),
		decoderx.HTTPDecoderAllowedMethods("POST", "GET"),
		decoderx.HTTPDecoderJSONFollowsFormFormat(),
	); err != nil {
		return errors.WithStack(err)
	}

	return nil
}
