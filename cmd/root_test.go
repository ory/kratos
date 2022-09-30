package cmd

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func assertUsageWorks(t *testing.T, cmd *cobra.Command) {
	var usage string
	require.NotPanics(t, func() {
		usage = cmd.UsageString()
	})
	assert.NotContains(t, usage, "{{")
	for _, child := range cmd.Commands() {
		assertUsageWorks(t, child)
	}
}

func TestUsageStrings(t *testing.T) {
	assertUsageWorks(t, NewRootCmd())
}
