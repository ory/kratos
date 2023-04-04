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

func performLogin() *ory.SuccessfulNativeLogin {
	ctx := context.Background()

	// Create a temporary user
	email, password := pkg.RandomCredentials()
	_, _ = pkg.CreateIdentityWithSession(client, email, password)

	// Initialize the flow
	flow, res, err := client.FrontendApi.CreateNativeLoginFlow(ctx).Execute()
	pkg.SDKExitOnError(err, res)

	// If you want, print the flow here:
	//
	//	pkg.PrintJSONPretty(flow)

	// Submit the form
	result, res, err := client.FrontendApi.UpdateLoginFlow(ctx).Flow(flow.Id).UpdateLoginFlowBody(
		ory.UpdateLoginFlowWithPasswordMethodAsUpdateLoginFlowBody(&ory.UpdateLoginFlowWithPasswordMethod{
			Method:             "password",
			Password:           password,
			PasswordIdentifier: &email,
		}),
	).Execute()
	pkg.SDKExitOnError(err, res)

	return result
}

func main() {
	pkg.PrintJSONPretty(
		performLogin(),
	)
}
