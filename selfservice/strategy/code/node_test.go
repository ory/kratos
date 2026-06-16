// Copyright © 2026 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package code

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/ui/node"
)

func TestNodeCodeInputFieldAutocomplete(t *testing.T) {
	n := nodeCodeInputField()

	attr, ok := n.Attributes.(*node.InputAttributes)
	require.True(t, ok, "expected input attributes")
	assert.Equal(t, node.InputAttributeAutocompleteOneTimeCode, attr.Autocomplete)
}
