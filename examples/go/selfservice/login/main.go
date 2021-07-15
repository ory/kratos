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

func performLogin() *ory.SuccessfulSelfServiceLoginWithoutBrowser {
	ctx := context.Background()

	// Create a temporary user
	email, password := pkg.RandomCredentials()
	_, _ = pkg.CreateIdentityWithSession(client, email, password)

	// Initialize the flow
	flow, res, err := client.V0alpha1Api.InitializeSelfServiceLoginFlowWithoutBrowser(ctx).Execute()
	pkg.SDKExitOnError(err, res)

	// If you want, print the flow here:
	//
	//	pkg.PrintJSONPretty(flow)

	// Submit the form
	result, res, err := client.V0alpha1Api.SubmitSelfServiceLoginFlow(ctx).Flow(flow.Id).SubmitSelfServiceLoginFlowBody(
		ory.SubmitSelfServiceLoginFlowWithPasswordMethodBodyAsSubmitSelfServiceLoginFlowBody(&ory.SubmitSelfServiceLoginFlowWithPasswordMethodBody{
			Method:             "password",
			Password:           password,
			PasswordIdentifier: email,
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
