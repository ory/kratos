package identity

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewCredentialsSAML(t *testing.T) {
	_, err := NewCredentialsSAML("not-empty", "")
	require.Error(t, err)

	_, err = NewCredentialsSAML("", "not-empty")
	require.Error(t, err)

	_, err = NewCredentialsSAML("not-empty", "not-empty")
	require.NoError(t, err)
}
