// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

// package main is the entry point for kratos.
package main

import (
	"log"
	"net/http"
	"os"

	"github.com/ory/kratos/cmd"

	_ "net/http/pprof"
)

func main() {
	go func() {
		log.Println(http.ListenAndServe("localhost:7777", nil))
	}()
	os.Exit(cmd.Execute())
}
