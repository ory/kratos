// Copyright © 2024 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package js

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToString(t *testing.T) {
	assert.Equal(t, "window.oryWebAuthnRegistration", WebAuthnTriggersWebAuthnRegistration.String())
}
