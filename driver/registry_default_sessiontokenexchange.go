// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package driver

import "github.com/ory/kratos/selfservice/sessiontokenexchange"

func (m *RegistryDefault) SessionTokenExchangePersister() sessiontokenexchange.Persister {
	return m.Persister()
}
