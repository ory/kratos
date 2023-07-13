// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package hook_test

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/internal/testhelpers"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/registration"
	"github.com/ory/kratos/selfservice/hook"
	"github.com/ory/kratos/selfservice/strategy/code"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/x"
	"github.com/ory/x/randx"
)

func TestCodeAddressVerifier(t *testing.T) {
	ctx := context.Background()
	conf, reg := internal.NewFastRegistryWithMocks(t)
	testhelpers.SetDefaultIdentitySchema(conf, "file://./stub/code.schema.json")
	verifier := hook.NewCodeAddressVerifier(reg)

	setup := func(t *testing.T) (address string, rf *registration.Flow) {
		t.Helper()
		address = testhelpers.RandomEmail()
		rawCode := strings.ToLower(randx.MustString(16, randx.Alpha))

		rf = &registration.Flow{Active: identity.CredentialsTypeCodeAuth, Type: "browser", State: flow.StatePassedChallenge}
		require.NoError(t, reg.RegistrationFlowPersister().CreateRegistrationFlow(ctx, rf))

		_, err := reg.RegistrationCodePersister().CreateRegistrationCode(ctx, &code.CreateRegistrationCodeParams{
			Address:     address,
			AddressType: identity.CodeAddressTypeEmail,
			RawCode:     rawCode,
			ExpiresIn:   time.Hour,
			FlowID:      rf.ID,
		})
		require.NoError(t, err)

		_, err = reg.RegistrationCodePersister().UseRegistrationCode(ctx, rf.ID, rawCode, address)
		require.NoError(t, err)

		return
	}

	setupIdentity := func(t *testing.T, address string) *identity.Identity {
		t.Helper()
		verifiableAddress := []identity.VerifiableAddress{{ID: uuid.UUID{}, Verified: false, Value: address, Via: identity.VerifiableAddressTypeEmail}}
		id := &identity.Identity{ID: x.NewUUID(), VerifiableAddresses: verifiableAddress, Credentials: map[identity.CredentialsType]identity.Credentials{
			identity.CredentialsTypeCodeAuth: {Type: identity.CredentialsTypeCodeAuth, Identifiers: []string{address}},
		}}

		require.NoError(t, reg.PrivilegedIdentityPool().CreateIdentity(ctx, id))
		return id
	}

	runHook := func(t *testing.T, id *identity.Identity, flow *registration.Flow) {
		t.Helper()

		sessions := &session.Session{
			ID:       x.NewUUID(),
			Identity: id,
		}

		r := &http.Request{}
		require.NoError(t, verifier.ExecutePostRegistrationPostPersistHook(nil, r, flow, sessions))
	}

	t.Run("case=should set the verifiable email address to verified", func(t *testing.T) {
		address, flow := setup(t)
		id := setupIdentity(t, address)

		runHook(t, id, flow)
		va, err := reg.IdentityPool().FindVerifiableAddressByValue(ctx, identity.VerifiableAddressTypeEmail, address)
		require.NoError(t, err)
		require.True(t, va.Verified)
	})

	t.Run("case=should ignore verifiable email address that does not match the code", func(t *testing.T) {
		_, flow := setup(t)
		newEmail := testhelpers.RandomEmail()
		id := setupIdentity(t, newEmail)

		runHook(t, id, flow)
		va, err := reg.IdentityPool().FindVerifiableAddressByValue(ctx, identity.VerifiableAddressTypeEmail, newEmail)
		require.NoError(t, err)
		require.False(t, va.Verified)
	})
}
