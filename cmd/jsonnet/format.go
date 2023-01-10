// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package jsonnet

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/go-jsonnet/formatter"
	"github.com/spf13/cobra"

	"github.com/ory/x/cmdx"
	"github.com/ory/x/flagx"
)

func NewFormatCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "format",
		Short: "Helpers for formatting code",
	}
	c.AddCommand(NewJsonnetFormatCmd())
	return c
}

func NewJsonnetFormatCmd() *cobra.Command {
	c := &cobra.Command{
		Use: "jsonnet path/to/files/*.jsonnet [more/files.jsonnet] [supports/**/{foo,bar}.jsonnet]",
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
					content, err := os.ReadFile(file)
					cmdx.Must(err, `Unable to read file "%s" because: %s`, file, err)

					output, err := formatter.Format(file, string(content), formatter.DefaultOptions())
					cmdx.Must(err, `JSONNet file "%s" could not be formatted: %s`, file, err)

					if shouldWrite {
						err := os.WriteFile(file, []byte(output), 0644) // #nosec
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
