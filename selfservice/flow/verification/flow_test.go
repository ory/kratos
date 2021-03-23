package verification

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/x/urlx"

	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/registration"
)

func TestFlow(t *testing.T) {
	must := func(r *Flow, err error) *Flow {
		require.NoError(t, err)
		return r
	}

	u := &http.Request{URL: urlx.ParseOrPanic("http://foo/bar/baz"), Host: "foo"}
	for k, tc := range []struct {
		r         *Flow
		expectErr bool
	}{
		{r: must(NewFlow(time.Hour, "", u, nil, flow.TypeBrowser))},
		{r: must(NewFlow(-time.Hour, "", u, nil, flow.TypeBrowser)), expectErr: true},
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

	assert.EqualValues(t, StateChooseMethod,
		must(NewFlow(time.Hour, "", u, nil, flow.TypeBrowser)).State)
}

func TestGetType(t *testing.T) {
	for _, ft := range []flow.Type{
		flow.TypeAPI,
		flow.TypeBrowser,
	} {
		t.Run(fmt.Sprintf("case=%s", ft), func(t *testing.T) {
			r := &Flow{Type: ft}
			assert.Equal(t, ft, r.GetType())
		})
	}
}

func TestGetRequestURL(t *testing.T) {
	expectedURL := "http://foo/bar/baz"
	f := &Flow{RequestURL: expectedURL}
	assert.Equal(t, expectedURL, f.GetRequestURL())
}

func TestNewPostHookFlow(t *testing.T) {
	expectReturnTo := func(t *testing.T, originalFlowRequestQueryParams url.Values, expectedReturnTo string) {
		originalFlow := registration.Flow{
			RequestURL: "http://foo.com/bar?" + originalFlowRequestQueryParams.Encode(),
		}
		t.Log(originalFlow.RequestURL)
		f, err := NewPostHookFlow(time.Second, &originalFlow)
		require.NoError(t, err)
		url, err := urlx.Parse(f.RequestURL)
		require.NoError(t, err)
		assert.Equal(t, "", url.Query().Get("after_verification_return_to"))
		assert.Equal(t, expectedReturnTo, url.Query().Get("return_to"))
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
