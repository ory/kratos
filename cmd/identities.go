package cmd

import (
	"github.com/spf13/cobra"
)

// identitiesCmd represents the identity command
var identitiesCmd = &cobra.Command{
	Use: "identities",
}

func init() {
	rootCmd.AddCommand(identitiesCmd)

	identitiesCmd.PersistentFlags().String("endpoint", "", "Specifies the Ory Hive Admin URL. Defaults to HIVE_URLS_ADMIN")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// identitiesCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// identitiesCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
