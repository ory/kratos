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

var ctx = context.Background()

func initFlow(email, password string) (string, *ory.SettingsFlow) {
	// Create a temporary user
	_, sessionToken := pkg.CreateIdentityWithSession(client, email, password)

	flow, res, err := client.FrontendApi.CreateNativeSettingsFlow(context.Background()).XSessionToken(sessionToken).Execute()
	pkg.SDKExitOnError(err, res)

	// If you want, print the flow here:
	//
	//	pkg.PrintJSONPretty(flow)

	return sessionToken, flow
}

func changePassword(email, password string) *ory.SettingsFlow {
	sessionToken, flow := initFlow(email, password)

	// Submit the form
	result, res, err := client.FrontendApi.UpdateSettingsFlow(ctx).Flow(flow.Id).XSessionToken(sessionToken).UpdateSettingsFlowBody(
		ory.UpdateSettingsFlowWithPasswordMethodAsUpdateSettingsFlowBody(&ory.UpdateSettingsFlowWithPasswordMethod{
			Method:   "password",
			Password: "not-" + password,
		}),
	).Execute()
	pkg.SDKExitOnError(err, res)

	return result
}

func changeTraits(email, password string) *ory.SettingsFlow {
	sessionToken, flow := initFlow(email, password)

	// Submit the form
	result, res, err := client.FrontendApi.UpdateSettingsFlow(ctx).Flow(flow.Id).XSessionToken(sessionToken).UpdateSettingsFlowBody(
		ory.UpdateSettingsFlowWithProfileMethodAsUpdateSettingsFlowBody(&ory.UpdateSettingsFlowWithProfileMethod{
			Method: "profile",
			Traits: map[string]interface{}{
				"email": "not-" + email,
			},
		}),
	).Execute()
	pkg.SDKExitOnError(err, res)

	return result
}

func main() {
	pkg.PrintJSONPretty(
		changePassword(pkg.RandomCredentials()),
	)
	pkg.PrintJSONPretty(
		changeTraits(pkg.RandomCredentials()),
	)
}
