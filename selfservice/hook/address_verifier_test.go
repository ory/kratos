package hook

import (
	"testing"

	"github.com/gofrs/uuid"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/x"
	"github.com/stretchr/testify/assert"
)

func TestAddressVerifier(t *testing.T) {
	verifier := NewAddressVerifier()

	t.Run("Single Address Not Verified", func(t *testing.T) {
		s := &session.Session{ID: x.NewUUID(), Identity: &identity.Identity{
			ID: x.NewUUID(),
			VerifiableAddresses: []identity.VerifiableAddress{
				{
					ID:       uuid.UUID{},
					Verified: false,
				},
			},
		}}

		err := verifier.ExecuteLoginPostHook(nil, nil, nil, s)

		assert.ErrorIs(t, err, login.ErrAddressNotVerified)
	})

	t.Run("Single Address Verified", func(t *testing.T) {
		s := &session.Session{ID: x.NewUUID(), Identity: &identity.Identity{
			ID: x.NewUUID(),
			VerifiableAddresses: []identity.VerifiableAddress{
				{
					ID:       uuid.UUID{},
					Verified: true,
				},
			},
		}}

		err := verifier.ExecuteLoginPostHook(nil, nil, nil, s)

		assert.NoError(t, err)
	})

	t.Run("Multiple Addresses Verified", func(t *testing.T) {
		s := &session.Session{ID: x.NewUUID(), Identity: &identity.Identity{
			ID: x.NewUUID(),
			VerifiableAddresses: []identity.VerifiableAddress{
				{
					ID:       uuid.UUID{},
					Verified: true,
				},
				{
					ID:       uuid.UUID{},
					Verified: true,
				},
			},
		}}

		err := verifier.ExecuteLoginPostHook(nil, nil, nil, s)

		assert.NoError(t, err)
	})

	t.Run("Multiple Addresses Not Verified", func(t *testing.T) {
		s := &session.Session{ID: x.NewUUID(), Identity: &identity.Identity{
			ID: x.NewUUID(),
			VerifiableAddresses: []identity.VerifiableAddress{
				{
					ID:       uuid.UUID{},
					Verified: false,
				},
				{
					ID:       uuid.UUID{},
					Verified: false,
				},
			},
		}}

		err := verifier.ExecuteLoginPostHook(nil, nil, nil, s)

		assert.ErrorIs(t, err, login.ErrAddressNotVerified)
	})

	t.Run("One Address Verified And One Not", func(t *testing.T) {
		s := &session.Session{ID: x.NewUUID(), Identity: &identity.Identity{
			ID: x.NewUUID(),
			VerifiableAddresses: []identity.VerifiableAddress{
				{
					ID:       uuid.UUID{},
					Verified: true,
				},
				{
					ID:       uuid.UUID{},
					Verified: false,
				},
			},
		}}

		err := verifier.ExecuteLoginPostHook(nil, nil, nil, s)

		assert.ErrorIs(t, err, login.ErrAddressNotVerified)
	})

}
