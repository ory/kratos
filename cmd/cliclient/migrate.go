// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package cliclient

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/ory/x/servicelocatorx"

	"github.com/pkg/errors"

	"github.com/ory/x/contextx"

	"github.com/ory/x/configx"

	"github.com/spf13/cobra"

	"github.com/ory/kratos/driver"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/x/cmdx"
	"github.com/ory/x/flagx"
)

type MigrateHandler struct{}

func NewMigrateHandler() *MigrateHandler {
	return &MigrateHandler{}
}

func (h *MigrateHandler) MigrateSQL(cmd *cobra.Command, args []string, opts ...driver.RegistryOption) error {
	var d driver.Registry
	var err error

	if flagx.MustGetBool(cmd, "read-from-env") {
		d, err = driver.NewWithoutInit(
			cmd.Context(),
			cmd.ErrOrStderr(),
			servicelocatorx.NewOptions(),
			nil,
			[]configx.OptionModifier{
				configx.WithFlags(cmd.Flags()),
				configx.SkipValidation(),
			})
		if err != nil {
			return err
		}
		if len(d.Config().DSN(cmd.Context())) == 0 {
			fmt.Println(cmd.UsageString())
			fmt.Println("")
			fmt.Println("When using flag -e, environment variable DSN must be set")
			return cmdx.FailSilently(cmd)
		}
		if err != nil {
			return err
		}
	} else {
		if len(args) != 1 {
			fmt.Println(cmd.UsageString())
			return cmdx.FailSilently(cmd)
		}
		d, err = driver.NewWithoutInit(
			cmd.Context(),
			cmd.ErrOrStderr(),
			servicelocatorx.NewOptions(),
			nil,
			[]configx.OptionModifier{
				configx.WithFlags(cmd.Flags()),
				configx.SkipValidation(),
				configx.WithValue(config.ViperKeyDSN, args[0]),
			})
		if err != nil {
			return err
		}
	}

	err = d.Init(cmd.Context(), &contextx.Default{}, append(opts, driver.SkipNetworkInit)...)
	if err != nil {
		return errors.Wrap(err, "an error occurred initializing migrations")
	}

	var plan bytes.Buffer
	_, err = d.Persister().MigrationStatus(cmd.Context())
	if err != nil {
		return errors.Wrap(err, "an error occurred planning migrations:")
	}

	if !flagx.MustGetBool(cmd, "yes") {
		fmt.Println("The following migration is planned:")
		fmt.Println("")
		fmt.Printf("%s", plan.String())
		fmt.Println("")
		fmt.Println("To skip the next question use flag --yes (at your own risk).")
		if !askForConfirmation("Do you wish to execute this migration plan?") {
			fmt.Println("Migration aborted.")
			return cmdx.FailSilently(cmd)
		}
	}

	if err = d.Persister().MigrateUp(cmd.Context()); err != nil {
		return err
	}
	fmt.Println("Successfully applied SQL migrations!")
	return nil
}

func askForConfirmation(s string) bool {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf("%s [y/n]: ", s)

		response, err := reader.ReadString('\n')
		cmdx.Must(err, "%s", err)

		response = strings.ToLower(strings.TrimSpace(response))
		if response == "y" || response == "yes" {
			return true
		} else if response == "n" || response == "no" {
			return false
		}
	}
}
