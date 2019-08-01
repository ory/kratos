package x

import (
	"net/url"
	"strings"

	"github.com/pkg/errors"

	"github.com/ory/herodot"
)

func DetermineReturnToURL(request *url.URL, defaultReturnTo *url.URL, whitelistedDomains []url.URL) (string, error) {
	u, err := url.ParseRequestURI(request.Query().Get("return_to"))
	if len(request.Query().Get("return_to")) == 0 || err != nil {
		return defaultReturnTo.String(), nil
	}

	var found bool
	for _, wd := range whitelistedDomains {
		if strings.EqualFold(wd.Scheme, u.Scheme) && strings.EqualFold(wd.Host, u.Host) {
			found = true
		}
	}

	if !found && len(u.Host) > 0 {
		return "", errors.WithStack(herodot.ErrBadRequest.WithReasonf("Requested return_to domain \"%s\" is not a whitelisted return domain", u.Host))
	}

	if len(u.Host) == 0 {
		u.Host = defaultReturnTo.Host
	}

	if len(u.Scheme) == 0 {
		u.Scheme = defaultReturnTo.Scheme
	}

	return u.String(), nil
}
