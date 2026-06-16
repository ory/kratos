// Copyright © 2026 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package password_test

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/pkg"
	"github.com/ory/kratos/pkg/testhelpers"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/ui/node"
	"github.com/ory/x/configx"
)

func TestLoginIdentifierAutocomplete(t *testing.T) {
	_, reg := pkg.NewFastRegistryWithMocks(t,
		configx.WithValue(config.ViperKeySelfServiceStrategyConfig+"."+string(identity.CredentialsTypePassword)+".enabled", true),
		configx.WithValues(testhelpers.DefaultIdentitySchemaConfig("file://stub/login.schema.json")),
	)

	s, err := reg.AllLoginStrategies().Strategy(identity.CredentialsTypePassword)
	require.NoError(t, err)
	fh, ok := s.(login.AAL1FormHydrator)
	require.True(t, ok)

	r := httptest.NewRequest("GET", "/self-service/login/browser", nil).WithContext(t.Context())
	f, err := login.NewFlow(reg, r, flow.TypeBrowser)
	require.NoError(t, err)

	require.NoError(t, fh.PopulateLoginMethodFirstFactor(r, f))

	identifier := f.UI.Nodes.Find("identifier")
	require.NotNil(t, identifier, "identifier node must exist")
	attr, ok := identifier.Attributes.(*node.InputAttributes)
	require.True(t, ok)
	require.Equal(t, node.InputAttributeAutocompleteUsername, attr.Autocomplete)
}
