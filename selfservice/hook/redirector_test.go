package hook

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/gobuffalo/httptest"
	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/selfservice/flow/registration"
	"github.com/ory/kratos/selfservice/flow/settings"
)

func TestRedirector(t *testing.T) {
	l := NewRedirector(json.RawMessage(""))
	assert.Error(t, l.ExecuteSettingsPrePersistHook(nil, nil, &settings.Request{Type: flow.TypeBrowser}, nil))
	assert.Error(t, l.ExecuteLoginPostHook(nil, nil, &login.Flow{Type: flow.TypeBrowser}, nil))
	assert.Error(t, l.ExecutePostRegistrationPrePersistHook(nil, nil,  &registration.Flow{Type: flow.TypeBrowser}, nil))
	assert.Error(t, l.ExecuteSettingsPostPersistHook(nil, nil,  &settings.Request{Type: flow.TypeBrowser}, nil))
	assert.Error(t, l.ExecuteLoginPreHook(nil, nil,  &login.Flow{Type: flow.TypeBrowser}))
	assert.Error(t, l.ExecuteRegistrationPreHook(nil, nil,  &registration.Flow{Type: flow.TypeBrowser}))

	assert.NoError(t, l.ExecuteSettingsPrePersistHook(nil, nil, &settings.Request{Type: flow.TypeAPI}, nil))
	assert.NoError(t, l.ExecuteLoginPostHook(nil, nil, &login.Flow{Type: flow.TypeAPI}, nil))
	assert.NoError(t, l.ExecutePostRegistrationPrePersistHook(nil, nil,  &registration.Flow{Type: flow.TypeAPI}, nil))
	assert.NoError(t, l.ExecuteSettingsPostPersistHook(nil, nil,  &settings.Request{Type: flow.TypeAPI}, nil))
	assert.NoError(t, l.ExecuteLoginPreHook(nil, nil,  &login.Flow{Type: flow.TypeAPI}))
	assert.NoError(t, l.ExecuteRegistrationPreHook(nil, nil,  &registration.Flow{Type: flow.TypeAPI}))

	l = NewRedirector(json.RawMessage(`{"to":"https://www.ory.sh/"}`))
	router := httprouter.New()
	router.GET("/a", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		require.Error(t, l.ExecuteSettingsPrePersistHook(w, r, &settings.Request{Type: flow.TypeAPI}, nil), settings.ErrHookAbortRequest)
	})
	router.GET("/b", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		require.Error(t, l.ExecuteLoginPostHook(w, r, &login.Flow{Type: flow.TypeAPI}, nil), login.ErrHookAbortFlow)
	})
	router.GET("/c", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		require.Error(t, l.ExecutePostRegistrationPrePersistHook(w, r, &registration.Flow{Type: flow.TypeAPI}, nil), registration.ErrHookAbortFlow)
	})
	router.GET("/d", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		require.Error(t, l.ExecuteSettingsPostPersistHook(w, r, &settings.Request{Type: flow.TypeAPI}, nil), settings.ErrHookAbortRequest)
	})
	router.GET("/e", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		require.Error(t, l.ExecutePostRegistrationPrePersistHook(w, r, &registration.Flow{Type: flow.TypeAPI}, nil), registration.ErrHookAbortFlow)
	})
	router.GET("/f", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		require.Error(t, l.ExecuteLoginPreHook(w, r, &login.Flow{Type: flow.TypeAPI}), login.ErrHookAbortFlow)
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
