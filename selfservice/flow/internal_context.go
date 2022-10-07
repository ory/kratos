// Copyright © 2022 Ory Corp

package flow

import "github.com/ory/kratos/identity"

func PrefixInternalContextKey(t identity.CredentialsType, suffix string) string {
	return string(t) + "_" + suffix
}
