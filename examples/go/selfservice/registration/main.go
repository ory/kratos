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

func initRegistration() *ory.SuccessfulNativeRegistration {
	ctx := context.Background()

	// Initialize the flow
	flow, res, err := client.FrontendApi.CreateNativeRegistrationFlow(ctx).Execute()
	pkg.SDKExitOnError(err, res)

	// If you want, print the flow here:
	//
	//	pkg.PrintJSONPretty(flow)

	email, password := pkg.RandomCredentials()

	// Submit the registration flow
	result, res, err := client.FrontendApi.UpdateRegistrationFlow(ctx).Flow(flow.Id).UpdateRegistrationFlowBody(
		ory.UpdateRegistrationFlowWithPasswordMethodAsUpdateRegistrationFlowBody(&ory.UpdateRegistrationFlowWithPasswordMethod{
			Method:   "password",
			Password: password,
			Traits:   map[string]interface{}{"email": email},
		}),
	).Execute()
	pkg.SDKExitOnError(err, res)
	return result
}

func main() {
	pkg.PrintJSONPretty(
		initRegistration(),
	)
}
