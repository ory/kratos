// Copyright © 2026 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package email_test

import (
	"strings"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/courier/template/email"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/pkg"
	"github.com/ory/kratos/x"
)

func TestAuthenticatorKeyAdded(t *testing.T) {
	ctx := t.Context()
	_, reg := pkg.NewFastRegistryWithMocks(t)

	// Build the Identity map exactly as the courier does in production
	// (identity.Manager.sendIdentityNotifications uses x.StructToMap), so the
	// template's key lookup is tested against the real key casing instead of a
	// hand-crafted map. StructToMap marshals via JSON tags, so the ID lands under
	// the lowercase "id" key.
	id := &identity.Identity{ID: uuid.Must(uuid.FromString("00000000-0000-0000-0000-000000000001"))}
	idMap, err := x.StructToMap(id)
	require.NoError(t, err)

	tpl := email.NewAuthenticatorKeyAdded(reg, &email.AuthenticatorKeyAddedModel{
		To:       "owner@example.com",
		AddedAt:  "2026-04-21T12:00:00Z",
		Identity: idMap,
	})

	recipient, err := tpl.EmailRecipient()
	require.NoError(t, err)
	assert.Equal(t, "owner@example.com", recipient)

	subject, err := tpl.EmailSubject(ctx)
	require.NoError(t, err)
	assert.Contains(t, strings.ToLower(subject), "sign-in method")

	body, err := tpl.EmailBody(ctx)
	require.NoError(t, err)
	assert.Contains(t, body, "00000000-0000-0000-0000-000000000001", "the account ID must render in the email body")

	plain, err := tpl.EmailBodyPlaintext(ctx)
	require.NoError(t, err)
	assert.Contains(t, plain, "00000000-0000-0000-0000-000000000001", "the account ID must render in the plaintext body")
	assert.Contains(t, plain, "2026-04-21T12:00:00Z")
}
