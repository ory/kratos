package argon2

import (
	"github.com/ory/kratos/hash"
	"github.com/spf13/cobra"
	"testing"
	"time"
)

func TestCalibrateCmd(t *testing.T) {
	var lastRuns int
	var lastVerbose bool
	var retDuration time.Duration

	mockProbe := func(_ *cobra.Command, _ hash.Hasher, runs int, verbose bool) (time.Duration, error) {
		lastRuns = runs
		lastVerbose = verbose
		return retDuration, nil
	}

	newCmd := func () *cobra.Command {
		return newCalibrateCmd(mockProbe)
	}

	t.Run("", func(t *testing.T) {

	})
}
