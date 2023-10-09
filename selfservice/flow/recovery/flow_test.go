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

	"github.com/tidwall/gjson"

	"github.com/ory/x/jsonx"

	"github.com/ory/kratos/internal"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/x/urlx"

	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/recovery"
)

func TestFlow(t *testing.T) {
	ctx := context.Background()
	conf, _ := internal.NewFastRegistryWithMocks(t)

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

func TestFlowEncodeJSON(t *testing.T) {
	assert.EqualValues(t, "", gjson.Get(jsonx.TestMarshalJSONString(t, &recovery.Flow{RequestURL: "https://foo.bar?foo=bar"}), "return_to").String())
	assert.EqualValues(t, "/bar", gjson.Get(jsonx.TestMarshalJSONString(t, &recovery.Flow{RequestURL: "https://foo.bar?return_to=/bar"}), "return_to").String())
	assert.EqualValues(t, "/bar", gjson.Get(jsonx.TestMarshalJSONString(t, recovery.Flow{RequestURL: "https://foo.bar?return_to=/bar"}), "return_to").String())
}

func TestFromOldFlow(t *testing.T) {
	ctx := context.Background()
	conf := internal.NewConfigurationWithDefaults(t)
	r := http.Request{URL: &url.URL{Path: "/", RawQuery: "return_to=" + urlx.AppendPaths(conf.SelfPublicURL(ctx), "/self-service/login/browser").String()}, Host: "ory.sh"}
	for _, ft := range []flow.Type{
		flow.TypeAPI,
		flow.TypeBrowser,
	} {
		t.Run(fmt.Sprintf("case=original flow is %s", ft), func(t *testing.T) {
			f, err := recovery.NewFlow(conf, 0, "csrf", &r, nil, ft)
			require.NoError(t, err)
			nF, err := recovery.FromOldFlow(conf, time.Duration(time.Hour), f.CSRFToken, &r, nil, *f)
			require.NoError(t, err)
			require.Equal(t, flow.TypeBrowser, nF.Type)
		})
	}
}

func TestFlowDontOverrideReturnTo(t *testing.T) {
	f := &recovery.Flow{ReturnTo: "/foo"}
	f.SetReturnTo()
	assert.Equal(t, "/foo", f.ReturnTo)

	f = &recovery.Flow{RequestURL: "https://foo.bar?return_to=/bar"}
	f.SetReturnTo()
	assert.Equal(t, "/bar", f.ReturnTo)
}
