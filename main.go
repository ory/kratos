// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

// package main is the entry point for kratos.
package main

import (
	"github.com/ory/kratos/driver"
	"github.com/ory/x/dbal"
	"github.com/ory/x/profilex"

	"github.com/ory/kratos/cmd"
)

func main() {
	defer profilex.Profile().Stop()
	dbal.RegisterDriver(func() dbal.Driver {
		return driver.NewRegistryDefault()
	})

	cmd.Execute()
}
