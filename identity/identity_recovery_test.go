package identity

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ory/kratos/x"
)

func TestNewRecoveryEmailAddress(t *testing.T) {
	iid := x.NewUUID()
	a := NewRecoveryEmailAddress("foo@ory.sh", iid)

	assert.Equal(t, a.Value, "foo@ory.sh")
	assert.Equal(t, a.Via, RecoveryAddressTypeEmail)
	assert.NotEmpty(t, a.ID)
}
