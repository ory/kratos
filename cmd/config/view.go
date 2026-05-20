// Copyright © 2025 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ory/kratos/driver"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/x/configx"
	"github.com/ory/x/contextx"
	"github.com/ory/x/logrusx"
)

func NewConfigViewCmd(dOpts []driver.RegistryOption) *cobra.Command {
	c := &cobra.Command{
		Use: "view",
		RunE: func(cmd *cobra.Command, args []string) error {
			l := logrusx.New("Ory Kratos", config.Version)
			config, err := config.New(
				cmd.Context(),
				l,
				cmd.ErrOrStderr(),
				&contextx.Default{},
				configx.WithFlags(cmd.Flags()),
				configx.SkipValidation(),
				configx.WithContext(cmd.Context()),
			)
			if err != nil {
				return err
			}

			out, err := config.View(cmd.Context())
			if err != nil {
				return err
			}

			fmt.Printf("%s\n", out)

			return nil
		},
	}

	return c
}
