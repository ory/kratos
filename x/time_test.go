package x

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestRandomDelay(t *testing.T) {
	base := time.Millisecond * 20
	deviation := time.Millisecond * 5
	s := time.Now()
	RandomDelay(base, deviation)
	e := time.Now()
	elapsed := e.Sub(s)
	require.LessOrEqual(t, elapsed, base+deviation)
	require.GreaterOrEqual(t, elapsed, base-deviation)
}
