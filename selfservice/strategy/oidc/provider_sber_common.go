// Copyright © 2026 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oidc

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"

	"github.com/ory/herodot"
)

const sberTokenDebugVersion = "sber-token-debug-v3"

func sberNonceFromRequest(providerID string, req ider) string {
	if req == nil {
		return ""
	}
	return sberNonceFromFlowID(providerID, req.GetID())
}

func sberNonceFromFlowID(providerID string, flowID uuid.UUID) string {
	sum := sha256.Sum256([]byte(providerID + ":" + flowID.String()))
	nonce := hex.EncodeToString(sum[:])
	if len(nonce) > 32 {
		return nonce[:32]
	}
	return nonce
}

func validateSberIDToken(token *oauth2.Token, providerID, clientID string, flowID uuid.UUID) error {
	if flowID == uuid.Nil {
		return nil
	}

	if token == nil {
		return errors.WithStack(herodot.ErrUpstreamError.WithReasonf("sber id_token validation failed: token is missing"))
	}

	rawIDToken, _ := token.Extra("id_token").(string)
	if rawIDToken == "" {
		return errors.WithStack(herodot.ErrUpstreamError.WithReasonf("sber id_token validation failed: id_token is missing"))
	}

	claims, err := decodeJWTClaims(rawIDToken)
	if err != nil {
		return errors.WithStack(herodot.ErrUpstreamError.WithWrap(err).WithReasonf("sber id_token validation failed: unable to parse id_token"))
	}

	aud, err := claimAudience(claims["aud"])
	if err != nil {
		return errors.WithStack(herodot.ErrUpstreamError.WithWrap(err).WithReasonf("sber id_token validation failed: invalid aud claim"))
	}
	if aud != clientID {
		return errors.WithStack(herodot.ErrUpstreamError.WithReasonf("sber id_token validation failed: aud mismatch for provider=%s", providerID))
	}

	now := time.Now().Unix()
	iat, err := claimUnixTime(claims["iat"])
	if err != nil {
		return errors.WithStack(herodot.ErrUpstreamError.WithWrap(err).WithReasonf("sber id_token validation failed: invalid iat claim"))
	}
	exp, err := claimUnixTime(claims["exp"])
	if err != nil {
		return errors.WithStack(herodot.ErrUpstreamError.WithWrap(err).WithReasonf("sber id_token validation failed: invalid exp claim"))
	}
	if exp <= now {
		return errors.WithStack(herodot.ErrUpstreamError.WithReasonf("sber id_token validation failed: id_token expired"))
	}
	if iat > now+60 {
		return errors.WithStack(herodot.ErrUpstreamError.WithReasonf("sber id_token validation failed: iat is in the future"))
	}

	expectedNonce := sberNonceFromFlowID(providerID, flowID)
	nonce, _ := claims["nonce"].(string)
	if expectedNonce == "" || nonce == "" || nonce != expectedNonce {
		return errors.WithStack(herodot.ErrUpstreamError.WithReasonf("sber id_token validation failed: nonce mismatch for provider=%s", providerID))
	}

	return nil
}

func decodeJWTClaims(rawToken string) (map[string]interface{}, error) {
	parts := strings.Split(rawToken, ".")
	if len(parts) < 2 {
		return nil, errors.New("jwt has invalid format")
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, errors.WithStack(err)
	}

	var claims map[string]interface{}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, errors.WithStack(err)
	}

	return claims, nil
}

func claimAudience(v interface{}) (string, error) {
	switch aud := v.(type) {
	case string:
		if aud == "" {
			return "", errors.New("aud is empty")
		}
		return aud, nil
	case []interface{}:
		if len(aud) == 0 {
			return "", errors.New("aud array is empty")
		}
		audStr, ok := aud[0].(string)
		if !ok || audStr == "" {
			return "", errors.New("aud[0] is invalid")
		}
		return audStr, nil
	default:
		return "", errors.New("aud has unexpected type")
	}
}

func claimUnixTime(v interface{}) (int64, error) {
	switch t := v.(type) {
	case float64:
		if math.IsNaN(t) || math.IsInf(t, 0) {
			return 0, errors.New("invalid numeric time")
		}
		return int64(t), nil
	case json.Number:
		return t.Int64()
	case string:
		if t == "" {
			return 0, errors.New("empty time string")
		}
		return strconv.ParseInt(t, 10, 64)
	default:
		return 0, fmt.Errorf("unexpected time type: %T", v)
	}
}

func sberAllSubjects(claims *Claims) []string {
	if claims == nil {
		return nil
	}

	subjects := make([]string, 0, 2)
	if claims.Subject != "" {
		subjects = append(subjects, claims.Subject)
	}

	if claims.RawClaims == nil {
		return subjects
	}

	subAlt, ok := claims.RawClaims["sub_alt"]
	if !ok {
		return subjects
	}

	appendUnique := func(v string) {
		if v == "" {
			return
		}
		for _, s := range subjects {
			if s == v {
				return
			}
		}
		subjects = append(subjects, v)
	}

	switch raw := subAlt.(type) {
	case string:
		appendUnique(raw)
	case []interface{}:
		for _, item := range raw {
			if s, ok := item.(string); ok {
				appendUnique(s)
			}
		}
	}

	return subjects
}

type sberFlowIDContextKey struct{}

func withSberFlowID(ctx context.Context, flowID uuid.UUID) context.Context {
	if flowID == uuid.Nil {
		return ctx
	}
	return context.WithValue(ctx, sberFlowIDContextKey{}, flowID)
}

func sberFlowIDFromContext(ctx context.Context) uuid.UUID {
	if ctx == nil {
		return uuid.Nil
	}
	if flowID, ok := ctx.Value(sberFlowIDContextKey{}).(uuid.UUID); ok {
		return flowID
	}
	return uuid.Nil
}

func isSberProviderID(providerID string) bool {
	return providerID == "sber" || providerID == "sber-ift"
}

func sberAuthCompletedURL(providerID string) string {
	if providerID == "sber-ift" {
		return "https://oauth-ift.sber.ru/api/v2/auth/completed"
	}
	return "https://oauth.sber.ru/api/v2/auth/completed"
}
