// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package x_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/x"
)

func TestRedirectToPublicAdminRoute(t *testing.T) {
	ctx := context.Background()
	conf, reg := internal.NewFastRegistryWithMocks(t)
	pub := x.NewRouterPublic()
	adm := x.NewRouterAdmin()
	adminTS := httptest.NewServer(adm)
	pubTS := httptest.NewServer(pub)
	t.Cleanup(pubTS.Close)
	t.Cleanup(adminTS.Close)

	conf.MustSet(ctx, config.ViperKeyAdminBaseURL, adminTS.URL)
	conf.MustSet(ctx, config.ViperKeyPublicBaseURL, pubTS.URL)

	pub.POST("/privileged", x.RedirectToAdminRoute(reg))
	pub.POST("/admin/privileged", x.RedirectToAdminRoute(reg))
	adm.POST("/privileged", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		body, _ := io.ReadAll(r.Body)
		_, _ = w.Write(body)
	})

	adm.POST("/read", x.RedirectToPublicRoute(reg))
	pub.POST("/read", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		body, _ := io.ReadAll(r.Body)
		_, _ = w.Write(body)
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
