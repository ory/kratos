package jsonnet

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/google/go-jsonnet/formatter"
	"github.com/spf13/cobra"

	"github.com/ory/x/cmdx"
	"github.com/ory/x/flagx"
)

func NewJsonnetFormatCmd() *cobra.Command {
	c := &cobra.Command{
		Use: "format path/to/files/*.jsonnet [more/files.jsonnet, [supports/**/{foo,bar}.jsonnet]]",
		Long: `Formats JSONNet files using the official JSONNet formatter.

Use -w or --write to write output back to files instead of stdout.

` + GlobHelp,
		Args: cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			for _, pattern := range args {
				files, err := filepath.Glob(pattern)
				cmdx.Must(err, `Glob path "%s" is not valid: %s`, pattern, err)

				shouldWrite := flagx.MustGetBool(cmd, "write")
				for _, file := range files {
					content, err := ioutil.ReadFile(file)
					cmdx.Must(err, `Unable to read file "%s" because: %s`, file, err)

					output, err := formatter.Format(file, string(content), formatter.DefaultOptions())
					cmdx.Must(err, `JSONNet file "%s" could not be formatted: %s`, file, err)

					if shouldWrite {
						err := ioutil.WriteFile(file, []byte(output), 0644) // #nosec
						cmdx.Must(err, `Could not write to file "%s" because: %s`, file, err)
					} else {
						fmt.Println(output)
					}
				}
			}
		},
	}
	c.Flags().BoolP("write", "w", false, "Write formatted output back to file.")
	return c
}
