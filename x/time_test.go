package x

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestRandomDelay(t *testing.T) {
	base := time.Millisecond * 100
	deviation := time.Millisecond * 50
	epsilon := time.Duration(float64(deviation) * .1)
	for i := 0; i < 100; i++ {
		delay := RandomDelay(base, deviation)
		require.LessOrEqual(t, delay, base+deviation+epsilon)
		require.GreaterOrEqual(t, delay, base-deviation-epsilon)
	}
}
