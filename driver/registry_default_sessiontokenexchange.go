// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package driver

import "github.com/ory/kratos/selfservice/sessiontokenexchange"

func (m *RegistryDefault) SessionTokenExchangeHandler() *sessiontokenexchange.Handler {
	if m.sessionTokenExchangeHandler == nil {
		m.sessionTokenExchangeHandler = sessiontokenexchange.NewHandler(m)
	}

	return m.sessionTokenExchangeHandler
}

func (m *RegistryDefault) SessionTokenExchangePersister() sessiontokenexchange.Persister {
	return m.Persister()
}
