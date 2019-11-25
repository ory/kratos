package x

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUUID(t *testing.T) {
	assert.True(t, IsZeroUUID(ParseUUID("foo")))
	assert.True(t, IsZeroUUID(ParseUUID("")))
	assert.True(t, IsZeroUUID(ParseUUID("asfdt4ifgdsl")))
	assert.False(t, IsZeroUUID(NewUUID()))
}
