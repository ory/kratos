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
	courierCmd.PersistentFlags().Int("expose-metrics-port", 0, "The port to expose the metrics endpoint on (not exposed by default)")
}

func RegisterCommandRecursive(parent *cobra.Command) {
	parent.AddCommand(courierCmd)

	courierCmd.AddCommand(watchCmd)
}
