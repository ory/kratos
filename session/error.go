// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package session

import (
	"net/http"

	"github.com/julienschmidt/httprouter"

	"github.com/ory/herodot"
)

func RespondWithJSONErrorOnAuthenticated(h herodot.Writer, err error) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		h.WriteError(w, r, err)
	}
}

var ErrNoSessionFound = herodot.ErrUnauthorized.WithReasonf("No valid session credentials found in the request.")
