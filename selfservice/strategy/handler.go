// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package strategy

import (
	"net/http"

	"github.com/julienschmidt/httprouter"

	"github.com/ory/herodot"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/x"
)

const EndpointDisabledMessage = "This endpoint was disabled by system administrator. Please check your url or contact the system administrator to enable it."

type disabledChecker interface {
	config.Provider
	x.WriterProvider
}

func disabledWriter(c disabledChecker, enabled bool, wrap httprouter.Handle, w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	if !enabled {
		c.Writer().WriteError(w, r, herodot.ErrNotFound.WithReason(EndpointDisabledMessage))
		return
	}
	wrap(w, r, ps)
}

func IsDisabled(c disabledChecker, strategy string, wrap httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		disabledWriter(c, c.Config().SelfServiceStrategy(r.Context(), strategy).Enabled, wrap, w, r, ps)
	}
}

func IsRecoveryDisabled(c disabledChecker, strategy string, wrap httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		disabledWriter(c,
			c.Config().SelfServiceStrategy(r.Context(), strategy).Enabled && c.Config().SelfServiceFlowRecoveryEnabled(r.Context()),
			wrap, w, r, ps)
	}
}

func IsVerificationDisabled(c disabledChecker, strategy string, wrap httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		disabledWriter(c,
			c.Config().SelfServiceStrategy(r.Context(), strategy).Enabled && c.Config().SelfServiceFlowVerificationEnabled(r.Context()),
			wrap, w, r, ps)
	}
}
