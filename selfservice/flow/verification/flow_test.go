// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package verification_test

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/tidwall/gjson"

	"github.com/ory/x/jsonx"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/pkg"
	"github.com/ory/kratos/selfservice/flow/verification"
	"github.com/ory/kratos/x"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/x/urlx"

	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/registration"
)

func TestFlow(t *testing.T) {
	ctx := context.Background()
	conf, _ := pkg.NewFastRegistryWithMocks(t)

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

		_, err = verification.NewFlow(conf, 0, "csrf", &http.Request{URL: &url.URL{Path: "/", RawQuery: "return_to=" + urlx.AppendPaths(conf.SelfPublicURL(ctx), "/self-service/login/browser").String()}, Host: "ory.sh"}, nil, flow.TypeBrowser)
		require.NoError(t, err)
	})

	assert.EqualValues(t, flow.StateChooseMethod,
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

func TestNewFlow_capturesCourierBaseURL(t *testing.T) {
	conf := pkg.NewConfigurationWithDefaults(t)

	t.Run("nothing in context", func(t *testing.T) {
		r := &http.Request{URL: urlx.ParseOrPanic("http://foo/bar"), Host: "foo"}
		f, err := verification.NewFlow(conf, time.Hour, "", r, nil, flow.TypeBrowser)
		require.NoError(t, err)
		assert.Equal(t, "", f.GetCourierBaseURL())
	})

	t.Run("BaseURL in request context is copied onto flow", func(t *testing.T) {
		base := urlx.ParseOrPanic("https://customer.example.com/")
		r := (&http.Request{URL: urlx.ParseOrPanic("http://foo/bar"), Host: "foo"}).
			WithContext(x.WithBaseURL(context.Background(), base))
		f, err := verification.NewFlow(conf, time.Hour, "", r, nil, flow.TypeBrowser)
		require.NoError(t, err)
		assert.Equal(t, base.String(), f.GetCourierBaseURL())
	})

	t.Run("http scheme is preserved (not coerced to https)", func(t *testing.T) {
		base := urlx.ParseOrPanic("http://localhost:4000")
		r := (&http.Request{URL: urlx.ParseOrPanic("http://foo/bar"), Host: "foo"}).
			WithContext(x.WithBaseURL(context.Background(), base))
		f, err := verification.NewFlow(conf, time.Hour, "", r, nil, flow.TypeBrowser)
		require.NoError(t, err)
		assert.Equal(t, "http://localhost:4000", f.GetCourierBaseURL())
	})
}

func TestNewPostHookFlow(t *testing.T) {
	conf := pkg.NewConfigurationWithDefaults(t)
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
		assert.True(t, strings.HasPrefix(f.RequestURL, "http://foo.com/bar?"))
		assert.Equal(t, "", u.Query().Get("after_verification_return_to"))
		assert.Equal(t, expectedReturnTo, u.Query().Get("return_to"))
	}

	t.Run("case=after_verification_return_to supplied", func(t *testing.T) {
		expectedReturnTo := "http://foo.com/verification_callback"
		expectReturnTo(t, url.Values{"after_verification_return_to": {expectedReturnTo}}, expectedReturnTo)
	})

	t.Run("case=return_to supplied", func(t *testing.T) {
		expectReturnTo(t, url.Values{"return_to": {"http://foo.com/original_flow_callback"}}, "http://foo.com/original_flow_callback")
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

func TestFromOldFlow_capturesCourierBaseURL(t *testing.T) {
	conf := pkg.NewConfigurationWithDefaults(t)

	t.Run("uses the new request's context, not the old flow's value", func(t *testing.T) {
		oldR := (&http.Request{URL: urlx.ParseOrPanic("http://foo/bar"), Host: "foo"}).
			WithContext(x.WithBaseURL(context.Background(), urlx.ParseOrPanic("https://customer.example.com/")))
		oldF, err := verification.NewFlow(conf, time.Hour, "csrf", oldR, nil, flow.TypeBrowser)
		require.NoError(t, err)
		require.Equal(t, "https://customer.example.com/", oldF.GetCourierBaseURL())

		newR := (&http.Request{URL: urlx.ParseOrPanic("http://foo/bar"), Host: "foo"}).
			WithContext(x.WithBaseURL(context.Background(), urlx.ParseOrPanic("http://localhost:4000")))
		newF, err := verification.FromOldFlow(conf, time.Hour, oldF.CSRFToken, newR, nil, oldF)
		require.NoError(t, err)
		assert.Equal(t, "http://localhost:4000", newF.GetCourierBaseURL())
	})

	t.Run("empty when neither old nor new have a base URL", func(t *testing.T) {
		r := &http.Request{URL: urlx.ParseOrPanic("http://foo/bar"), Host: "foo"}
		oldF, err := verification.NewFlow(conf, time.Hour, "csrf", r, nil, flow.TypeBrowser)
		require.NoError(t, err)
		newF, err := verification.FromOldFlow(conf, time.Hour, oldF.CSRFToken, r, nil, oldF)
		require.NoError(t, err)
		assert.Equal(t, "", newF.GetCourierBaseURL())
	})
}

func TestNewPostHookFlow_capturesCourierBaseURL(t *testing.T) {
	conf := pkg.NewConfigurationWithDefaults(t)

	// NewPostHookFlow is what the registration → auto-verification chain
	// calls during the registration submit handler. The verification flow
	// it creates needs CourierBaseURL captured from the *triggering*
	// request's context (the registration submit, with the CNAME-aware
	// X-Ory-Original-Host set by the cloud middleware) — not from the
	// original flow's RequestURL.
	t.Run("captures from triggering request context, ignoring original flow's URL", func(t *testing.T) {
		original := registration.Flow{
			RequestURL: "https://" + "different.host/bar?return_to=https://elsewhere/path",
		}
		r := (&http.Request{URL: urlx.ParseOrPanic("http://foo/bar"), Host: "foo"}).
			WithContext(x.WithBaseURL(context.Background(), urlx.ParseOrPanic("https://customer.example.com/")))
		f, err := verification.NewPostHookFlow(conf, time.Hour, "csrf", r, nil, &original)
		require.NoError(t, err)
		assert.Equal(t, "https://customer.example.com/", f.GetCourierBaseURL())
	})

	t.Run("preserves http scheme through the post-hook chain", func(t *testing.T) {
		// Regression test for x.WithBaseURL no longer coercing scheme to
		// https. The Ory CLI / NextJS proxy case carries http://localhost.
		original := registration.Flow{RequestURL: "http://foo.com/bar"}
		r := (&http.Request{URL: urlx.ParseOrPanic("http://foo/bar"), Host: "foo"}).
			WithContext(x.WithBaseURL(context.Background(), urlx.ParseOrPanic("http://localhost:4000")))
		f, err := verification.NewPostHookFlow(conf, time.Hour, "csrf", r, nil, &original)
		require.NoError(t, err)
		assert.Equal(t, "http://localhost:4000", f.GetCourierBaseURL())
	})

	t.Run("empty when triggering request has no base URL", func(t *testing.T) {
		original := registration.Flow{RequestURL: "http://foo.com/bar"}
		r := &http.Request{URL: urlx.ParseOrPanic("http://foo/bar"), Host: "foo"}
		f, err := verification.NewPostHookFlow(conf, time.Hour, "csrf", r, nil, &original)
		require.NoError(t, err)
		assert.Equal(t, "", f.GetCourierBaseURL())
	})
}

// TestCourierBaseURLStoredInInternalContext locks down the storage
// location: the captured base URL must land in InternalContext under
// flow.InternalContextKeyCourierBaseURL, not in a dedicated column. If a
// future refactor moves it elsewhere, this fails — separately from the
// getter contract.
func TestCourierBaseURLStoredInInternalContext(t *testing.T) {
	conf := pkg.NewConfigurationWithDefaults(t)

	r := (&http.Request{URL: urlx.ParseOrPanic("http://foo/bar"), Host: "foo"}).
		WithContext(x.WithBaseURL(context.Background(), urlx.ParseOrPanic("https://customer.example.com/")))
	f, err := verification.NewFlow(conf, time.Hour, "csrf", r, nil, flow.TypeBrowser)
	require.NoError(t, err)

	got := gjson.GetBytes(f.InternalContext, flow.InternalContextKeyCourierBaseURL).String()
	assert.Equal(t, "https://customer.example.com/", got)
}

func TestFromOldFlow(t *testing.T) {
	ctx := context.Background()
	conf := pkg.NewConfigurationWithDefaults(t)
	r := http.Request{URL: &url.URL{Path: "/", RawQuery: "return_to=" + urlx.AppendPaths(conf.SelfPublicURL(ctx), "/self-service/login/browser").String()}, Host: "ory.sh"}
	for _, ft := range []flow.Type{
		flow.TypeAPI,
		flow.TypeBrowser,
	} {
		t.Run(fmt.Sprintf("case=original flow is %s", ft), func(t *testing.T) {
			f, err := verification.NewFlow(conf, 0, "csrf", &r, nil, ft)
			require.NoError(t, err)
			nf, err := verification.FromOldFlow(conf, time.Duration(time.Hour), f.CSRFToken, &r, nil, f)
			require.NoError(t, err)
			require.Equal(t, flow.TypeBrowser, nf.Type)
		})
	}
}

func TestContinueURL(t *testing.T) {
	const globalReturnTo = "https://ory.sh/global-return-to"
	const localReturnTo = "https://ory.sh/local-return-to"
	const flowReturnTo = "https://ory.sh/flow-return-to"

	for _, tc := range []struct {
		desc       string
		prep       func(conf *config.Config)
		requestURL string
		expect     string
	}{
		{
			desc: "return_to has precedence over global return to",
			prep: func(conf *config.Config) {
				conf.MustSet(context.Background(), config.ViperKeyURLsAllowedReturnToDomains, []string{localReturnTo})
			},
			requestURL: fmt.Sprintf("http://kratos:4433/verification?return_to=%s", localReturnTo),
			expect:     localReturnTo,
		},
		{
			desc:       "with return_to not allowed",
			requestURL: fmt.Sprintf("http://kratos:4433/verification?return_to=%s", localReturnTo),
			expect:     globalReturnTo,
		},
		{
			desc:       "with invalid request url",
			requestURL: string([]byte{0x7f}), // 0x7f is an ASCII control char, and fails URL validation
			expect:     globalReturnTo,
		},
		{
			desc: "flow return to has precedence over global return to",
			prep: func(conf *config.Config) {
				conf.MustSet(context.Background(), config.ViperKeySelfServiceVerificationBrowserDefaultReturnTo, flowReturnTo)
			},
			requestURL: "http://kratos:4433/verification",
			expect:     flowReturnTo,
		},
		{
			desc: "return_to has precedence over flow return to",
			prep: func(conf *config.Config) {
				conf.MustSet(context.Background(), config.ViperKeySelfServiceVerificationBrowserDefaultReturnTo, flowReturnTo)
				conf.MustSet(context.Background(), config.ViperKeyURLsAllowedReturnToDomains, []string{localReturnTo})
			},
			requestURL: fmt.Sprintf("http://kratos:4433/verification?return_to=%s", localReturnTo),
			expect:     localReturnTo,
		},
	} {
		t.Run(fmt.Sprintf("case=%s", tc.desc), func(t *testing.T) {
			conf := pkg.NewConfigurationWithDefaults(t)
			conf.MustSet(context.Background(), config.ViperKeySelfServiceBrowserDefaultReturnTo, globalReturnTo)
			if tc.prep != nil {
				tc.prep(conf)
			}
			flow := verification.Flow{
				RequestURL: tc.requestURL,
			}

			url := flow.ContinueURL(context.Background(), conf)
			require.NotNil(t, url)

			require.Equal(t, tc.expect, url.String())
		})
	}
}
