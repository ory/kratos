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

func performRecovery(email string) *ory.SelfServiceRecoveryFlow {
	ctx := context.Background()

	// Initialize the flow
	flow, res, err := client.V0alpha1Api.InitializeSelfServiceRecoveryFlowWithoutBrowser(ctx).Execute()
	pkg.SDKExitOnError(err, res)

	// If you want, print the flow here:
	//
	//	pkg.PrintJSONPretty(flow)

	// Submit the form
	afterSubmit, res, err := client.V0alpha1Api.SubmitSelfServiceRecoveryFlow(ctx).Flow(flow.Id).
		SubmitSelfServiceRecoveryFlowBody(ory.SubmitSelfServiceRecoveryFlowWithLinkMethodBodyAsSubmitSelfServiceRecoveryFlowBody(&ory.SubmitSelfServiceRecoveryFlowWithLinkMethodBody{
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
