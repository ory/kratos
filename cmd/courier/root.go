package courier

import (
	"github.com/ory/x/configx"
	"github.com/spf13/cobra"
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
