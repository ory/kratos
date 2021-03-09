package flow

import (
	"github.com/gofrs/uuid"
	"github.com/ory/x/urlx"
	"net/url"
)

func AppendFlowTo(src *url.URL, id uuid.UUID) *url.URL {
	return urlx.CopyWithQuery(src, url.Values{"flow": {id.String()}})
}
