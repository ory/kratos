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
