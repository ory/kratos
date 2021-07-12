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

func performVerification(email string) *ory.SelfServiceVerificationFlow {
	ctx := context.Background()

	// Initialize the flow
	flow, res, err := client.V0alpha1Api.InitializeSelfServiceVerificationFlowWithoutBrowser(ctx).Execute()
	pkg.SDKExitOnError(err, res)

	// If you want, print the flow here:
	//
	//	pkg.PrintJSONPretty(flow)

	// Submit the form
	afterSubmit, res, err := client.V0alpha1Api.SubmitSelfServiceVerificationFlow(ctx).Flow(flow.Id).
		SubmitSelfServiceVerificationFlowBody(ory.SubmitSelfServiceVerificationFlowWithLinkMethodBodyAsSubmitSelfServiceVerificationFlowBody(&ory.SubmitSelfServiceVerificationFlowWithLinkMethodBody{
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
