// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package x

import (
	"github.com/ory/x/httprouterx"
)

type PublicHandler interface {
	RegisterPublicRoutes(public *httprouterx.RouterPublic)
}

type AdminHandler interface {
	RegisterAdminRoutes(admin *httprouterx.RouterAdmin)
}

type Handler interface {
	PublicHandler
	AdminHandler
}
