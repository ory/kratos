package hook

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/viper"

	"github.com/ory/herodot"
	"github.com/ory/x/urlx"

	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/selfservice/flow/registration"
)

func TestRedirector(t *testing.T) {
	r := http.Request{
		Header: http.Header{},
		URL:    urlx.ParseOrPanic("https://www.ory.sh"),
	}

	h := NewRedirector(
		func() *url.URL {
			return urlx.ParseOrPanic("https://www.ory.sh/fallback")
		},
		func() []url.URL {
			return []url.URL{
				*urlx.ParseOrPanic("https://www.ory.sh"),
				*urlx.ParseOrPanic("https://apis.ory.sh"),
			}
		},
		func() bool {
			return true
		},
	)

	type testCase struct {
		requrl    string
		e         string
		expectErr string
	}

	viper.Set(configuration.ViperKeyDefaultIdentityTraitsSchemaURL, "file://./stub/stub.schema.json")

	var assert = func(t *testing.T, tc testCase, w *httptest.ResponseRecorder, err error) {
		if tc.expectErr != "" {
			require.Error(t, err)
			assert.Contains(t, errors.Cause(err).(*herodot.DefaultError).Reason(), tc.expectErr)
			return
		}
		require.NoError(t, err)
		assert.Equal(t, w.Code, http.StatusFound)
		assert.Equal(t, w.Header().Get("Location"), tc.e)
	}

	for k, tc := range []testCase{
		{requrl: "https://www.ory.sh/?return_to=/foo", e: "https://www.ory.sh/foo"},
		{requrl: "https://login.ory.sh/?return_to=https://not-allowed/foo", e: "https://www.ory.sh/foo", expectErr: "not a whitelisted return domain"},
		{requrl: "https://login.ory.sh/?return_to=https://apis.ory.sh/foo", e: "https://apis.ory.sh/foo"},
		{requrl: "https://www.ory.sh/", e: "https://www.ory.sh/fallback"},
	} {
		t.Run(fmt.Sprintf("method=register/case=%d", k), func(t *testing.T) {
			w := httptest.NewRecorder()
			assert(t, tc, w, h.ExecuteRegistrationPostHook(w, &r, &registration.Request{RequestURL: tc.requrl}, nil))
		})

		t.Run(fmt.Sprintf("method=Login/case=%d", k), func(t *testing.T) {
			w := httptest.NewRecorder()
			assert(t, tc, w, h.ExecuteLoginPostHook(w, &r, &login.Request{RequestURL: tc.requrl}, nil))
		})
	}
}
