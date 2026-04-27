// Copyright © 2025 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"github.com/spf13/cobra"

	"github.com/ory/kratos/driver"
	"github.com/ory/x/configx"
)

func NewConfigCmd() *cobra.Command {
	c := &cobra.Command{
		Use: "config",
	}
	configx.RegisterFlags(c.PersistentFlags())
	return c
}

func RegisterCommandRecursive(parent *cobra.Command, dOpts []driver.RegistryOption) {
	c := NewConfigCmd()
	parent.AddCommand(c)
	c.AddCommand(NewConfigViewCmd(dOpts))
}
