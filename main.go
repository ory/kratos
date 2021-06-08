// Copyright © 2018 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// package main is the entry point for kratos.
package main

import (
	"github.com/ory/kratos/corp"
	"github.com/ory/kratos/driver"
	"github.com/ory/x/dbal"
	"github.com/ory/x/profilex"

	"github.com/ory/kratos/cmd"
)


func main() {
	corp.SetContextualizer(new(corp.ContextNoOp))

	defer profilex.Profile().Stop()
	dbal.RegisterDriver(func() dbal.Driver {
		return driver.NewRegistryDefault()
	})

	cmd.Execute()
}
