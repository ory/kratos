// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

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
