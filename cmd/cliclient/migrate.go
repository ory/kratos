package cliclient

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ory/viper"
	"github.com/ory/x/cmdx"
	"github.com/ory/x/flagx"
	"github.com/ory/x/logrusx"

	"github.com/ory/kratos/driver"
	"github.com/ory/kratos/driver/configuration"
)

type MigrateHandler struct{}

func NewMigrateHandler() *MigrateHandler {
	return &MigrateHandler{}
}

func (h *MigrateHandler) MigrateSQL(cmd *cobra.Command, args []string) {
	var d driver.Driver

	logger := logrusx.New("ORY Kratos", cmd.Version)
	if flagx.MustGetBool(cmd, "read-from-env") {
		d = driver.MustNewDefaultDriver(logger, "", "", "", true)
		if len(d.Configuration().DSN()) == 0 {
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
		viper.Set(configuration.ViperKeyDSN, args[0])
		d = driver.MustNewDefaultDriver(logger, "", "", "", true)
	}

	var plan bytes.Buffer
	err := d.Registry().Persister().MigrationStatus(context.Background(), &plan)
	cmdx.Must(err, "An error occurred planning migrations: %s", err)

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

	err = d.Registry().Persister().MigrateUp(context.Background())
	cmdx.Must(err, "An error occurred while connecting to SQL: %s", err)
	fmt.Println("Successfully applied SQL migrations!")

	// if !ok {
	// 	fmt.Println(cmd.UsageString())
	// 	fmt.Println("")
	// 	fmt.Printf("Migrations can only be executed against a SQL-compatible driver but DSN is not a SQL source.\n")
	// 	os.Exit(1)
	// 	return
	// }
	//
	// scheme := sqlcon.GetDriverName(d.Configuration().DSN())
	// plan, err := reg.SchemaMigrationPlan(scheme)
	// cmdx.Must(err, "An error occurred planning migrations: %s", err)
	//
	// fmt.Println("The following migration is planned:")
	// fmt.Println("")
	// plan.Render()
	//
	// if !flagx.MustGetBool(cmd, "yes") {
	// 	fmt.Println("")
	// 	fmt.Println("To skip the next question use flag --yes (at your own risk).")
	// 	if !askForConfirmation("Do you wish to execute this migration plan?") {
	// 		fmt.Println("Migration aborted.")
	// 		return
	// 	}
	// }
	//
	// n, err := reg.CreateSchemas(scheme)
	// cmdx.Must(err, "An error occurred while connecting to SQL: %s", err)
	// fmt.Printf("Successfully applied %d SQL migrations!\n", n)
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
