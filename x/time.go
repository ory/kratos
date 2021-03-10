package x

import (
	"math"
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

// RandomDelay returns a time randomly chosen from a normal distribution with mean of base and max/min of base +- deviation
// From the docstring for the rand.NormFloat64():
// To produce a different normal distribution, callers can
// adjust the output using:
//
//  sample = NormFloat64() * desiredStdDev + desiredMean
//
// Since 99.73% of values in a normal distribution lie within three standard deviations from the mean (https://en.wikipedia.org/wiki/68%E2%80%9395%E2%80%9399.7_rule),
// by taking the standard deviation to be deviation/3, we can get a distribution which fits our bounds nicely with minimal clipping when we take max/mins to cut off the tails.
func RandomDelay(base, deviation time.Duration) time.Duration {
	max := float64(base + deviation)
	min := float64(base - deviation)
	stddev := float64(deviation) / 3
	sample := rnd.NormFloat64() * (stddev + float64(base))
	boundedSample := math.Min(math.Max(sample, min), max)
	return time.Duration(boundedSample)
}
