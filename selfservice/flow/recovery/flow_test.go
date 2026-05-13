// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package recovery_test

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/ory/kratos/pkg"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/recovery"
	"github.com/ory/kratos/selfservice/strategy/code"
	"github.com/ory/kratos/selfservice/strategy/link"
	"github.com/ory/kratos/x"
	"github.com/ory/x/jsonx"
	"github.com/ory/x/urlx"
)

func TestFlow(t *testing.T) {
	ctx := context.Background()
	conf := pkg.NewConfigurationWithDefaults(t)

	must := func(r *recovery.Flow, err error) *recovery.Flow {
		require.NoError(t, err)
		return r
	}

	u := &http.Request{URL: urlx.ParseOrPanic("http://foo/bar/baz"), Host: "foo"}
	for k, tc := range []struct {
		r         *recovery.Flow
		expectErr bool
	}{
		{r: must(recovery.NewFlow(conf, time.Hour, "", u, nil, flow.TypeBrowser))},
		{r: must(recovery.NewFlow(conf, -time.Hour, "", u, nil, flow.TypeBrowser)), expectErr: true},
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

	assert.EqualValues(t, flow.StateChooseMethod,
		must(recovery.NewFlow(conf, time.Hour, "", u, nil, flow.TypeBrowser)).State)

	t.Run("type=return_to", func(t *testing.T) {
		_, err := recovery.NewFlow(conf, 0, "csrf", &http.Request{URL: &url.URL{Path: "/", RawQuery: "return_to=https://not-allowed/foobar"}, Host: "ory.sh"}, nil, flow.TypeBrowser)
		require.Error(t, err)

		_, err = recovery.NewFlow(conf, 0, "csrf", &http.Request{URL: &url.URL{Path: "/", RawQuery: "return_to=" + urlx.AppendPaths(conf.SelfPublicURL(ctx), "/self-service/login/browser").String()}, Host: "ory.sh"}, nil, flow.TypeBrowser)
		require.NoError(t, err)
	})
}

func TestGetType(t *testing.T) {
	for _, ft := range []flow.Type{
		flow.TypeAPI,
		flow.TypeBrowser,
	} {
		t.Run(fmt.Sprintf("case=%s", ft), func(t *testing.T) {
			r := &recovery.Flow{Type: ft}
			assert.Equal(t, ft, r.GetType())
		})
	}
}

func TestGetRequestURL(t *testing.T) {
	expectedURL := "http://foo/bar/baz"
	f := &recovery.Flow{RequestURL: expectedURL}
	assert.Equal(t, expectedURL, f.GetRequestURL())
}

func TestNewFlow_capturesCourierBaseURL(t *testing.T) {
	conf := pkg.NewConfigurationWithDefaults(t)

	t.Run("nothing in context", func(t *testing.T) {
		r := &http.Request{URL: urlx.ParseOrPanic("http://foo/bar"), Host: "foo"}
		f, err := recovery.NewFlow(conf, time.Hour, "", r, nil, flow.TypeBrowser)
		require.NoError(t, err)
		assert.Equal(t, "", f.GetCourierBaseURL())
	})

	t.Run("BaseURL in request context is copied onto flow", func(t *testing.T) {
		base := urlx.ParseOrPanic("https://customer.example.com/")
		r := (&http.Request{URL: urlx.ParseOrPanic("http://foo/bar"), Host: "foo"}).
			WithContext(x.WithBaseURL(context.Background(), base))
		f, err := recovery.NewFlow(conf, time.Hour, "", r, nil, flow.TypeBrowser)
		require.NoError(t, err)
		assert.Equal(t, base.String(), f.GetCourierBaseURL())
	})

	t.Run("http scheme is preserved (not coerced to https)", func(t *testing.T) {
		base := urlx.ParseOrPanic("http://localhost:4000")
		r := (&http.Request{URL: urlx.ParseOrPanic("http://foo/bar"), Host: "foo"}).
			WithContext(x.WithBaseURL(context.Background(), base))
		f, err := recovery.NewFlow(conf, time.Hour, "", r, nil, flow.TypeBrowser)
		require.NoError(t, err)
		assert.Equal(t, "http://localhost:4000", f.GetCourierBaseURL())
	})
}

func TestFlowEncodeJSON(t *testing.T) {
	assert.EqualValues(t, "", gjson.Get(jsonx.TestMarshalJSONString(t, &recovery.Flow{RequestURL: "https://foo.bar?foo=bar"}), "return_to").String())
	assert.EqualValues(t, "/bar", gjson.Get(jsonx.TestMarshalJSONString(t, &recovery.Flow{RequestURL: "https://foo.bar?return_to=/bar"}), "return_to").String())
	assert.EqualValues(t, "/bar", gjson.Get(jsonx.TestMarshalJSONString(t, recovery.Flow{RequestURL: "https://foo.bar?return_to=/bar"}), "return_to").String())
}

func TestFromOldFlow(t *testing.T) {
	ctx := context.Background()
	conf, reg := pkg.NewVeryFastRegistryWithoutDB(t)
	r := http.Request{URL: &url.URL{Path: "/", RawQuery: "return_to=" + urlx.AppendPaths(conf.SelfPublicURL(ctx), "/self-service/login/browser").String()}, Host: "ory.sh"}
	t.Run("strategy=code", func(t *testing.T) {
		for _, ft := range []flow.Type{
			flow.TypeAPI,
			flow.TypeBrowser,
		} {
			t.Run(fmt.Sprintf("case=original flow is %s", ft), func(t *testing.T) {
				f, err := recovery.NewFlow(conf, 0, "csrf", &r, recovery.Strategies{code.NewStrategy(reg)}, ft)
				require.NoError(t, err)
				nF, err := recovery.FromOldFlow(conf, time.Duration(time.Hour), f.CSRFToken, &r, nil, *f)
				require.NoError(t, err)
				require.Equal(t, ft, nF.Type)
			})
		}
	})

	t.Run("strategy=link", func(t *testing.T) {
		for _, ft := range []flow.Type{
			flow.TypeAPI,
			flow.TypeBrowser,
		} {
			t.Run(fmt.Sprintf("case=original flow is %s", ft), func(t *testing.T) {
				f, err := recovery.NewFlow(conf, 0, "csrf", &r, recovery.Strategies{link.NewStrategy(reg)}, ft)
				require.NoError(t, err)
				nF, err := recovery.FromOldFlow(conf, time.Duration(time.Hour), f.CSRFToken, &r, nil, *f)
				require.NoError(t, err)
				require.Equal(t, flow.TypeBrowser, nF.Type)
			})
		}
	})
}

func TestFromOldFlow_capturesCourierBaseURL(t *testing.T) {
	conf := pkg.NewConfigurationWithDefaults(t)

	t.Run("uses the new request's context, not the old flow's value", func(t *testing.T) {
		// Old flow was created at customer.example.com.
		oldR := (&http.Request{URL: urlx.ParseOrPanic("http://foo/bar"), Host: "foo"}).
			WithContext(x.WithBaseURL(context.Background(), urlx.ParseOrPanic("https://customer.example.com/")))
		oldF, err := recovery.NewFlow(conf, time.Hour, "csrf", oldR, nil, flow.TypeBrowser)
		require.NoError(t, err)
		require.Equal(t, "https://customer.example.com/", oldF.GetCourierBaseURL())

		// New request comes through a different proxy / CNAME — we should
		// reflect that, not silently inherit the old flow's captured URL.
		newR := (&http.Request{URL: urlx.ParseOrPanic("http://foo/bar"), Host: "foo"}).
			WithContext(x.WithBaseURL(context.Background(), urlx.ParseOrPanic("http://localhost:4000")))
		newF, err := recovery.FromOldFlow(conf, time.Hour, oldF.CSRFToken, newR, nil, *oldF)
		require.NoError(t, err)
		assert.Equal(t, "http://localhost:4000", newF.GetCourierBaseURL())
	})

	t.Run("empty when neither old nor new have a base URL", func(t *testing.T) {
		r := &http.Request{URL: urlx.ParseOrPanic("http://foo/bar"), Host: "foo"}
		oldF, err := recovery.NewFlow(conf, time.Hour, "csrf", r, nil, flow.TypeBrowser)
		require.NoError(t, err)
		newF, err := recovery.FromOldFlow(conf, time.Hour, oldF.CSRFToken, r, nil, *oldF)
		require.NoError(t, err)
		assert.Equal(t, "", newF.GetCourierBaseURL())
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
	f, err := recovery.NewFlow(conf, time.Hour, "csrf", r, nil, flow.TypeBrowser)
	require.NoError(t, err)

	got := gjson.GetBytes(f.InternalContext, flow.InternalContextKeyCourierBaseURL).String()
	assert.Equal(t, "https://customer.example.com/", got)
}

func TestFlowDontOverrideReturnTo(t *testing.T) {
	f := &recovery.Flow{ReturnTo: "/foo"}
	f.SetReturnTo()
	assert.Equal(t, "/foo", f.ReturnTo)

	f = &recovery.Flow{RequestURL: "https://foo.bar?return_to=/bar"}
	f.SetReturnTo()
	assert.Equal(t, "/bar", f.ReturnTo)
}
