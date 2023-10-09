// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package x

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestRandomDelay(t *testing.T) {
	base := time.Millisecond * 2
	deviation := time.Millisecond
	for i := 0; i < 100; i++ {
		delay := RandomDelay(base, deviation)
		require.LessOrEqual(t, delay, base+deviation)
		require.GreaterOrEqual(t, delay, base-deviation)
	}
}
