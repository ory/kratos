// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package redir

import (
	"net/http"
	"path"
	"strings"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/x"
)

func RedirectToAdminRoute(reg config.Provider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		admin := reg.Config().SelfAdminURL(r.Context())

		dest := *r.URL
		dest.Host = admin.Host
		dest.Scheme = admin.Scheme
		dest.Path = strings.TrimPrefix(dest.Path, x.AdminPrefix)
		dest.Path = path.Join(admin.Path, x.AdminPrefix, dest.Path)

		http.Redirect(w, r, dest.String(), http.StatusTemporaryRedirect)
	}
}

func RedirectToPublicRoute(reg config.Provider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		public := reg.Config().SelfPublicURL(r.Context())

		dest := *r.URL
		dest.Host = public.Host
		dest.Scheme = public.Scheme
		dest.Path = strings.TrimPrefix(dest.Path, x.AdminPrefix)
		dest.Path = path.Join(public.Path, dest.Path)

		http.Redirect(w, r, dest.String(), http.StatusTemporaryRedirect)
	}
}
