// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

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

func TestExpandables_ToEager_skips_Identity(t *testing.T) {
	e := ExpandEverything

	res := e.ToEager()
	assert.Equal(t, 2, len(res))
	assert.Equal(t, []string{string(ExpandSessionDevices), string(ExpandSessionIdentity)}, res)
}

func TestExpandables_ExpandNothing_IsEmpty(t *testing.T) {
	e := ExpandNothing

	assert.True(t, len(e) == 0)
}
