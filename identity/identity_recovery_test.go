// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package identity

import (
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/ory/kratos/x"
)

func TestNewRecoveryEmailAddress(t *testing.T) {
	iid := x.NewUUID()
	a := NewRecoveryEmailAddress("foo@ory.sh", iid)

	assert.Equal(t, a.Value, "foo@ory.sh")
	assert.Equal(t, a.Via, RecoveryAddressTypeEmail)
	assert.Equal(t, iid, a.IdentityID)
	assert.Equal(t, uuid.Nil, a.ID)
}

// TestRecoveryAddress_Hash tests that the hash considers all fields that are
// written to the database (ignoring some well-known fields like the ID or
// timestamps).
func TestRecoveryAddress_Hash(t *testing.T) {
	cases := []struct {
		name string
		a    RecoveryAddress
	}{
		{
			name: "full fields",
			a: RecoveryAddress{
				ID:         x.NewUUID(),
				Value:      "foo@bar.me",
				Via:        AddressTypeEmail,
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
				IdentityID: x.NewUUID(),
				NID:        x.NewUUID(),
			},
		}, {
			name: "empty fields",
			a:    RecoveryAddress{},
		}, {
			name: "constructor",
			a:    *NewRecoveryEmailAddress("foo@ory.sh", x.NewUUID()),
		},
	}

	for _, tc := range cases {
		t.Run("case="+tc.name, func(t *testing.T) {
			assert.Equal(t,
				reflectiveHash(tc.a),
				tc.a.Hash(),
			)
		})
	}

}
