package cmd

import (
	"github.com/ory/x/cmdx"
)

var (
	BuildVersion = ""
	BuildTime    = ""
	BuildGitHash = ""
)

func init() {
	rootCmd.AddCommand(cmdx.Version(&BuildVersion, &BuildGitHash, &BuildTime))
}
