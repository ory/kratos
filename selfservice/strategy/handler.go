// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package strategy

import (
	"net/http"

	"github.com/ory/herodot"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/x"
)

const EndpointDisabledMessage = "This endpoint was disabled by system administrator. Please check your url or contact the system administrator to enable it."

type disabledChecker interface {
	config.Provider
	x.WriterProvider
}

func disabledWriter(c disabledChecker, enabled bool, wrap http.HandlerFunc, w http.ResponseWriter, r *http.Request) {
	if !enabled {
		c.Writer().WriteError(w, r, herodot.ErrNotFound.WithReason(EndpointDisabledMessage))
		return
	}
	wrap(w, r)
}

func IsDisabled(c disabledChecker, strategy string, wrap http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		disabledWriter(c, c.Config().SelfServiceStrategy(r.Context(), strategy).Enabled, wrap, w, r)
	}
}
