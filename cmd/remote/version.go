// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package remote

import (
	"github.com/spf13/cobra"

	"github.com/ory/kratos/cmd/cliclient"
	"github.com/ory/x/cmdx"
)

type versionValue struct {
	Version string `json:"version"`
}

func (v *versionValue) Header() []string {
	return []string{"VERSION"}
}

func (v *versionValue) Columns() []string {
	return []string{v.Version}
}

func (v *versionValue) Interface() interface{} {
	return v
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version of an Ory Kratos instance",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := cliclient.NewClient(cmd)
		if err != nil {
			return err
		}

		resp, _, err := c.MetadataApi.GetVersion(cmd.Context()).Execute()
		if err != nil {
			return err
		}

		cmdx.PrintRow(cmd, &versionValue{Version: resp.Version})
		return nil
	},
}
