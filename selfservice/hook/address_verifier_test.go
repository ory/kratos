package hook

import (
	"testing"

	"github.com/gofrs/uuid"
	"github.com/ory/kratos/ui/node"
	"github.com/stretchr/testify/assert"

	"github.com/ory/herodot"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/x"
)

func TestAddressVerifier(t *testing.T) {
	verifier := NewAddressVerifier()

	for _, tc := range []struct {
		flow       *login.Flow
		neverError bool
	}{
		{&login.Flow{Active: identity.CredentialsTypePassword}, false},
		{&login.Flow{Active: identity.CredentialsTypeOIDC}, true},
	} {
		t.Run(tc.flow.Active.String()+" flow", func(t *testing.T) {
			for _, uc := range []struct {
				name                string
				verifiableAddresses []identity.VerifiableAddress
				expectedError       error
			}{
				{
					name:                "No Verification Address",
					verifiableAddresses: []identity.VerifiableAddress{},
					expectedError:       herodot.ErrInternalServerError.WithReason("A misconfiguration prevents login. Expected to find a verification address but this identity does not have one assigned."),
				},
				{
					name: "Single Address Not Verified",
					verifiableAddresses: []identity.VerifiableAddress{
						{ID: uuid.UUID{}, Verified: false},
					},
					expectedError: login.ErrAddressNotVerified,
				},
				{
					name: "Single Address Verified",
					verifiableAddresses: []identity.VerifiableAddress{
						{ID: uuid.UUID{}, Verified: true},
					},
				},
				{
					name: "Multiple Addresses Verified",
					verifiableAddresses: []identity.VerifiableAddress{
						{ID: uuid.UUID{}, Verified: true},
						{ID: uuid.UUID{}, Verified: true},
					},
				},
				{
					name: "Multiple Addresses Not Verified",
					verifiableAddresses: []identity.VerifiableAddress{
						{ID: uuid.UUID{}, Verified: false},
						{ID: uuid.UUID{}, Verified: false},
					},
					expectedError: login.ErrAddressNotVerified,
				},
				{
					name: "One Address Verified And One Not",
					verifiableAddresses: []identity.VerifiableAddress{
						{ID: uuid.UUID{}, Verified: true},
						{ID: uuid.UUID{}, Verified: false},
					},
				},
			} {
				t.Run(uc.name, func(t *testing.T) {
					sessions := &session.Session{
						ID:       x.NewUUID(),
						Identity: &identity.Identity{ID: x.NewUUID(), VerifiableAddresses: uc.verifiableAddresses},
					}

					err := verifier.ExecuteLoginPostHook(nil, nil, node.DefaultGroup, tc.flow, sessions)

					if tc.neverError || uc.expectedError == nil {
						assert.NoError(t, err)
					} else {
						assert.ErrorIs(t, err, uc.expectedError)
					}
				})
			}
		})
	}
}
