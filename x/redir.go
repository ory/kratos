// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package x

import (
	"context"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/julienschmidt/httprouter"
)

func RedirectToAdminRoute(c interface {
	SelfAdminURL(ctx context.Context) *url.URL
}) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		admin := c.SelfAdminURL(r.Context())

		dest := *r.URL
		dest.Host = admin.Host
		dest.Scheme = admin.Scheme
		dest.Path = strings.TrimPrefix(dest.Path, AdminPrefix)
		dest.Path = path.Join(admin.Path, AdminPrefix, dest.Path)

		http.Redirect(w, r, dest.String(), http.StatusTemporaryRedirect)
	}
}

func RedirectToPublicRoute(c interface {
	SelfPublicURL(ctx context.Context) *url.URL
}) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		public := c.SelfPublicURL(r.Context())

		dest := *r.URL
		dest.Host = public.Host
		dest.Scheme = public.Scheme
		dest.Path = strings.TrimPrefix(dest.Path, AdminPrefix)
		dest.Path = path.Join(public.Path, dest.Path)

		http.Redirect(w, r, dest.String(), http.StatusTemporaryRedirect)
	}
}
