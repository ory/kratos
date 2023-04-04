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

func performRecovery(email string) *ory.RecoveryFlow {
	ctx := context.Background()

	// Initialize the flow
	flow, res, err := client.FrontendApi.CreateNativeRecoveryFlow(ctx).Execute()
	pkg.SDKExitOnError(err, res)

	// If you want, print the flow here:
	//
	//	pkg.PrintJSONPretty(flow)

	// Submit the form
	afterSubmit, res, err := client.FrontendApi.UpdateRecoveryFlow(ctx).Flow(flow.Id).
		UpdateRecoveryFlowBody(ory.UpdateRecoveryFlowWithLinkMethodAsUpdateRecoveryFlowBody(&ory.UpdateRecoveryFlowWithLinkMethod{
			Email:  email,
			Method: "link",
		})).Execute()
	pkg.SDKExitOnError(err, res)

	return afterSubmit
}

func main() {
	pkg.PrintJSONPretty(
		performRecovery("someone@foobar.com"),
	)
}
