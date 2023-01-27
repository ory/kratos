// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

//go:build tools
// +build tools

package main

import (
	_ "github.com/go-swagger/go-swagger/cmd/swagger"
	_ "github.com/mattn/goveralls"
	_ "github.com/sqs/goreturns"
	_ "golang.org/x/tools/cmd/cover"

	_ "github.com/gobuffalo/fizz"

	_ "github.com/ory/go-acc"

	_ "github.com/jteeuwen/go-bindata"

	_ "github.com/davidrjonas/semver-cli"

	_ "github.com/cortesi/modd/cmd/modd"
	_ "github.com/hashicorp/consul/api"

	_ "github.com/mikefarah/yq/v4"
)
