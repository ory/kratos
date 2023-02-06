// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package recovery_test

import (
	"context"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/ory/kratos/session"

	"github.com/ory/kratos/selfservice/flow/recovery"
	"github.com/ory/kratos/selfservice/strategy/code"

	"github.com/gobuffalo/httptest"
	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/internal/testhelpers"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/x"
)

func TestRecoveryExecutor(t *testing.T) {
	ctx := context.Background()
	conf, reg := internal.NewFastRegistryWithMocks(t)
	s := code.NewStrategy(reg)

	newServer := func(t *testing.T, i *identity.Identity, ft flow.Type) *httptest.Server {
		router := httprouter.New()
		router.GET("/recovery/pre", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
			a, err := recovery.NewFlow(conf, time.Minute, x.FakeCSRFToken, r, s, ft)
			require.NoError(t, err)
			if testhelpers.SelfServiceHookErrorHandler(t, w, r, recovery.ErrHookAbortFlow, reg.RecoveryExecutor().PreRecoveryHook(w, r, a)) {
				_, _ = w.Write([]byte("ok"))
			}
		})

		router.GET("/recovery/post", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
			a, err := recovery.NewFlow(conf, time.Minute, x.FakeCSRFToken, r, s, ft)
			require.NoError(t, err)
			s, _ := session.NewActiveSession(r,
				i,
				conf,
				time.Now().UTC(),
				identity.CredentialsTypeRecoveryLink,
				identity.AuthenticatorAssuranceLevel1,
			)
			a.RequestURL = x.RequestURL(r).String()
			if testhelpers.SelfServiceHookErrorHandler(t, w, r, recovery.ErrHookAbortFlow, reg.RecoveryExecutor().PostRecoveryHook(w, r, a, s)) {
				_, _ = w.Write([]byte("ok"))
			}
		})

		ts := httptest.NewServer(router)
		t.Cleanup(ts.Close)
		conf.MustSet(ctx, config.ViperKeyPublicBaseURL, ts.URL)
		return ts
	}

	t.Run("method=PostRecoveryHook", func(t *testing.T) {
		t.Run("case=pass without hooks", func(t *testing.T) {
			t.Cleanup(testhelpers.SelfServiceHookConfigReset(t, conf))
			i := testhelpers.SelfServiceHookFakeIdentity(t)
			ts := newServer(t, i, flow.TypeBrowser)

			res, _ := testhelpers.SelfServiceMakeHookRequest(t, ts, "/recovery/post", false, url.Values{})

			assert.EqualValues(t, http.StatusOK, res.StatusCode)
		})

		t.Run("case=pass if hooks pass", func(t *testing.T) {
			t.Cleanup(testhelpers.SelfServiceHookConfigReset(t, conf))
			conf.MustSet(ctx, config.HookStrategyKey(config.ViperKeySelfServiceRecoveryAfter, config.HookGlobal),
				[]config.SelfServiceHook{{Name: "err", Config: []byte(`{}`)}})
			i := testhelpers.SelfServiceHookFakeIdentity(t)
			ts := newServer(t, i, flow.TypeBrowser)

			res, _ := testhelpers.SelfServiceMakeHookRequest(t, ts, "/recovery/post", false, url.Values{})

			assert.EqualValues(t, http.StatusOK, res.StatusCode)
		})

		t.Run("case=fail if hooks fail", func(t *testing.T) {
			t.Cleanup(testhelpers.SelfServiceHookConfigReset(t, conf))
			conf.MustSet(ctx, config.HookStrategyKey(config.ViperKeySelfServiceRecoveryAfter, config.HookGlobal),
				[]config.SelfServiceHook{{Name: "err", Config: []byte(`{"ExecutePostRecoveryHook": "abort"}`)}})
			i := testhelpers.SelfServiceHookFakeIdentity(t)
			ts := newServer(t, i, flow.TypeBrowser)

			res, body := testhelpers.SelfServiceMakeHookRequest(t, ts, "/recovery/post", false, url.Values{})

			assert.EqualValues(t, http.StatusOK, res.StatusCode)
			assert.Equal(t, "", body)
		})
	})

	for _, kind := range []flow.Type{flow.TypeBrowser, flow.TypeAPI} {
		t.Run("type="+string(kind)+"/method=PreRecoveryHook", testhelpers.TestSelfServicePreHook(
			config.ViperKeySelfServiceRecoveryBeforeHooks,
			testhelpers.SelfServiceMakeRecoveryPreHookRequest,
			func(t *testing.T) *httptest.Server {
				return newServer(t, nil, kind)
			},
			conf,
		))
	}
}
