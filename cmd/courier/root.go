// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package courier

import (
	"github.com/spf13/cobra"

	"github.com/ory/kratos/driver"
	"github.com/ory/x/configx"
)

// NewCourierCmd creates a new courier command
func NewCourierCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "courier",
		Short: "Commands related to the Ory Kratos message courier",
	}
	configx.RegisterFlags(c.PersistentFlags())
	return c
}

func RegisterCommandRecursive(parent *cobra.Command, dOpts []driver.RegistryOption) {
	c := NewCourierCmd()
	parent.AddCommand(c)
	c.AddCommand(NewWatchCmd(dOpts))
}
