package x

import (
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var rnd = rand.New(rand.NewSource(time.Now().Unix()))

func AssertEqualTime(t *testing.T, expected, actual time.Time) {
	assert.EqualValues(t, expected.UTC().Round(time.Second), actual.UTC().Round(time.Second))
}

func RequireEqualTime(t *testing.T, expected, actual time.Time) {
	require.EqualValues(t, expected.UTC().Round(time.Second), actual.UTC().Round(time.Second))
}

func RandomDelay(base, deviation time.Duration) time.Duration {
	return time.Duration(rnd.NormFloat64()*(float64(deviation)/3) + float64(base))
}
