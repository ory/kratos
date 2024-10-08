// Copyright © 2024 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oidc

import (
	"context"
	"crypto/sha512"
	"crypto/subtle"

	"github.com/gofrs/uuid"
	"golang.org/x/oauth2"
	"google.golang.org/protobuf/proto"

	"github.com/ory/herodot"
	"github.com/ory/kratos/cipher"
	oidcv1 "github.com/ory/kratos/gen/oidc/v1"
	"github.com/ory/kratos/x"
)

func encryptState(ctx context.Context, c cipher.Cipher, state *oidcv1.State) (ciphertext string, err error) {
	m, err := proto.Marshal(state)
	if err != nil {
		return "", herodot.ErrInternalServerError.WithReasonf("Unable to marshal state: %s", err)
	}
	return c.Encrypt(ctx, m)
}

func DecryptState(ctx context.Context, c cipher.Cipher, ciphertext string) (*oidcv1.State, error) {
	plaintext, err := c.Decrypt(ctx, ciphertext)
	if err != nil {
		return nil, herodot.ErrBadRequest.WithReasonf("Unable to decrypt state: %s", err)
	}
	var state oidcv1.State
	if err := proto.Unmarshal(plaintext, &state); err != nil {
		return nil, herodot.ErrBadRequest.WithReasonf("Unable to unmarshal state: %s", err)
	}
	return &state, nil
}

func (s *Strategy) GenerateState(ctx context.Context, p Provider, flowID uuid.UUID) (stateParam string, pkce []oauth2.AuthCodeOption, err error) {
	state := oidcv1.State{
		FlowId:                         flowID.Bytes(),
		SessionTokenExchangeCodeSha512: x.NewUUID().Bytes(),
		ProviderId:                     p.Config().ID,
		PkceVerifier:                   maybePKCE(ctx, s.d, p),
	}
	if code, hasCode, _ := s.d.SessionTokenExchangePersister().CodeForFlow(ctx, flowID); hasCode {
		sum := sha512.Sum512([]byte(code.InitCode))
		state.SessionTokenExchangeCodeSha512 = sum[:]
	}

	param, err := encryptState(ctx, s.d.Cipher(ctx), &state)
	if err != nil {
		return "", nil, herodot.ErrInternalServerError.WithReason("Unable to encrypt state").WithWrap(err)
	}
	return param, PKCEChallenge(&state), nil
}

func codeMatches(s *oidcv1.State, code string) bool {
	sum := sha512.Sum512([]byte(code))
	return subtle.ConstantTimeCompare(s.GetSessionTokenExchangeCodeSha512(), sum[:]) == 1
}
