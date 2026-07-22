// Copyright © 2026 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package settings_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/pkg"
	"github.com/ory/kratos/pkg/testhelpers"
)

func TestWellKnownChangePassword(t *testing.T) {
	ctx := t.Context()
	conf, reg := pkg.NewFastRegistryWithMocks(t)
	testhelpers.SetDefaultIdentitySchema(conf, "file://./stub/identity.schema.json")
	conf.MustSet(ctx, config.ViperKeySelfServiceSettingsURL, "https://ui.example.com/settings")

	public, _ := testhelpers.NewKratosServer(t, reg)

	hc := &http.Client{Transport: testhelpers.NewTestTransport(t), CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	res, err := hc.Get(public.URL + "/.well-known/change-password")
	require.NoError(t, err)
	t.Cleanup(func() { _ = res.Body.Close() })

	assert.Equal(t, http.StatusSeeOther, res.StatusCode)
	assert.Equal(t, "https://ui.example.com/settings", res.Header.Get("Location"))
}
