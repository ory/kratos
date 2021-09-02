package sql_runner

import (
	"embed"
	_ "embed"
	"fmt"
	"github.com/ory/kratos/driver"
	"github.com/ory/x/configx"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

//go:embed *.sql
var sqlFiles embed.FS

// runSQLCmd represents the migrate command
var runSQLCmd = &cobra.Command{
	Use:   "run-sql",
	Short: "Run SQL commands",
	Run: func(cmd *cobra.Command, args []string) {
		d := initDB(cmd)

		if len(args) != 1 {
			fmt.Println("Specify SQL file name")
			os.Exit(1)
			return
		}

		filename := args[0]
		fmt.Printf("Reading file %s\n\n", filename)

		data, err := sqlFiles.ReadFile(filename)

		if err != nil {
			d.Logger().WithError(err).Fatalf("Error reading file %s", filename)
			os.Exit(1)
			return
		}

		queries := strings.Split(strings.TrimSpace(string(data)), "\n")

		for idx, query := range queries {
			fmt.Printf("Query %d:\n", idx)
			fmt.Println(query)
			fmt.Println()
		}

		for idx, query := range queries {
			fmt.Printf("Running query %d:\n", idx)
			fmt.Println(query)

			cnt, err := d.Persister().GetConnection(cmd.Context()).RawQuery(query).ExecWithCount()

			if err != nil {
				d.Logger().WithError(err).Fatalf("Error")
				os.Exit(1)
				return
			} else {
				fmt.Println("Success")
			}

			fmt.Printf("Number of affected rows: %d\n\n", cnt)
		}

		fmt.Println("done")
	},
}

func initDB(cmd *cobra.Command) driver.Registry {
	d := driver.NewWithoutInit(
		cmd.Context(),
		configx.WithFlags(cmd.Flags()),
		configx.SkipValidation())

	err := d.Init(cmd.Context(), driver.SkipNetworkInit)

	if err != nil {
		d.Logger().WithError(err).Fatal("Unable to instantiate configuration.")
	}

	p := d.Persister()
	err = p.Ping()

	if err != nil {
		d.Logger().WithError(err).Fatal("DB ping failed")
		os.Exit(1)
	}

	return d
}

func RegisterCommandRecursive(parent *cobra.Command) {
	parent.AddCommand(runSQLCmd)
}

func init() {
	configx.RegisterFlags(runSQLCmd.PersistentFlags())
}
