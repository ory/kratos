// Copyright © 2022 Ory Corp

package identity

import (
	"net/url"
)

type Registry interface {
	IdentityPool() Pool
}

type Configuration interface {
	SelfAdminURL() *url.URL
	DefaultIdentityTraitsSchemaURL() *url.URL
}
