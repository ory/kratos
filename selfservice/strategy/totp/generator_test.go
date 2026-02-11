// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package totp_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/x/contextx"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/pkg"
	"github.com/ory/kratos/selfservice/strategy/totp"
)

func TestGenerator(t *testing.T) {
	_, reg := pkg.NewFastRegistryWithMocks(t)

	t.Run("no issuer set", func(t *testing.T) {
		key, err := totp.NewKey(t.Context(), "foo", reg)
		require.NoError(t, err)
		assert.Equal(t, reg.Config().SelfPublicURL(t.Context()).Hostname(), key.Issuer(), "if issuer is not set explicitly it should be the public URL")
	})

	t.Run("custom issuer set", func(t *testing.T) {
		ctx := contextx.WithConfigValue(t.Context(), config.ViperKeyTOTPIssuer, "foobar.com")

		key, err := totp.NewKey(ctx, "foo", reg)
		require.NoError(t, err)
		assert.Equal(t, "foobar.com", key.Issuer(), "if issuer is set explicitly it should be the correct value")
		assert.Equal(t, "foo", key.AccountName())
	})

	t.Run("generate HTML image", func(t *testing.T) {
		key, err := totp.NewKey(t.Context(), "foo", reg)
		require.NoError(t, err)

		img, err := totp.KeyToHTMLImage(key)
		require.NoError(t, err)
		assert.Truef(t, strings.HasPrefix(img, "data:image/png;base64,"), "image is a base64 encoded png: %s", img)
	})
}
