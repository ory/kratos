package port

import (
	"github.com/spf13/cobra"

	"github.com/ory/kratos/cmd/client"
)

// importCmd represents the import command
var importCmd = &cobra.Command{
	Use: "import <file.json [file-2.json [file-3.json] ...]>",
	Run: client.NewIdentityClient().Import,
}

func RegisterCommandRecursive(parent *cobra.Command) {
	parent.AddCommand(importCmd)
}
