// Copyright © 2024 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package login

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ory/kratos/identity"
)

func TestWithIdentityHint(t *testing.T) {
	expected := new(identity.Identity)
	opts := NewFormHydratorOptions([]FormHydratorModifier{WithIdentityHint(expected)})
	assert.Equal(t, expected, opts.IdentityHint)
}
