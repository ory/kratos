package cliclient

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strings"

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

func (h *MigrateHandler) MigrateSQL(cmd *cobra.Command, args []string) {
	var d driver.Registry

	if flagx.MustGetBool(cmd, "read-from-env") {
		d = driver.NewWithoutInit(
			cmd.Context(),
			configx.WithFlags(cmd.Flags()),
			configx.SkipValidation())
		if len(d.Config(cmd.Context()).DSN()) == 0 {
			fmt.Println(cmd.UsageString())
			fmt.Println("")
			fmt.Println("When using flag -e, environment variable DSN must be set")
			os.Exit(1)
			return
		}
	} else {
		if len(args) != 1 {
			fmt.Println(cmd.UsageString())
			os.Exit(1)
			return
		}
		d = driver.NewWithoutInit(
			cmd.Context(),
			configx.WithFlags(cmd.Flags()),
			configx.SkipValidation(),
			configx.WithValue(config.ViperKeyDSN, args[0]))
	}

	err := d.Init(cmd.Context(), driver.SkipNetworkInit)
	cmdx.Must(err, "An error occurred planning migrations: %s", err)

	var plan bytes.Buffer
	statuses, err := d.Persister().MigrationStatus(cmd.Context())
	cmdx.Must(err, "An error occurred planning migrations: %s", err)
	cmdx.Must(err, "An error occurred planning migrations: %s", statuses.Write(&plan))

	fmt.Println("The following migration is planned:")
	fmt.Println("")
	fmt.Printf("%s", plan.String())

	if !flagx.MustGetBool(cmd, "yes") {
		fmt.Println("")
		fmt.Println("To skip the next question use flag --yes (at your own risk).")
		if !askForConfirmation("Do you wish to execute this migration plan?") {
			fmt.Println("Migration aborted.")
			return
		}
	}

	err = d.Persister().MigrateUp(cmd.Context())
	cmdx.Must(err, "An error occurred while connecting to SQL: %s", err)
	fmt.Println("Successfully applied SQL migrations!")
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
