package lint

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/ory/kratos/internal/clihelpers"

	"github.com/google/go-jsonnet"
	"github.com/google/go-jsonnet/linter"
	"github.com/spf13/cobra"

	"github.com/ory/x/cmdx"
)

// jsonnetLintCmd represents the lint command
var jsonnetLintCmd = &cobra.Command{
	Use: "lint path/to/files/*.jsonnet [more/files.jsonnet, [supports/**/{foo,bar}.jsonnet]]",
	Long: `Lints JSONNet files using the official JSONNet linter and exits with a status code of 1 when issues are detected.

` + clihelpers.JsonnetGlobHelp,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		for _, pattern := range args {
			files, err := filepath.Glob(pattern)
			cmdx.Must(err, `Glob path "%s" is not valid: %s`, pattern, err)

			for _, file := range files {
				content, err := ioutil.ReadFile(file)
				cmdx.Must(err, `Unable to read file "%s" because: %s`, file, err)

				node, err := jsonnet.SnippetToAST(file, string(content))
				cmdx.Must(err, `Unable to parse JSONNet source "%s" because: %s`, file, err)

				ew := &linter.ErrorWriter{Writer: os.Stderr}
				linter.Lint(node, ew)
				if ew.ErrorsFound {
					_, _ = fmt.Fprintf(os.Stderr, "Linter found issues.")
					os.Exit(1)
				}
			}
		}
	},
}

func RegisterCommandRecursive(parent *cobra.Command) {
	parent.AddCommand(jsonnetLintCmd)
}
