package identities

import (
	"fmt"

	"github.com/spf13/cobra"
)

var patchCmd = &cobra.Command{
	Use:  "patch <file.json [file-2.json [file-3.json] ...]>",
	Short: "Patch identities by ID (not yet implemented)",
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// TODO
		fmt.Println("not yet implemented")
	},
}
