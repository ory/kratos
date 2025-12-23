// Copyright © 2025 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package gomigrations

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/ory/pop/v6"
	"github.com/ory/x/popx"
)

var backfillIdentityID = []popx.Migration{
	{
		Version:    "20251105000000000001",
		Path:       path(),
		Name:       "Backfill column identity_id in identity_credential_identifiers",
		Direction:  "up",
		Type:       "go",
		DBType:     "cockroach",
		Autocommit: true,
		Runner: func(m popx.Migration, c *pop.Connection) error {
			_, err := c.Store.Exec("CREATE INDEX IF NOT EXISTS ici_identity_id_backfill_storing ON identity_credential_identifiers (identity_id ASC) STORING (identity_credential_id) NOT VISIBLE")
			if err != nil {
				return errors.WithStack(err)
			}
			for {
				res, err := c.Store.Exec(`
					UPDATE
						identity_credential_identifiers@ici_identity_id_backfill_storing ici
					SET
						identity_id = ic.identity_id
					FROM
						identity_credentials ic
					WHERE
						ici.identity_credential_id = ic.id
						-- AND ici.nid = ic.nid -- not needed because JOIN predicate is on primary key
						AND ici.identity_id IS NULL
					LIMIT 10000`)
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
				fmt.Printf("Backfilled column identity_id for %d rows in identity_credential_identifiers table\n", n)
			}
			_, err = c.Store.Exec("DROP INDEX identity_credential_identifiers@ici_identity_id_backfill_storing")
			return errors.WithStack(err)
		},
	},
	{
		Version:    "20251105000000000001",
		Path:       path(),
		Name:       "Backfill column identity_id in identity_credential_identifiers",
		Direction:  "up",
		Type:       "go",
		DBType:     "postgres",
		Autocommit: true,
		Runner: func(m popx.Migration, c *pop.Connection) error {
			for {
				res, err := c.Store.Exec(`
					WITH to_update AS (
						SELECT ici.id, ic.identity_id
						FROM identity_credential_identifiers ici
						JOIN identity_credentials ic ON ici.identity_credential_id = ic.id AND ici.nid = ic.nid
						WHERE ici.identity_id IS NULL
						LIMIT 10000
					)
					UPDATE identity_credential_identifiers ici
					SET identity_id = to_update.identity_id
					FROM to_update
					WHERE ici.id = to_update.id`)
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
				fmt.Printf("Backfilled column identity_id for %d rows in identity_credential_identifiers table\n", n)
			}
			return nil
		},
	},
	{
		Version:    "20251105000000000001",
		Path:       path(),
		Name:       "Backfill column identity_id in identity_credential_identifiers",
		Direction:  "up",
		Type:       "go",
		DBType:     "mysql",
		Autocommit: true,
		Runner: func(m popx.Migration, c *pop.Connection) error {
			for {
				res, err := c.Store.Exec(`
					UPDATE identity_credential_identifiers ici
					JOIN (
						SELECT ici2.id
						FROM identity_credential_identifiers ici2
						JOIN identity_credentials ic ON ici2.identity_credential_id = ic.id AND ici2.nid = ic.nid
						WHERE ici2.identity_id IS NULL
						LIMIT 10000
					) t ON ici.id = t.id
					JOIN identity_credentials ic ON ici.identity_credential_id = ic.id
					SET ici.identity_id = ic.identity_id`)
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
				fmt.Printf("Backfilled column identity_id for %d rows in identity_credential_identifiers table\n", n)
			}
			return nil
		},
	},
	{
		Version:    "20251105000000000001",
		Path:       path(),
		Name:       "Backfill column identity_id in identity_credential_identifiers (noop)",
		Direction:  "up",
		Type:       "go",
		DBType:     "sqlite3",
		Autocommit: false,
		Runner: func(m popx.Migration, c *pop.Connection) error {
			return nil // nothing, see 20251104000000000000_identifiers_devices_identity_id.sqlite.up.sql
		},
	},
	{
		Version:   "20251105000000000001",
		Path:      path(),
		Name:      "Revert backfill column identity_id in identity_credential_identifiers (noop)",
		Direction: "down",
		Type:      "go",
		DBType:    "all",
		Runner: func(m popx.Migration, c *pop.Connection) error {
			return nil
		},
	},
	{
		Version:    "20251105000000000002",
		Path:       path(),
		Name:       "Backfill column identity_id in session_devices",
		Direction:  "up",
		Type:       "go",
		DBType:     "cockroach",
		Autocommit: true,
		Runner: func(m popx.Migration, c *pop.Connection) error {
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
					LIMIT 10000`)
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
				fmt.Printf("Backfilled column identity_id for %d rows in session_devices table\n", n)
			}
			return nil
		},
	},
	{
		Version:    "20251105000000000002",
		Path:       path(),
		Name:       "Backfill column identity_id in session_devices",
		Direction:  "up",
		Type:       "go",
		DBType:     "postgres",
		Autocommit: true,
		Runner: func(m popx.Migration, c *pop.Connection) error {
			for {
				res, err := c.Store.Exec(`
					WITH to_update AS (
						SELECT sd.id, s.identity_id
						FROM session_devices sd
						JOIN sessions s ON sd.session_id = s.id AND sd.nid = s.nid
						WHERE sd.identity_id IS NULL
						LIMIT 10000
					)
					UPDATE session_devices sd
					SET identity_id = to_update.identity_id
					FROM to_update
					WHERE sd.id = to_update.id`)
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
				fmt.Printf("Backfilled column identity_id for %d rows in session_devices table\n", n)
			}
			return nil
		},
	},
	{
		Version:    "20251105000000000002",
		Path:       path(),
		Name:       "Backfill column identity_id in session_devices",
		Direction:  "up",
		Type:       "go",
		DBType:     "mysql",
		Autocommit: true,
		Runner: func(m popx.Migration, c *pop.Connection) error {
			for {
				res, err := c.Store.Exec(`
					UPDATE session_devices sd
					JOIN (
						SELECT sd2.id
						FROM session_devices sd2
						JOIN sessions s2 ON sd2.session_id = s2.id AND sd2.nid = s2.nid
						WHERE sd2.identity_id IS NULL
						LIMIT 10000
					) t ON sd.id = t.id
					JOIN sessions s ON sd.session_id = s.id
					SET sd.identity_id = s.identity_id`)
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
				fmt.Printf("Backfilled column identity_id for %d rows in session_devices table\n", n)
			}
			return nil
		},
	},
	{
		Version:    "20251105000000000002",
		Path:       path(),
		Name:       "Backfill column identity_id in session_devices (noop)",
		Direction:  "up",
		Type:       "go",
		DBType:     "sqlite3",
		Autocommit: false,
		Runner: func(m popx.Migration, c *pop.Connection) error {
			return nil // nothing, see 20251104000000000000_identifiers_devices_identity_id.sqlite.up.sql
		},
	},

	{
		Version:    "20251105000000000002",
		Path:       path(),
		Name:       "Revert backfill column identity_id in session_devices (noop)",
		Direction:  "down",
		Type:       "go",
		DBType:     "all",
		Autocommit: false,
		Runner: func(m popx.Migration, c *pop.Connection) error {
			return nil
		},
	},
}
