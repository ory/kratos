// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package flow

import "github.com/ory/kratos/identity"

// InternalContextKeyCourierBaseURL is the key under which recovery and
// verification flows stash the base URL captured at flow init. Email
// senders read it back at courier-send time to build the link domain. The
// key is top-level (not strategy-scoped) because the value is flow-wide.
const InternalContextKeyCourierBaseURL = "courier_base_url"

func PrefixInternalContextKey(t identity.CredentialsType, suffix string) string {
	return string(t) + "_" + suffix
}
