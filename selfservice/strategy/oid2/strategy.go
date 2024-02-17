// Copyright © 2024 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oid2

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/julienschmidt/httprouter"
	"github.com/ory/herodot"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/selfservice/flowhelpers"
	"github.com/ory/kratos/selfservice/strategy"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/text"
	"github.com/ory/kratos/ui/node"
	"github.com/ory/kratos/x"
	"github.com/ory/x/decoderx"
	"github.com/ory/x/jsonx"
	"github.com/pkg/errors"
	"net/http"
)

const (
	RouteBase = "/self-service/methods/oid2"

	RouteCallback = RouteBase + "/callback/:provider"
)

type Dependencies interface {
	config.Provider

	identity.PrivilegedPoolProvider

	session.ManagementProvider

	x.CSRFTokenGeneratorProvider
	x.LoggingProvider
	x.TracingProvider
	x.WriterProvider
}

type Strategy struct {
	d   Dependencies
	dec *decoderx.HTTP
}

func NewStrategy(d any) *Strategy {
	return &Strategy{
		d: d.(Dependencies),
	}
}

func (s *Strategy) ID() identity.CredentialsType {
	return identity.CredentialsTypeOID2
}

func (s *Strategy) NodeGroup() node.UiNodeGroup {
	return node.OpenID2Group
}

func (s *Strategy) populateMethod(r *http.Request, f flow.Flow, message func(provider string) *text.Message) error {
	conf, err := s.Config(r.Context())
	if err != nil {
		return err
	}

	providers := conf.Providers

	if lf, ok := f.(*login.Flow); ok && lf.IsForced() {
		if _, id, c := flowhelpers.GuessForcedLoginIdentifier(r, s.d, lf, s.ID()); id != nil {
			if c == nil {
				// no OID2 credentials, don't add any providers
				providers = nil
			} else {
				var credentials identity.CredentialsOid2
				if err := json.Unmarshal(c.Config, &credentials); err != nil {
					// failed to read OID2 credentials, don't add any providers
					providers = nil
				} else {
					// add only providers that can actually be used to log in as this identity
					providers = make([]Configuration, 0, len(conf.Providers))
					for i := range conf.Providers {
						for j := range credentials.Providers {
							if conf.Providers[i].ID == credentials.Providers[j].Provider {
								providers = append(providers, conf.Providers[i])
								break
							}
						}
					}
				}
			}
		}
	}

	c := f.GetUI()
	c.SetCSRF(s.d.GenerateCSRFToken(r))
	AddProviders(c, providers, message)

	return nil
}

func (s *Strategy) HandleCallback(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

func (s *Strategy) Config(ctx context.Context) (*ConfigurationCollection, error) {
	var c ConfigurationCollection

	conf := s.d.Config().SelfServiceStrategy(ctx, string(s.ID())).Config
	if err := jsonx.
		NewStrictDecoder(bytes.NewBuffer(conf)).
		Decode(&c); err != nil {
		s.d.Logger().WithError(err).WithField("config", conf)
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("Unable to decode OpenID Connect Provider configuration: %s", err))
	}

	return &c, nil
}

func (s *Strategy) provider(ctx context.Context, id string) (Provider, error) {
	if c, err := s.Config(ctx); err != nil {
		return nil, err
	} else if provider, err := c.Provider(id, s.d); err != nil {
		return nil, err
	} else {
		return provider, nil
	}
}

func (s *Strategy) setRoutes(r *x.RouterPublic) {
	wrappedHandleCallback := strategy.IsDisabled(s.d, s.ID().String(), s.HandleCallback)
	if handle, _, _ := r.Lookup("GET", RouteCallback); handle == nil {
		r.GET(RouteCallback, wrappedHandleCallback)
	}
}
