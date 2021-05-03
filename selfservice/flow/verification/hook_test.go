package verification_test

import (
	"context"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/ory/kratos/selfservice/flow/verification"

	"github.com/gobuffalo/httptest"
	"github.com/julienschmidt/httprouter"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/internal/testhelpers"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/x"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVerificationExecutor(t *testing.T) {
	conf, reg := internal.NewFastRegistryWithMocks(t)

	newServer := func(t *testing.T, i *identity.Identity, ft flow.Type) *httptest.Server {
		router := httprouter.New()

		router.GET("/verification/post", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
			a, err := verification.NewFlow(conf, time.Minute, x.FakeCSRFToken, r, reg.VerificationStrategies(context.Background()), ft)
			require.NoError(t, err)
			a.RequestURL = x.RequestURL(r).String()
			if testhelpers.SelfServiceHookErrorHandler(t, w, r, verification.ErrHookAbortFlow, reg.VerificationExecutor().PostVerificationHook(w, r, a, i)) {
				_, _ = w.Write([]byte("ok"))
			}
		})

		ts := httptest.NewServer(router)
		t.Cleanup(ts.Close)
		conf.MustSet(config.ViperKeyPublicBaseURL, ts.URL)
		return ts
	}

	t.Run("method=PostVerificationHook", func(t *testing.T) {
		t.Run("case=pass without hooks", func(t *testing.T) {
			t.Cleanup(testhelpers.SelfServiceHookConfigReset(t, conf))
			i := testhelpers.SelfServiceHookFakeIdentity(t)
			ts := newServer(t, i, flow.TypeBrowser)

			res, _ := testhelpers.SelfServiceMakeHookRequest(t, ts, "/verification/post", false, url.Values{})

			assert.EqualValues(t, http.StatusOK, res.StatusCode)
		})

		t.Run("case=pass if hooks pass", func(t *testing.T) {
			t.Cleanup(testhelpers.SelfServiceHookConfigReset(t, conf))
			conf.MustSet(config.HookStrategyKey(config.ViperKeySelfServiceVerificationAfter, config.HookGlobal),
				[]config.SelfServiceHook{{Name: "err", Config: []byte(`{}`)}})
			i := testhelpers.SelfServiceHookFakeIdentity(t)
			ts := newServer(t, i, flow.TypeBrowser)

			res, _ := testhelpers.SelfServiceMakeHookRequest(t, ts, "/verification/post", false, url.Values{})

			assert.EqualValues(t, http.StatusOK, res.StatusCode)
		})

		t.Run("case=fail if hooks fail", func(t *testing.T) {
			t.Cleanup(testhelpers.SelfServiceHookConfigReset(t, conf))
			conf.MustSet(config.HookStrategyKey(config.ViperKeySelfServiceVerificationAfter, config.HookGlobal),
				[]config.SelfServiceHook{{Name: "err", Config: []byte(`{"ExecutePostVerificationHook": "abort"}`)}})
			i := testhelpers.SelfServiceHookFakeIdentity(t)
			ts := newServer(t, i, flow.TypeBrowser)

			res, body := testhelpers.SelfServiceMakeHookRequest(t, ts, "/verification/post", false, url.Values{})

			assert.EqualValues(t, http.StatusOK, res.StatusCode)
			assert.Equal(t, "", body)
		})
	})
}
