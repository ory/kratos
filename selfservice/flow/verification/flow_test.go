package verification_test

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/tidwall/gjson"

	"github.com/ory/x/jsonx"

	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/selfservice/flow/verification"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/x/urlx"

	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/registration"
)

func TestFlow(t *testing.T) {
	conf, _ := internal.NewFastRegistryWithMocks(t)

	must := func(r *verification.Flow, err error) *verification.Flow {
		require.NoError(t, err)
		return r
	}

	u := &http.Request{URL: urlx.ParseOrPanic("http://foo/bar/baz"), Host: "foo"}
	for k, tc := range []struct {
		r         *verification.Flow
		expectErr bool
	}{
		{r: must(verification.NewFlow(conf, time.Hour, "", u, nil, flow.TypeBrowser))},
		{r: must(verification.NewFlow(conf, -time.Hour, "", u, nil, flow.TypeBrowser)), expectErr: true},
	} {
		t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
			err := tc.r.Valid()
			if tc.expectErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}

	t.Run("type=return_to", func(t *testing.T) {
		_, err := verification.NewFlow(conf, 0, "csrf", &http.Request{URL: &url.URL{Path: "/", RawQuery: "return_to=https://not-allowed/foobar"}, Host: "ory.sh"}, nil, flow.TypeBrowser)
		require.Error(t, err)

		_, err = verification.NewFlow(conf, 0, "csrf", &http.Request{URL: &url.URL{Path: "/", RawQuery: "return_to=" + urlx.AppendPaths(conf.SelfPublicURL(), "/self-service/login/browser").String()}, Host: "ory.sh"}, nil, flow.TypeBrowser)
		require.NoError(t, err)
	})

	assert.EqualValues(t, verification.StateChooseMethod,
		must(verification.NewFlow(conf, time.Hour, "", u, nil, flow.TypeBrowser)).State)
}

func TestGetType(t *testing.T) {
	for _, ft := range []flow.Type{
		flow.TypeAPI,
		flow.TypeBrowser,
	} {
		t.Run(fmt.Sprintf("case=%s", ft), func(t *testing.T) {
			r := &verification.Flow{Type: ft}
			assert.Equal(t, ft, r.GetType())
		})
	}
}

func TestGetRequestURL(t *testing.T) {
	expectedURL := "http://foo/bar/baz"
	f := &verification.Flow{RequestURL: expectedURL}
	assert.Equal(t, expectedURL, f.GetRequestURL())
}

func TestNewPostHookFlow(t *testing.T) {
	conf := internal.NewConfigurationWithDefaults(t)
	u := &http.Request{URL: urlx.ParseOrPanic("http://foo/bar/baz"), Host: "foo"}
	expectReturnTo := func(t *testing.T, originalFlowRequestQueryParams url.Values, expectedReturnTo string) {
		originalFlow := registration.Flow{
			RequestURL: "http://foo.com/bar?" + originalFlowRequestQueryParams.Encode(),
		}
		t.Log(originalFlow.RequestURL)
		f, err := verification.NewPostHookFlow(conf, time.Second, "", u, nil, &originalFlow)
		require.NoError(t, err)
		u, err := urlx.Parse(f.RequestURL)
		require.NoError(t, err)
		assert.Equal(t, "", u.Query().Get("after_verification_return_to"))
		assert.Equal(t, expectedReturnTo, u.Query().Get("return_to"))
	}

	t.Run("case=after_verification_return_to supplied", func(t *testing.T) {
		expectedReturnTo := "http://foo.com/verification_callback"
		expectReturnTo(t, url.Values{"after_verification_return_to": {expectedReturnTo}}, expectedReturnTo)
	})

	t.Run("case=return_to supplied", func(t *testing.T) {
		expectReturnTo(t, url.Values{"return_to": {"http://foo.com/original_flow_callback"}}, "")
	})

	t.Run("case=return_to and after_verification_return_to supplied", func(t *testing.T) {
		expectedReturnTo := "http://foo.com/verification_callback"
		expectReturnTo(t, url.Values{
			"return_to":                    {"http://foo.com/original_flow_callback"},
			"after_verification_return_to": {expectedReturnTo},
		}, expectedReturnTo)
	})
}

func TestFlowEncodeJSON(t *testing.T) {
	assert.EqualValues(t, "", gjson.Get(jsonx.TestMarshalJSONString(t, &verification.Flow{RequestURL: "https://foo.bar?foo=bar"}), "return_to").String())
	assert.EqualValues(t, "/bar", gjson.Get(jsonx.TestMarshalJSONString(t, &verification.Flow{RequestURL: "https://foo.bar?return_to=/bar"}), "return_to").String())
	assert.EqualValues(t, "/bar", gjson.Get(jsonx.TestMarshalJSONString(t, verification.Flow{RequestURL: "https://foo.bar?return_to=/bar"}), "return_to").String())
}
