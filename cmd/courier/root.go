package courier

import (
	"github.com/spf13/cobra"

	"github.com/ory/x/configx"
)

// courierCmd represents the courier command
var courierCmd = &cobra.Command{
	Use:   "courier",
	Short: "Commands related to the ORY Kratos message courier",
}

func init() {
	configx.RegisterFlags(courierCmd.PersistentFlags())
}

func RegisterCommandRecursive(parent *cobra.Command) {
	parent.AddCommand(courierCmd)

	courierCmd.AddCommand(watchCmd)
}
