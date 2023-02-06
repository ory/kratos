// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package flow

import "github.com/ory/kratos/identity"

func PrefixInternalContextKey(t identity.CredentialsType, suffix string) string {
	return string(t) + "_" + suffix
}
