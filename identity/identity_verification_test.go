package identity

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ory/x/sqlxx"

	"github.com/ory/kratos/x"
)

func TestNewVerifiableEmailAddress(t *testing.T) {
	iid := x.NewUUID()
	a := NewVerifiableEmailAddress("foo@ory.sh", iid)
	var nullTime *sqlxx.NullTime

	assert.Equal(t, a.Value, "foo@ory.sh")
	assert.Equal(t, a.Via, VerifiableAddressTypeEmail)
	assert.Equal(t, a.Status, VerifiableAddressStatusPending)
	assert.Equal(t, a.Verified, false)
	assert.EqualValues(t, nullTime, a.VerifiedAt)
	assert.NotEmpty(t, a.ID)
}
