package identity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewIdentity(t *testing.T) {
	i := NewIdentity("")
	assert.NotEmpty(t, i.ID)
	// assert.NotEmpty(t, i.Metadata)
	assert.NotEmpty(t, i.Traits)
	assert.NotNil(t, i.Credentials)
}
