// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package session_test

import (
	"context"
	_ "embed"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ory/herodot"

	"github.com/gofrs/uuid"
	"github.com/golang-jwt/jwt/v5"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/session"
	"github.com/ory/x/snapshotx"
)

//go:embed stub/jwk.es256.json
var es256Key []byte

//go:embed stub/jwk.es512.json
var es512Key []byte

func validateTokenized(t *testing.T, raw string, key []byte) *jwt.Token {
	token, err := jwt.Parse(
		raw,
		func(token *jwt.Token) (target interface{}, _ error) {
			set, err := jwk.Parse(key)
			if err != nil {
				return nil, err
			}
			key, _ := set.Key(0)
			if pk, err := key.PublicKey(); err != nil {
				return nil, err
			} else if err := pk.Raw(&target); err != nil {
				return nil, err
			}
			return target, nil
		},
		// We use a fixed time function for snapshot testing, and thus can not validate claims.
		jwt.WithoutClaimsValidation(),
	)
	require.NoError(t, err)
	return token
}

func setTokenizeConfig(conf *config.Config, templateID string, keyFile string, mapper string) {
	conf.MustSet(context.Background(), config.ViperKeySessionTokenizerTemplates+"."+templateID, &config.SessionTokenizeFormat{
		TTL:             time.Minute,
		JWKSURL:         "file://stub/" + keyFile,
		ClaimsMapperURL: mapper,
	})
}

func TestTokenizer(t *testing.T) {
	ctx := context.Background()
	now := time.Now()

	conf, reg := internal.NewFastRegistryWithMocks(t)
	conf.MustSet(ctx, config.ViperKeyPublicBaseURL, "http://localhost/")
	tkn := session.NewTokenizer(reg)
	nowDate := time.Date(2023, 02, 01, 00, 00, 00, 0, time.UTC)
	tkn.SetNowFunc(func() time.Time {
		return nowDate
	})

	r := httptest.NewRequest("GET", "/sessions/whoami", nil)
	i := identity.NewIdentity("default")
	i.ID = uuid.FromStringOrNil("7458af86-c1d8-401c-978a-8da89133f78b")

	s, err := session.NewActiveSession(r, i, conf, now, identity.CredentialsTypePassword, identity.AuthenticatorAssuranceLevel1)
	require.NoError(t, err)
	s.ID = uuid.FromStringOrNil("432caf86-c1d8-401c-978a-8da89133f78b")

	t.Run("case=es256-without-jsonnet", func(t *testing.T) {
		tid := "es256-no-template"
		setTokenizeConfig(conf, tid, "jwk.es256.json", "")

		require.NoError(t, tkn.TokenizeSession(ctx, tid, s))
		token := validateTokenized(t, s.Tokenized, es256Key)

		resultClaims := token.Claims.(jwt.MapClaims)
		assert.Equal(t, i.ID.String(), resultClaims["sub"])
		assert.Equal(t, s.ID.String(), resultClaims["sid"])
		assert.NotEmpty(t, resultClaims["jti"])
		assert.EqualValues(t, resultClaims["exp"], nowDate.Add(time.Minute).Unix())

		snapshotx.SnapshotT(t, token.Claims, snapshotx.ExceptPaths("jti"))
	})

	t.Run("case=es512-without-jsonnet", func(t *testing.T) {
		tid := "es512-no-template"
		setTokenizeConfig(conf, tid, "jwk.es512.json", "")

		require.NoError(t, tkn.TokenizeSession(ctx, tid, s))
		token := validateTokenized(t, s.Tokenized, es512Key)

		snapshotx.SnapshotT(t, token.Claims, snapshotx.ExceptPaths("jti"))
	})

	t.Run("case=rs512-with-jsonnet", func(t *testing.T) {
		tid := "rs512-template"
		setTokenizeConfig(conf, tid, "jwk.es512.json", "file://stub/rs512-template.jsonnet")

		require.NoError(t, tkn.TokenizeSession(ctx, tid, s))
		token := validateTokenized(t, s.Tokenized, es512Key)

		snapshotx.SnapshotT(t, token.Claims, snapshotx.ExceptPaths("jti"))
	})

	t.Run("case=rs512-with-broken-keyfile", func(t *testing.T) {
		tid := "rs512-template"
		setTokenizeConfig(conf, tid, "jwk.es512.broken.json", "file://stub/rs512-template.jsonnet")
		err := tkn.TokenizeSession(ctx, tid, s)
		require.ErrorIs(t, err, herodot.ErrBadRequest)
	})
}
