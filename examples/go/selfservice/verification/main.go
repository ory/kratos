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

func performVerification(email string) *ory.VerificationFlow {
	ctx := context.Background()

	// Initialize the flow
	flow, res, err := client.FrontendApi.CreateNativeVerificationFlow(ctx).Execute()
	pkg.SDKExitOnError(err, res)

	// If you want, print the flow here:
	//
	//	pkg.PrintJSONPretty(flow)

	// Submit the form
	afterSubmit, res, err := client.FrontendApi.UpdateVerificationFlow(ctx).Flow(flow.Id).
		UpdateVerificationFlowBody(ory.UpdateVerificationFlowWithLinkMethodAsUpdateVerificationFlowBody(&ory.UpdateVerificationFlowWithLinkMethod{
			Email:  email,
			Method: "link",
		})).Execute()
	pkg.SDKExitOnError(err, res)

	return afterSubmit
}

func main() {
	pkg.PrintJSONPretty(
		performVerification("someone@foobar.com"),
	)
}
