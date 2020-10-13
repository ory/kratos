// +build !refresh

package migratest

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func writeFixtureOnError(t *testing.T, err error, actual interface{}, location string) {
	require.NoError(t, err)
}
