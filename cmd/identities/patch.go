package identities

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func NewPatchCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "patch <file.json [file-2.json [file-3.json] ...]>",
		Short: "Patch identities by ID (not yet implemented)",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			// TODO
			fmt.Println("not yet implemented")
			os.Exit(1)
		},
	}
}
