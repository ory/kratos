// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package flow

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestState(t *testing.T) {
	assert.EqualValues(t, StateEmailSent, NextState(StateChooseMethod))
	assert.EqualValues(t, StatePassedChallenge, NextState(StateEmailSent))
	assert.EqualValues(t, StatePassedChallenge, NextState(StatePassedChallenge))

	assert.True(t, HasReachedState(StatePassedChallenge, StatePassedChallenge))
	assert.False(t, HasReachedState(StatePassedChallenge, StateEmailSent))
	assert.False(t, HasReachedState(StateEmailSent, StateChooseMethod))
}
