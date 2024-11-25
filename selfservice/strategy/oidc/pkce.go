// Copyright © 2024 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oidc

import (
	"context"
	"slices"

	gooidc "github.com/coreos/go-oidc/v3/oidc"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"

	oidcv1 "github.com/ory/kratos/gen/oidc/v1"
	"github.com/ory/kratos/x"
)

type pkceDependencies interface {
	x.LoggingProvider
	x.HTTPClientProvider
}

func PKCEChallenge(s *oidcv1.State) []oauth2.AuthCodeOption {
	if s.GetPkceVerifier() == "" {
		return nil
	}
	return []oauth2.AuthCodeOption{oauth2.S256ChallengeOption(s.GetPkceVerifier())}
}

func PKCEVerifier(s *oidcv1.State) []oauth2.AuthCodeOption {
	if s.GetPkceVerifier() == "" {
		return nil
	}
	return []oauth2.AuthCodeOption{oauth2.VerifierOption(s.GetPkceVerifier())}
}

func maybePKCE(ctx context.Context, d pkceDependencies, _p Provider) (verifier string) {
	if _p.Config().PKCE == "never" {
		return ""
	}

	p, ok := _p.(PKCEEnabledProvider)
	if !ok {
		return ""
	}

	if p.Config().PKCE != "force" {
		// autodiscover PKCE support
		pkceSupported, err := p.PKCEEnabled(ctx)
		if err != nil {
			d.Logger().WithError(err).Warnf("Failed to autodiscover PKCE support for provider %q. Continuing without PKCE.", p.Config().ID)
			return ""
		}
		if !pkceSupported {
			d.Logger().Infof("Provider %q does not advertise support for PKCE. Continuing without PKCE.", p.Config().ID)
			return ""
		}
	}
	return oauth2.GenerateVerifier()
}

func discoverPKCE(p *gooidc.Provider) (pkceSupported bool, err error) {
	var claims struct {
		CodeChallengeMethodsSupported []string `json:"code_challenge_methods_supported"`
	}
	if err := p.Claims(&claims); err != nil {
		return false, errors.Wrap(err, "failed to deserialize provider claims")
	}
	return slices.Contains(claims.CodeChallengeMethodsSupported, "S256"), nil
}
