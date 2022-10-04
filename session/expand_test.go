package session

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExpandableSearch_ExistingExpand(t *testing.T) {
	e := ExpandEverything

	assert.True(t, e.Has(ExpandSessionDevices))
}

func TestExpandableSearch_NonExistingExpandOption(t *testing.T) {
	e := ExpandEverything

	assert.False(t, e.Has("SomeExpand"))
}
