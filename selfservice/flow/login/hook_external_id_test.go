// Copyright © 2026 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package login_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/hydra"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/internal/testhelpers"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/session"
	"github.com/ory/x/sqlxx"
)

func TestLoginExecutorWithExternalID(t *testing.T) {
	ctx := context.Background()
	conf, reg := internal.NewFastRegistryWithMocks(t)
	fakeHydra := hydra.NewFake()
	reg.SetHydra(fakeHydra)

	testhelpers.SetDefaultIdentitySchema(conf, "file://./stub/login.schema.json")
	conf.MustSet(ctx, config.ViperKeySelfServiceBrowserDefaultReturnTo, "https://www.ory.sh/kratos/return_to")
	conf.MustSet(ctx, config.ViperKeyOAuth2ProviderURL, "https://hydra.example.com")

	i := &identity.Identity{
		ID:         uuid.Must(uuid.NewV4()),
		ExternalID: sqlxx.NullString("external-id"),
		SchemaID:   config.DefaultIdentityTraitsSchemaID,
		State:      identity.StateActive,
	}
	require.NoError(t, reg.Persister().CreateIdentity(ctx, i))

	t.Run("case=use_external_id=false", func(t *testing.T) {
		conf.MustSet(ctx, config.ViperKeyOAuth2ProviderUseExternalID, false)
		loginFlow, err := login.NewFlow(conf, time.Minute, hydra.FakeValidLoginChallenge, &http.Request{URL: &url.URL{Path: "/", RawQuery: "login_challenge=" + hydra.FakeValidLoginChallenge}}, flow.TypeBrowser)
		require.NoError(t, err)
		loginFlow.OAuth2LoginChallenge = hydra.FakeValidLoginChallenge

		w := httptest.NewRecorder()
		r := &http.Request{URL: &url.URL{Path: "/login/post"}}
		sess := session.NewInactiveSession()
		sess.CompletedLoginFor(identity.CredentialsTypePassword, identity.AuthenticatorAssuranceLevel1)

		err = reg.LoginHookExecutor().PostLoginHook(w, r, identity.CredentialsTypePassword.ToUiNodeGroup(), loginFlow, i, sess, "")
		require.NoError(t, err)

		require.Len(t, fakeHydra.Params(), 1)
		assert.Equal(t, i.ID.String(), fakeHydra.Params()[0].IdentityID)
		assert.Equal(t, "external-id", fakeHydra.Params()[0].ExternalID)
	})

	t.Run("case=use_external_id=true", func(t *testing.T) {
		conf.MustSet(ctx, config.ViperKeyOAuth2ProviderUseExternalID, true)
		loginFlow, err := login.NewFlow(conf, time.Minute, hydra.FakeValidLoginChallenge, &http.Request{URL: &url.URL{Path: "/", RawQuery: "login_challenge=" + hydra.FakeValidLoginChallenge}}, flow.TypeBrowser)
		require.NoError(t, err)
		loginFlow.OAuth2LoginChallenge = hydra.FakeValidLoginChallenge

		w := httptest.NewRecorder()
		r := &http.Request{URL: &url.URL{Path: "/login/post"}}
		sess := session.NewInactiveSession()
		sess.CompletedLoginFor(identity.CredentialsTypePassword, identity.AuthenticatorAssuranceLevel1)

		fakeHydra.Params()

		err = reg.LoginHookExecutor().PostLoginHook(w, r, identity.CredentialsTypePassword.ToUiNodeGroup(), loginFlow, i, sess, "")
		require.NoError(t, err)

		params := fakeHydra.Params()
		require.NotEmpty(t, params)
		lastParams := params[len(params)-1]
		assert.Equal(t, i.ID.String(), lastParams.IdentityID)
		assert.Equal(t, "external-id", lastParams.ExternalID)
	})
}
