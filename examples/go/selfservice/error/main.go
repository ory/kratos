// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"

	"github.com/ory/kratos/examples/go/pkg"

	ory "github.com/ory/client-go"
)

// If you use Open Source this would be:
//
// var client = pkg.NewSDKForSelfHosted("http://127.0.0.1:4433")
var client = pkg.NewSDK("playground")

func getError() *ory.FlowError {
	e, res, err := client.FrontendApi.GetFlowError(context.Background()).Id("stub:500").Execute()
	pkg.SDKExitOnError(err, res)
	return e
}

func main() {
	pkg.PrintJSONPretty(
		getError(),
	)
}
