package node

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIDs(t *testing.T) {
	assert.EqualValues(t, "foo", (&AnchorAttributes{Identifier: "foo"}).ID())
	assert.EqualValues(t, "foo", (&ImageAttributes{Identifier: "foo"}).ID())
	assert.EqualValues(t, "foo", (&TextAttributes{Identifier: "foo"}).ID())
	assert.EqualValues(t, "foo", (&InputAttributes{Name: "foo"}).ID())
}
