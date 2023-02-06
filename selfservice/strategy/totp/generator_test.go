// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package totp_test

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/selfservice/strategy/totp"
)

func TestGenerator(t *testing.T) {
	ctx := context.Background()
	conf, reg := internal.NewFastRegistryWithMocks(t)

	key, err := totp.NewKey(context.Background(), "foo", reg)
	require.NoError(t, err)
	assert.Equal(t, conf.SelfPublicURL(ctx).Hostname(), key.Issuer(), "if issuer is not set explicitly it should be the public URL")

	require.NoError(t, conf.Set(ctx, config.ViperKeyTOTPIssuer, "foobar.com"))

	key, err = totp.NewKey(context.Background(), "foo", reg)
	require.NoError(t, err)
	assert.Equal(t, "foobar.com", key.Issuer(), "if issuer is set explicitly it should be the correct value")
	assert.Equal(t, "foo", key.AccountName())

	img, err := totp.KeyToHTMLImage(key)
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(img, "data:image/png;base64,"), "image is a base64 encoded png")
}
