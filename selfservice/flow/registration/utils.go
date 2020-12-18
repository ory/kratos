package registration

import (
	"net/http"

	"github.com/pkg/errors"

	"github.com/ory/herodot"
	"github.com/ory/kratos/x"
)

func GetFlow(r *http.Request, reg interface {
	FlowPersistenceProvider
}) (*Flow, error) {

	rid := x.ParseUUID(r.URL.Query().Get("flow"))
	if x.IsZeroUUID(rid) {
		return nil, errors.WithStack(herodot.ErrBadRequest.WithReasonf("The flow query parameter is missing."))
	}

	ar, err := reg.RegistrationFlowPersister().GetRegistrationFlow(r.Context(), rid)
	if err != nil {
		return nil, err
	}

	return ar, ar.Valid()
}
