package identity

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/x"
)

func TestNewRecoveryEmailAddress(t *testing.T) {
	iid := x.NewUUID()
	a, err := NewRecoveryEmailAddress("foo@ory.sh", iid, time.Minute)
	require.NoError(t, err)

	assert.Len(t, a.Code, 32)
	assert.Equal(t, a.Value, "foo@ory.sh")
	assert.Equal(t, a.Via, RecoveryAddressTypeEmail)
	assert.NotEmpty(t, a.ID)

	out, err := json.Marshal(a)
	require.NoError(t, err)
	assert.NotContains(t, out, a.Code)
}
