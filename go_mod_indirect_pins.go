// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

//go:build tools
// +build tools

package main

import (
	_ "github.com/cortesi/modd/cmd/modd"
	_ "github.com/go-swagger/go-swagger/cmd/swagger"
	_ "github.com/mailhog/MailHog"
	_ "github.com/mattn/goveralls"

	_ "github.com/ory/go-acc"
)
