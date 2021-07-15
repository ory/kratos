package x

import (
	"net/http"
	"path/filepath"

	"github.com/julienschmidt/httprouter"

	"github.com/ory/kratos/driver/config"
)

func RedirectToAdminRoute(reg config.Provider) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		admin := reg.Config(r.Context()).SelfAdminURL()

		dest := *r.URL
		dest.Host = admin.Host
		dest.Scheme = admin.Scheme
		dest.Path = filepath.Join(admin.Path + dest.Path)

		http.Redirect(w, r, dest.String(), http.StatusTemporaryRedirect)
	}
}

func RedirectToPublicRoute(reg config.Provider) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		public := reg.Config(r.Context()).SelfPublicURL(r)

		dest := *r.URL
		dest.Host = public.Host
		dest.Scheme = public.Scheme
		dest.Path = filepath.Join(public.Path + dest.Path)

		http.Redirect(w, r, dest.String(), http.StatusTemporaryRedirect)
	}
}
