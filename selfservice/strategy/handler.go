package strategy

import (
	"github.com/julienschmidt/httprouter"
	"github.com/ory/herodot"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/x"
	"net/http"
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
		disabledWriter(c, c.Config(r.Context()).SelfServiceStrategy(strategy).Enabled, wrap, w, r, ps)
	}
}

func IsRecoveryDisabled(c disabledChecker, strategy string, wrap httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		disabledWriter(c,
			c.Config(r.Context()).SelfServiceStrategy(strategy).Enabled && c.Config(r.Context()).SelfServiceFlowRecoveryEnabled(),
			wrap, w, r, ps)
	}
}

func IsVerificationDisabled(c disabledChecker, strategy string, wrap httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		disabledWriter(c,
			c.Config(r.Context()).SelfServiceStrategy(strategy).Enabled && c.Config(r.Context()).SelfServiceFlowVerificationEnabled(),
			wrap, w, r, ps)
	}
}
