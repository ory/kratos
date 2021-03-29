package flow

import (
	"net/http"
	"net/url"

	"github.com/pkg/errors"

	"github.com/ory/herodot"
	"github.com/ory/kratos/x"

	"github.com/gofrs/uuid"

	"github.com/ory/x/urlx"
)

func AppendFlowTo(src *url.URL, id uuid.UUID) *url.URL {
	return urlx.CopyWithQuery(src, url.Values{"flow": {id.String()}})
}

func GetFlowID(r *http.Request) (uuid.UUID, error) {
	rid := x.ParseUUID(r.URL.Query().Get("flow"))
	if rid == uuid.Nil {
		return rid, errors.WithStack(herodot.ErrBadRequest.WithReasonf("The flow query parameter is missing or malformed."))
	}
	return rid, nil
}

type Flow interface {
	GetID() uuid.UUID
}
