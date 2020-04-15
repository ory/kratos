package hook

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/gobuffalo/httptest"
	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/selfservice/flow/registration"
	"github.com/ory/kratos/selfservice/flow/settings"
)

func TestRedirector(t *testing.T) {
	l := NewRedirector(json.RawMessage(""))
	assert.Error(t, l.ExecuteSettingsPrePersistHook(nil, nil, nil, nil))
	assert.Error(t, l.ExecuteLoginPostHook(nil, nil, nil, nil))
	assert.Error(t, l.ExecutePostRegistrationPrePersistHook(nil, nil, nil, nil))
	assert.Error(t, l.ExecuteSettingsPostPersistHook(nil, nil, nil, nil))
	assert.Error(t, l.ExecuteLoginPreHook(nil, nil, nil))
	assert.Error(t, l.ExecuteRegistrationPreHook(nil, nil, nil))

	l = NewRedirector(json.RawMessage(`{"to":"https://www.ory.sh/"}`))
	router := httprouter.New()
	router.GET("/a", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		require.Error(t, l.ExecuteSettingsPrePersistHook(w, r, nil, nil), settings.ErrHookAbortRequest)
	})
	router.GET("/b", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		require.Error(t, l.ExecuteLoginPostHook(w, r, nil, nil), login.ErrHookAbortRequest)
	})
	router.GET("/c", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		require.Error(t, l.ExecutePostRegistrationPrePersistHook(w, r, nil, nil), registration.ErrHookAbortRequest)
	})
	router.GET("/d", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		require.Error(t, l.ExecuteSettingsPostPersistHook(w, r, nil, nil), settings.ErrHookAbortRequest)
	})
	router.GET("/e", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		require.Error(t, l.ExecutePostRegistrationPrePersistHook(w, r, nil, nil), registration.ErrHookAbortRequest)
	})
	router.GET("/f", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		require.Error(t, l.ExecuteLoginPreHook(w, r, nil), login.ErrHookAbortRequest)
	})
	ts := httptest.NewServer(router)
	t.Cleanup(ts.Close)

	for _, p := range []string{"a", "b", "c", "d", "e", "f"} {
		t.Run("route="+p, func(t *testing.T) {
			res, err := http.Get(ts.URL + "/" + p)
			require.NoError(t, err)
			require.NoError(t, res.Body.Close())
			assert.Equal(t, http.StatusOK, res.StatusCode)
			assert.Equal(t, "https://www.ory.sh/", res.Request.URL.String())
		})
	}
}
