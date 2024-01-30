// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oidc

import (
	"context"
	"fmt"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
)

func verifyToken(ctx context.Context, keySet oidc.KeySet, config *Configuration, rawIDToken, issuerURL string) (*Claims, error) {
	tokenAudiences := append([]string{config.ClientID}, config.AdditionalIDTokenAudiences...)
	var token *oidc.IDToken
	err := fmt.Errorf("no audience matched the token's audience")
	for _, aud := range tokenAudiences {
		verifier := oidc.NewVerifier(issuerURL, keySet, &oidc.Config{
			ClientID: aud,
		})
		token, err = verifier.Verify(ctx, rawIDToken)
		if err != nil && strings.Contains(err.Error(), "oidc: expected audience") {
			// The audience is not the one we expect, try the next one
			continue
		} else if err != nil {
			// Something else went wrong
			return nil, err
		}
		// The token was verified successfully
		break
	}
	if err != nil {
		// None of the allowed audiences matched the audience in the token
		return nil, fmt.Errorf("token audience didn't match allowed audiences: %+v %w", tokenAudiences, err)
	}
	claims := &Claims{}
	if err := token.Claims(claims); err != nil {
		return nil, err
	}
	return claims, nil
}
