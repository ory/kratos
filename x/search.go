package x

import (
	"github.com/ory/herodot"
	"net/http"

	"github.com/pkg/errors"

	"github.com/ory/kratos/identity"
)

func ParseSearch(r *http.Request) (match string, kind identity.CredentialsType, err error) {
	match = r.URL.Query().Get("match")
	var matchType interface{} = r.URL.Query().Get("type")
	kind, ok := matchType.(identity.CredentialsType)
	if !ok || match == "" {
		return "", "", errors.WithStack(herodot.ErrInternalServerError.WithReasonf("You must set the credential type and match"))
	}
	return
}
