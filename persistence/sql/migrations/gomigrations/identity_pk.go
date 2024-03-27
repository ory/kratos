// Copyright © 2024 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package gomigrations

import (
	"fmt"
	"path/filepath"
	"runtime"

	"github.com/gobuffalo/pop/v6"
	"github.com/pkg/errors"

	"github.com/ory/x/popx"
)

func path() string {
	_, file, line, _ := runtime.Caller(1)
	return fmt.Sprintf("%s:%d", filepath.Base(file), line)
}

var IdentityPrimaryKeysStep1 = []popx.Migration{
	{
		Version:   "20240208000000000000",
		Path:      path(),
		Name:      "Change primary key for identity_verifiable_addresses",
		Direction: "up",
		Type:      "go",
		DBType:    "cockroach",
		RunnerNoTx: func(m popx.Migration, c *pop.Connection) error {
			_, err := c.Store.Exec("ALTER TABLE identity_verifiable_addresses ALTER PRIMARY KEY USING COLUMNS (identity_id,id)")
			return errors.WithStack(err)
		},
	},
	{
		Version:   "20240208000000000000",
		Path:      path(),
		Name:      "Revert primary key for identity_verifiable_addresses",
		Direction: "down",
		Type:      "go",
		DBType:    "cockroach",
		RunnerNoTx: func(m popx.Migration, c *pop.Connection) error {
			_, err := c.Store.Exec("ALTER TABLE identity_verifiable_addresses ALTER PRIMARY KEY USING COLUMNS (id)")
			return errors.WithStack(err)
		},
	},
	{
		Version:   "20240208000000000001",
		Path:      path(),
		Name:      "Change primary key for identity_recovery_addresses",
		Direction: "up",
		Type:      "go",
		DBType:    "cockroach",
		RunnerNoTx: func(m popx.Migration, c *pop.Connection) error {
			_, err := c.Store.Exec("ALTER TABLE identity_recovery_addresses ALTER PRIMARY KEY USING COLUMNS (identity_id,id)")
			return errors.WithStack(err)
		},
	},
	{
		Version:   "20240208000000000001",
		Path:      path(),
		Name:      "Revert primary key for identity_recovery_addresses",
		Direction: "down",
		Type:      "go",
		DBType:    "cockroach",
		RunnerNoTx: func(m popx.Migration, c *pop.Connection) error {
			_, err := c.Store.Exec("ALTER TABLE identity_recovery_addresses ALTER PRIMARY KEY USING COLUMNS (id)")
			return errors.WithStack(err)
		},
	},
	{
		Version:   "20240208000000000002",
		Path:      path(),
		Name:      "Change primary key for identity_credentials",
		Direction: "up",
		Type:      "go",
		DBType:    "cockroach",
		RunnerNoTx: func(m popx.Migration, c *pop.Connection) error {
			_, err := c.Store.Exec("ALTER TABLE identity_credentials ALTER PRIMARY KEY USING COLUMNS (identity_id,id)")
			return errors.WithStack(err)
		},
	},
	{
		Version:   "20240208000000000002",
		Path:      path(),
		Name:      "Revert primary key for identity_credentials",
		Direction: "down",
		Type:      "go",
		DBType:    "cockroach",
		RunnerNoTx: func(m popx.Migration, c *pop.Connection) error {
			_, err := c.Store.Exec("ALTER TABLE identity_credentials ALTER PRIMARY KEY USING COLUMNS (id)")
			return errors.WithStack(err)
		},
	},
	{
		Version:   "20240208000000000003",
		Path:      path(),
		Name:      "Add column identity_id to identity_credential_identifiers and session_devices",
		Direction: "up",
		Type:      "go",
		DBType:    "cockroach",
		RunnerNoTx: func(m popx.Migration, c *pop.Connection) error {
			_, err := c.Store.Exec("ALTER TABLE identity_credential_identifiers ADD COLUMN identity_id UUID NULL REFERENCES identities(id) ON DELETE CASCADE")
			if err != nil {
				return errors.WithStack(err)
			}
			_, err = c.Store.Exec("ALTER TABLE session_devices ADD COLUMN identity_id UUID NULL REFERENCES identities(id) ON DELETE CASCADE")
			return errors.WithStack(err)
		},
	},
	{
		Version:   "20240208000000000003",
		Path:      path(),
		Name:      "Drop column identity_id to identity_credential_identifiers and session_devices",
		Direction: "down",
		Type:      "go",
		DBType:    "cockroach",
		RunnerNoTx: func(m popx.Migration, c *pop.Connection) error {
			_, err := c.Store.Exec("ALTER TABLE identity_credential_identifiers DROP COLUMN identity_id")
			if err != nil {
				return errors.WithStack(err)
			}
			_, err = c.Store.Exec("ALTER TABLE session_devices DROP COLUMN identity_id")
			return errors.WithStack(err)
		},
	},
}

var IdentityPrimaryKeysStep2 = []popx.Migration{
	{
		Version:   "20240208000000000004",
		Path:      path(),
		Name:      "Backfill column identity_id in identity_credential_identifiers",
		Direction: "up",
		Type:      "go",
		DBType:    "cockroach",
		RunnerNoTx: func(m popx.Migration, c *pop.Connection) error {
			for {
				res, err := c.Store.Exec(`
					UPDATE
						identity_credential_identifiers ici
					SET
						identity_id = ic.identity_id
					FROM
						identity_credentials ic
					WHERE
						ici.identity_credential_id = ic.id
						AND ici.nid = ic.nid
						AND ici.identity_id IS NULL
					LIMIT 100`)
				if err != nil {
					return errors.WithStack(err)
				}
				n, err := res.RowsAffected()
				if err != nil {
					return errors.WithStack(err)
				}
				if n == 0 {
					break
				}
				fmt.Printf("Backfilled %d rows in identity_credential_identifiers\n", n)
			}
			return nil
		},
	},
	{
		Version:   "20240208000000000004",
		Path:      path(),
		Name:      "Revert backfill column identity_id in identity_credential_identifiers (noop)",
		Direction: "down",
		Type:      "go",
		DBType:    "cockroach",
		RunnerNoTx: func(m popx.Migration, c *pop.Connection) error {
			// nothing
			return nil
		},
	},
	{
		Version:   "20240208000000000005",
		Path:      path(),
		Name:      "Change primary key of identity_credential_identifiers",
		Direction: "up",
		Type:      "go",
		DBType:    "cockroach",
		RunnerNoTx: func(m popx.Migration, c *pop.Connection) error {
			_, err := c.Store.Exec("ALTER TABLE identity_credential_identifiers ALTER identity_id SET NOT NULL")
			if err != nil {
				return errors.WithStack(err)
			}
			_, err = c.Store.Exec("ALTER TABLE identity_credential_identifiers ALTER PRIMARY KEY USING COLUMNS (identity_id, identity_credential_id, id)")
			return errors.WithStack(err)
		},
	},
	{
		Version:   "20240208000000000005",
		Path:      path(),
		Name:      "Revert primary key of identity_credential_identifiers",
		Direction: "down",
		Type:      "go",
		DBType:    "cockroach",
		RunnerNoTx: func(m popx.Migration, c *pop.Connection) error {
			_, err := c.Store.Exec("ALTER TABLE identity_credential_identifiers ALTER PRIMARY KEY USING COLUMNS (id)")
			if err != nil {
				return errors.WithStack(err)
			}
			_, err = c.Store.Exec("ALTER TABLE identity_credential_identifiers ALTER identity_id DROP NOT NULL")
			return errors.WithStack(err)
		},
	},
	{
		Version:   "20240208000000000006",
		Path:      path(),
		Name:      "Backfill column identity_id in session_devices",
		Direction: "up",
		Type:      "go",
		DBType:    "cockroach",
		RunnerNoTx: func(m popx.Migration, c *pop.Connection) error {
			for {
				res, err := c.Store.Exec(`
					UPDATE
						session_devices sd
					SET
						identity_id = s.identity_id
					FROM
						sessions s
					WHERE
						sd.session_id = s.id
						AND sd.nid = s.nid
						AND sd.identity_id IS NULL
					LIMIT 100`)
				if err != nil {
					return errors.WithStack(err)
				}
				n, err := res.RowsAffected()
				if err != nil {
					return errors.WithStack(err)
				}
				if n == 0 {
					break
				}
				fmt.Printf("Backfilled %d rows in session_devices\n", n)
			}
			return nil
		},
	},
	{
		Version:   "20240208000000000006",
		Path:      path(),
		Name:      "Revert backfill column identity_id in Backfill column identity_id in session_devices (noop)",
		Direction: "down",
		Type:      "go",
		DBType:    "cockroach",
		RunnerNoTx: func(m popx.Migration, c *pop.Connection) error {
			// nothing
			return nil
		},
	},
	{
		Version:   "20240208000000000007",
		Path:      path(),
		Name:      "Change primary key of session_devices",
		Direction: "up",
		Type:      "go",
		DBType:    "cockroach",
		RunnerNoTx: func(m popx.Migration, c *pop.Connection) error {
			_, err := c.Store.Exec("ALTER TABLE session_devices ALTER identity_id SET NOT NULL")
			if err != nil {
				return errors.WithStack(err)
			}
			_, err = c.Store.Exec("ALTER TABLE session_devices ALTER PRIMARY KEY USING COLUMNS (session_id, identity_id, id)")
			return errors.WithStack(err)
		},
	},
	{
		Version:   "20240208000000000007",
		Path:      path(),
		Name:      "Revert primary key of session_devices",
		Direction: "down",
		Type:      "go",
		DBType:    "cockroach",
		RunnerNoTx: func(m popx.Migration, c *pop.Connection) error {
			_, err := c.Store.Exec("ALTER TABLE session_devices ALTER PRIMARY KEY USING COLUMNS (id)")
			if err != nil {
				return errors.WithStack(err)
			}
			_, err = c.Store.Exec("ALTER TABLE session_devices ALTER identity_id DROP NOT NULL")
			return errors.WithStack(err)
		},
	},
}
