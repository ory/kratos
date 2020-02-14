package identity

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/x"
)

func TestNewVerifiableEmailAddress(t *testing.T) {
	iid := x.NewUUID()
	a, err := NewVerifiableEmailAddress("foo@ory.sh", iid, time.Minute)
	require.NoError(t, err)

	assert.Len(t, a.Code, 32)
	assert.Equal(t, a.Value, "foo@ory.sh")
	assert.Equal(t, a.Via, VerifiableAddressTypeEmail)
	assert.Equal(t, a.Status, VerifiableAddressStatusPending)
	assert.Equal(t, a.Verified, false)
	assert.Nil(t, a.VerifiedAt)
	assert.NotEmpty(t, a.ID)

	out, err := json.Marshal(a)
	require.NoError(t, err)
	assert.NotContains(t, out, a.Code)
}
