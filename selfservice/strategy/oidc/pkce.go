// Copyright © 2024 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oidc

import (
	"context"
	"slices"

	gooidc "github.com/coreos/go-oidc/v3/oidc"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"golang.org/x/oauth2"

	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/x"
)

type pkceDependencies interface {
	x.LoggingProvider
	x.HTTPClientProvider
}

func MaybeUsePKCE(ctx context.Context, d pkceDependencies, _p Provider, f flow.InternalContexter) (pkceEnabled bool, err error) {
	if _p.Config().PKCE == "never" {
		return false, nil
	}
	p, ok := _p.(OAuth2Provider)
	if !ok {
		if _p.Config().PKCE == "force" {
			return false, errors.New("Provider does not support OAuth2, cannot force PKCE")
		}
		return false, nil
	}

	if p.Config().PKCE != "force" {
		// autodiscover PKCE support
		pkceSupported, err := discoverPKCE(ctx, d, p)
		if err != nil {
			d.Logger().WithError(err).Warnf("Failed to autodiscover PKCE support for provider %q. Continuing without PKCE.", p.Config().ID)
			return false, nil
		}
		if !pkceSupported {
			d.Logger().Infof("Provider %q does not advertise support for PKCE. Continuing without PKCE.", p.Config().ID)
			return false, nil
		}
	}

	f.EnsureInternalContext()
	bytes, err := sjson.SetBytes(
		f.GetInternalContext(),
		"pkce_verifier",
		oauth2.GenerateVerifier(),
	)
	if err != nil {
		return false, errors.Wrap(err, "failed to store PKCE verifier to internal context")
	}
	f.SetInternalContext(bytes)
	return true, nil
}

func discoverPKCE(ctx context.Context, d pkceDependencies, p OAuth2Provider) (pkceSupported bool, err error) {
	if p.Config().IssuerURL == "" {
		return false, errors.New("Issuer URL must be set to autodiscover PKCE support")
	}

	ctx = gooidc.ClientContext(ctx, d.HTTPClient(ctx).HTTPClient)
	gp, err := gooidc.NewProvider(ctx, p.Config().IssuerURL)
	if err != nil {
		return false, errors.Wrap(err, "failed to initialize provider")
	}
	var claims struct {
		CodeChallengeMethodsSupported []string `json:"code_challenge_methods_supported"`
	}
	if err := gp.Claims(&claims); err != nil {
		return false, errors.Wrap(err, "failed to deserialize provider claims")
	}
	return slices.Contains(claims.CodeChallengeMethodsSupported, "S256"), nil
}

func PKCEChallenge(f flow.InternalContexter) []oauth2.AuthCodeOption {
	if f.GetInternalContext() == nil {
		return nil
	}
	raw := gjson.GetBytes(f.GetInternalContext(), "pkce_verifier")
	if !raw.Exists() {
		return nil
	}
	return []oauth2.AuthCodeOption{oauth2.S256ChallengeOption(raw.String())}
}

func PKCEVerifier(f flow.InternalContexter) []oauth2.AuthCodeOption {
	if f.GetInternalContext() == nil {
		return nil
	}
	raw := gjson.GetBytes(f.GetInternalContext(), "pkce_verifier")
	if !raw.Exists() {
		return nil
	}
	return []oauth2.AuthCodeOption{oauth2.VerifierOption(raw.String())}
}
