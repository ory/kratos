// Copyright © 2026 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package migrate

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/ory/kratos/cmd/cliclient"
	"github.com/ory/kratos/driver"
	"github.com/ory/x/configx"
)

func NewNormalizePhoneCmd(opts ...driver.RegistryOption) *cobra.Command {
	c := &cobra.Command{
		Use:   "normalize-phone-numbers [database-url]",
		Short: "Normalize phone numbers to E.164 format in the database",
		Long: `Normalizes all phone numbers in identity_credential_identifiers,
identity_verifiable_addresses, and identity_recovery_addresses to E.164 format.

This command uses keyset pagination to iterate over the database in batches.
It is safe to run multiple times (idempotent) and can be interrupted and resumed
using the --start-after flag with the last ID printed in the progress output.

Run this command AFTER deploying the code changes that normalize phone numbers
on write, to ensure all legacy data is also normalized.

You can read in the database URL using the -e flag, for example:
	export DSN=...
	kratos migrate normalize-phone-numbers -e

### WARNING ###
Before running this command on an existing database, create a back up!
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := cliclient.NewNormalizePhoneHandler().NormalizePhoneNumbers(cmd, args, opts...)
			if err != nil {
				_, _ = fmt.Fprintln(cmd.ErrOrStderr(), err)
				return err
			}
			return nil
		},
	}

	configx.RegisterFlags(c.PersistentFlags())
	c.Flags().BoolP("read-from-env", "e", false, "If set, reads the database connection string from the environment variable DSN or config file key dsn.")
	c.Flags().IntP("batch-size", "b", 1000, "Number of rows to process per batch")
	c.Flags().Bool("dry-run", false, "If set, only report what would change without writing")
	c.Flags().Duration("batch-delay", time.Second, "Delay between batches to reduce database load (e.g. 100ms, 1s)")
	c.Flags().StringSlice("start-after", nil, "Resume after a table's last processed ID, e.g. --start-after credentials=<id> --start-after verifiable=<id>")

	return c
}
