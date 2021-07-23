package main

import (
	"context"

	"github.com/ory/kratos/examples/go/pkg"

	ory "github.com/ory/kratos-client-go"
)

// If you use Open Source this would be:
//
//var client = pkg.NewSDKForSelfHosted("http://127.0.0.1:4433")
var client = pkg.NewSDK("playground")

var ctx = context.Background()

func initFlow(email, password string) (string, *ory.SelfServiceSettingsFlow) {
	// Create a temporary user
	_, sessionToken := pkg.CreateIdentityWithSession(client, email, password)

	flow, res, err := client.V0alpha1Api.InitializeSelfServiceSettingsFlowWithoutBrowserExecute(ory.V0alpha1ApiApiInitializeSelfServiceSettingsFlowWithoutBrowserRequest{}.
		XSessionToken(sessionToken))
	pkg.SDKExitOnError(err, res)

	// If you want, print the flow here:
	//
	//	pkg.PrintJSONPretty(flow)

	return sessionToken, flow
}

func changePassword(email, password string) *ory.SuccessfulSelfServiceSettingsWithoutBrowser {
	sessionToken, flow := initFlow(email, password)

	// Submit the form
	result, res, err := client.V0alpha1Api.SubmitSelfServiceSettingsFlow(ctx).Flow(flow.Id).XSessionToken(sessionToken).SubmitSelfServiceSettingsFlowBody(
		ory.SubmitSelfServiceSettingsFlowWithPasswordMethodBodyAsSubmitSelfServiceSettingsFlowBody(&ory.SubmitSelfServiceSettingsFlowWithPasswordMethodBody{
			Method:   "password",
			Password: "not-" + password,
		}),
	).Execute()
	pkg.SDKExitOnError(err, res)

	return result
}

func changeTraits(email, password string) *ory.SuccessfulSelfServiceSettingsWithoutBrowser {
	sessionToken, flow := initFlow(email, password)

	// Submit the form
	result, res, err := client.V0alpha1Api.SubmitSelfServiceSettingsFlow(ctx).Flow(flow.Id).XSessionToken(sessionToken).SubmitSelfServiceSettingsFlowBody(
		ory.SubmitSelfServiceSettingsFlowWithProfileMethodBodyAsSubmitSelfServiceSettingsFlowBody(&ory.SubmitSelfServiceSettingsFlowWithProfileMethodBody{
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
