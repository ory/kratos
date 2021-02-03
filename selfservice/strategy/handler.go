package strategy

import (
	"github.com/julienschmidt/httprouter"
	"github.com/ory/herodot"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/x"
	"net/http"
)

const EndpointDisabledMessage = "This endpoint was disabled by system administrator. Please check your url or contact the system administrator to enable it."

func IsDisabled(c interface {
	config.Provider
	x.WriterProvider
}, strategy string, wrap httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		enabled := c.Config(r.Context()).SelfServiceStrategy(strategy).Enabled
		if !enabled {
			c.Writer().WriteError(w, r, herodot.ErrNotFound.WithReason(EndpointDisabledMessage))
			return
		}
		wrap(w, r, ps)
	}
}
