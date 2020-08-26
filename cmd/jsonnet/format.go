/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package jsonnet

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/ory/kratos/internal/clihelpers"

	"github.com/google/go-jsonnet/formatter"
	"github.com/spf13/cobra"

	"github.com/ory/x/cmdx"
	"github.com/ory/x/flagx"
)

// jsonnetFormatCmd represents the format command
var jsonnetFormatCmd = &cobra.Command{
	Use: "format path/to/files/*.jsonnet [more/files.jsonnet, [supports/**/{foo,bar}.jsonnet]]",
	Long: `Formats JSONNet files using the official JSONNet formatter.

Use -w or --write to write output back to files instead of stdout.

` + clihelpers.JsonnetGlobHelp,
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
					err := ioutil.WriteFile(file, []byte(output), 0644)
					cmdx.Must(err, `Could not write to file "%s" because: %s`, file, err)
				} else {
					fmt.Println(output)
				}
			}
		}
	},
}

func init() {
	jsonnetFormatCmd.Flags().BoolP("write", "w", false, "Write formatted output back to file.")
}
