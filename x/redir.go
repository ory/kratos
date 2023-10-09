// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package x

import (
	"net/http"
	"path"
	"strings"

	"github.com/julienschmidt/httprouter"

	"github.com/ory/kratos/driver/config"
)

func RedirectToAdminRoute(reg config.Provider) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		admin := reg.Config().SelfAdminURL(r.Context())

		dest := *r.URL
		dest.Host = admin.Host
		dest.Scheme = admin.Scheme
		dest.Path = strings.TrimPrefix(dest.Path, AdminPrefix)
		dest.Path = path.Join(admin.Path, AdminPrefix, dest.Path)

		http.Redirect(w, r, dest.String(), http.StatusTemporaryRedirect)
	}
}

func RedirectToPublicRoute(reg config.Provider) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		public := reg.Config().SelfPublicURL(r.Context())

		dest := *r.URL
		dest.Host = public.Host
		dest.Scheme = public.Scheme
		dest.Path = strings.TrimPrefix(dest.Path, AdminPrefix)
		dest.Path = path.Join(public.Path, dest.Path)

		http.Redirect(w, r, dest.String(), http.StatusTemporaryRedirect)
	}
}
