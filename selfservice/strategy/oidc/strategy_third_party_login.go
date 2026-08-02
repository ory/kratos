// Copyright © 2024 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oidc

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/ory/herodot"
	"github.com/ory/kratos/continuity"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/x/redir"
	"github.com/ory/x/otelx"
)

// HandleThirdPartyLoginInit implements OpenID Connect Third-Party Login
// Initiation (spec Section 4). An external party redirects the user here with
// an `iss` parameter identifying the OIDC provider. Kratos looks up the
// matching provider, creates a login flow, and redirects directly to the
// provider's authorization endpoint — no login UI is shown.
func (s *Strategy) HandleThirdPartyLoginInit(w http.ResponseWriter, r *http.Request) {
	var err error
	ctx := r.Context()
	ctx, span := s.d.Tracer(ctx).Tracer().Start(ctx, "selfservice.strategy.oidc.HandleThirdPartyLoginInit")
	defer otelx.End(span, &err)
	r = r.WithContext(ctx)

	if err = r.ParseForm(); err != nil {
		s.d.SelfServiceErrorManager().Forward(ctx, w, r,
			errors.WithStack(herodot.ErrBadRequest.WithReasonf("Unable to parse form: %s", err)))
		return
	}

	iss := r.FormValue("iss")
	loginHint := r.FormValue("login_hint")
	targetLinkURI := r.FormValue("target_link_uri")

	if iss == "" {
		s.d.SelfServiceErrorManager().Forward(ctx, w, r,
			errors.WithStack(herodot.ErrBadRequest.WithReason("The `iss` parameter is required.")))
		return
	}

	issURL, parseErr := url.Parse(iss)
	if parseErr != nil || issURL.Host == "" ||
		(issURL.Scheme != "https" && !s.d.Config().IsInsecureDevMode(ctx)) {
		s.d.SelfServiceErrorManager().Forward(ctx, w, r,
			errors.WithStack(herodot.ErrBadRequest.WithReasonf(
				"The `iss` parameter must be a valid HTTPS URL, got: %q", iss)))
		return
	}

	provider, _, err := s.findProviderByIssuer(ctx, iss)
	if err != nil {
		s.d.SelfServiceErrorManager().Forward(ctx, w, r, err)
		return
	}

	conf := s.d.Config()
	var validatedTargetURI *url.URL
	if targetLinkURI != "" {
		if validatedTargetURI, err = redir.SecureRedirectTo(r,
			conf.SelfServiceBrowserDefaultReturnTo(ctx),
			redir.SecureRedirectReturnTo(targetLinkURI),
			redir.SecureRedirectAllowURLs(conf.SelfServiceBrowserAllowedReturnToDomains(ctx)),
			redir.SecureRedirectAllowSelfServiceURLs(conf.SelfPublicURL(ctx)),
		); err != nil {
			s.d.SelfServiceErrorManager().Forward(ctx, w, r,
				errors.WithStack(herodot.ErrBadRequest.WithReasonf(
					"The `target_link_uri` is not allowed: %s", err)))
			return
		}

		q := r.URL.Query()
		q.Set("return_to", targetLinkURI)
		r.URL.RawQuery = q.Encode()
	}

	loginFlow, _, err := s.d.LoginHandler().NewLoginFlow(w, r, flow.TypeBrowser)
	if err != nil {
		if errors.Is(err, login.ErrAlreadyLoggedIn) {
			returnTo := conf.SelfServiceBrowserDefaultReturnTo(ctx)
			if validatedTargetURI != nil {
				returnTo = validatedTargetURI
			}
			http.Redirect(w, r, returnTo.String(), http.StatusSeeOther)
			return
		}
		if errors.Is(err, flow.ErrCompletedByStrategy) {
			return
		}
		s.d.SelfServiceErrorManager().Forward(ctx, w, r, err)
		return
	}
	if loginFlow == nil {
		// PreLoginHook already wrote the response.
		return
	}

	state, pkce, err := s.GenerateState(ctx, provider, loginFlow)
	if err != nil {
		s.d.SelfServiceErrorManager().Forward(ctx, w, r, err)
		return
	}

	if err = s.d.ContinuityManager().Pause(ctx, w, r, sessionName,
		continuity.WithPayload(&AuthCodeContainer{
			State:  state,
			FlowID: loginFlow.ID.String(),
		}),
		continuity.WithLifespan(time.Minute*30)); err != nil {
		s.d.SelfServiceErrorManager().Forward(ctx, w, r, err)
		return
	}

	loginFlow.Active = s.ID()
	if err = s.d.LoginFlowPersister().UpdateLoginFlow(ctx, loginFlow); err != nil {
		s.d.SelfServiceErrorManager().Forward(ctx, w, r,
			errors.WithStack(herodot.ErrInternalServerError.WithReason("Could not update flow").WithWrap(err)))
		return
	}

	up := make(map[string]string)
	if loginHint != "" {
		up["login_hint"] = loginHint
	}

	codeURL, err := getAuthRedirectURL(ctx, provider, loginFlow, state, up, pkce)
	if err != nil {
		s.d.SelfServiceErrorManager().Forward(ctx, w, r, err)
		return
	}

	http.Redirect(w, r, codeURL, http.StatusSeeOther)
}

// findProviderByIssuer looks up a configured OIDC provider whose IssuerURL
// matches the given issuer string (trailing-slash normalized).
func (s *Strategy) findProviderByIssuer(ctx context.Context, issuer string) (Provider, *Configuration, error) {
	conf, err := s.Config(ctx)
	if err != nil {
		return nil, nil, err
	}

	issuer = strings.TrimRight(issuer, "/")
	for _, p := range conf.Providers {
		if strings.TrimRight(p.IssuerURL, "/") == issuer {
			provider, err := conf.Provider(p.ID, s.d)
			if err != nil {
				return nil, nil, err
			}
			return provider, &p, nil
		}
	}

	return nil, nil, errors.WithStack(
		herodot.ErrNotFound.WithReasonf("No configured OpenID Connect provider matches the issuer %q", issuer),
	)
}
