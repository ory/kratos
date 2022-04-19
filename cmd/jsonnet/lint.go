package jsonnet

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/go-jsonnet"
	"github.com/google/go-jsonnet/linter"
	"github.com/spf13/cobra"

	"github.com/ory/x/cmdx"
)

func NewLintCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "lint",
		Short: "Helpers for linting code",
	}
	c.AddCommand(NewJsonnetLintCmd())
	return c
}

func NewJsonnetLintCmd() *cobra.Command {
	return &cobra.Command{
		Use: "lint path/to/files/*.jsonnet [more/files.jsonnet] [supports/**/{foo,bar}.jsonnet]",
		Long: `Lints JSONNet files using the official JSONNet linter and exits with a status code of 1 when issues are detected.

` + GlobHelp,
		Args: cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			vm := jsonnet.MakeVM()

			for _, pattern := range args {
				files, err := filepath.Glob(pattern)
				cmdx.Must(err, `Glob path "%s" is not valid: %s`, pattern, err)

				for _, file := range files {
					content, err := ioutil.ReadFile(file)
					cmdx.Must(err, `Unable to read file "%s" because: %s`, file, err)

					var outBuilder strings.Builder
					errorsFound := linter.LintSnippet(vm, &outBuilder, []linter.Snippet{{FileName: file, Code: string(content)}})

					if errorsFound {
						_, _ = fmt.Fprintf(os.Stderr, "Linter found issues.")
						os.Exit(1)
					}
				}
			}
		},
	}
}
