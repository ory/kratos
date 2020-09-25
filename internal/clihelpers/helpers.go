package clihelpers

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var (
	BuildVersion        = ""
	BuildTime           = ""
	BuildGitHash        = ""
	NoPrintButFailError = errors.New("this error should never be printed")
)

const (
	WarningJQIsComplicated = "We have to admit, this is not easy if you don't speak jq fluently. What about opening an issue and telling us what predefined selectors you want to have? https://github.com/ory/kratos/issues/new/choose"
)

func FailSilently(cmd *cobra.Command) error {
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	return errors.WithStack(NoPrintButFailError)
}
