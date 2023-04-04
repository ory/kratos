// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package x

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/urfave/negroni"
)

var CleanPath negroni.HandlerFunc = func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	r.URL.Path = httprouter.CleanPath(r.URL.Path)
	next(rw, r)
}
