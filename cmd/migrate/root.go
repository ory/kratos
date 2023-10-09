// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package migrate

import (
	"github.com/spf13/cobra"
)

func NewMigrateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "migrate",
		Short: "Various migration helpers",
	}
}

func RegisterCommandRecursive(parent *cobra.Command) {
	c := NewMigrateCmd()
	parent.AddCommand(c)
	c.AddCommand(NewMigrateSQLCmd())
}
