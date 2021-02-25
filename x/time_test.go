package x

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestRandomDelay(t *testing.T) {
	base := time.Millisecond * 20
	deviation := time.Millisecond * 5
	start := time.Now()
	RandomDelay(base, deviation)
	delay := time.Now().Sub(start)
	require.LessOrEqual(t, delay, base+deviation)
	require.GreaterOrEqual(t, delay, base-deviation)
}
