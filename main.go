// Copyright © UBnity Co.,Ltd. All Rights Reserved.

// package main is the entry point for kratos.
package main

import (
	"fmt"
	"github.com/ory/kratos/driver"
	"github.com/ory/x/dbal"
	"github.com/ory/x/profilex"

	"github.com/ory/kratos/cmd"
)

func main() {
	fmt.Println("Running UBnity version of Kratos")

	defer profilex.Profile().Stop()
	dbal.RegisterDriver(func() dbal.Driver {
		return driver.NewRegistryDefault()
	})

	cmd.Execute()
}
