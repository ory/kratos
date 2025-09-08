// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package redir_test

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/x"
	"github.com/ory/kratos/x/redir"
	"github.com/ory/x/configx"
)

func TestRedirectToPublicAdminRoute(t *testing.T) {
	pub := x.NewTestRouterPublic(t)
	adm := x.NewTestRouterAdmin(t)
	adminTS := httptest.NewServer(adm)
	pubTS := httptest.NewServer(pub)
	t.Cleanup(pubTS.Close)
	t.Cleanup(adminTS.Close)
	_, reg := internal.NewFastRegistryWithMocks(t, configx.WithValues(map[string]any{
		config.ViperKeyAdminBaseURL:  adminTS.URL,
		config.ViperKeyPublicBaseURL: pubTS.URL,
	}))

	pub.POST("/privileged", redir.RedirectToAdminRoute(reg))
	pub.POST("/admin/privileged", redir.RedirectToAdminRoute(reg))
	adm.POST("/privileged", func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(w, r.Body)
	})

	adm.POST("/read", redir.RedirectToPublicRoute(reg))
	pub.POST("/read", func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(w, r.Body)
	})

	for k, tc := range []struct {
		source string
		dest   string
	}{
		{
			source: pubTS.URL + "/privileged?foo=bar",
			dest:   adminTS.URL + "/admin/privileged?foo=bar",
		},
		{
			source: pubTS.URL + "/admin/privileged?foo=bar",
			dest:   adminTS.URL + "/admin/privileged?foo=bar",
		},
		{
			source: adminTS.URL + "/admin/read?foo=bar",
			dest:   pubTS.URL + "/read?foo=bar",
		},
	} {
		t.Run(fmt.Sprintf("%d", k), func(t *testing.T) {
			id := x.NewUUID().String()
			res, err := adminTS.Client().Post(tc.source, "", strings.NewReader(id))
			require.NoError(t, err)
			assert.EqualValues(t, http.StatusOK, res.StatusCode)
			assert.Equal(t, tc.dest, res.Request.URL.String())
			body, err := io.ReadAll(res.Body)
			require.NoError(t, err)
			assert.Equal(t, id, string(body))
		})
	}
}
