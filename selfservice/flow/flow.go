package flow

import (
	"net/url"

	"github.com/gofrs/uuid"

	"github.com/ory/x/urlx"
)

func AppendFlowTo(src *url.URL, id uuid.UUID) *url.URL {
	return urlx.CopyWithQuery(src, url.Values{"flow": {id.String()}})
}
