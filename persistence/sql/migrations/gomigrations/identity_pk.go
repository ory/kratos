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

var ChangeAddressesPK = []popx.Migration{
	{
		Version:   "20241001000000000000",
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
		Version:   "20241001000000000000",
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
		Version:   "20241001000000000001",
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
		Version:   "20241001000000000001",
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
}
