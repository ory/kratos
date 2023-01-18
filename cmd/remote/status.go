// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package remote

import (
	"github.com/ory/x/cmdx"

	"github.com/spf13/cobra"

	"github.com/ory/kratos/cmd/cliclient"
)

type statusState struct {
	Alive bool `json:"alive"`
	Ready bool `json:"ready"`
}

func (s *statusState) Header() []string {
	return []string{"ALIVE", "READY"}
}

func (s *statusState) Columns() []string {
	f := [2]string{
		"false",
		"false",
	}
	if s.Alive {
		f[0] = "true"
	}
	if s.Ready {
		f[1] = "true"
	}
	return f[:]
}

func (s *statusState) Interface() interface{} {
	return s
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Print the alive and readiness status of a Ory Kratos instance",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := cliclient.NewClient(cmd)
		if err != nil {
			return err
		}

		state := &statusState{}
		defer cmdx.PrintRow(cmd, state)

		alive, _, err := c.MetadataApi.IsAlive(cmd.Context()).Execute()
		if err != nil {
			return err
		}

		state.Alive = alive.Status == "ok"

		ready, _, err := c.MetadataApi.IsReady(cmd.Context()).Execute()
		if err != nil {
			return err
		}

		state.Ready = ready.Status == "ok"
		return nil
	},
}
