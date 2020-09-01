// +build refresh

package migratest

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/require"
)

func writeFixtureOnError(t *testing.T, err error, actual interface{}, location string) {
	content, err := json.MarshalIndent(actual, "", "  ")
	require.NoError(t, err)
	require.NoError(t, ioutil.WriteFile(location, content, 0666))
}
