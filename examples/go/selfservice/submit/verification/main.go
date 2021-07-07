package main

import (
	"context"

	ory "github.com/ory/kratos-client-go"
	"github.com/ory/kratos/examples/go/pkg"
)

// If you use Open Source this would be:
//
//var client = pkg.NewSDKForSelfHosted("http://127.0.0.1:4433")
var client = pkg.NewSDK("playground")

var ctx = context.Background()

func performVerification(email string) *ory.VerificationFlow {
	flow, res, err := client.PublicApi.InitializeSelfServiceVerificationWithoutBrowser(ctx).Execute()
	pkg.SDKExitOnError(err, res)

	flow, res, err = client.PublicApi.
		SubmitSelfServiceVerificationFlowExecute(ory.PublicApiApiSubmitSelfServiceVerificationFlowRequest{}.
			Flow(flow.Id).
			SubmitSelfServiceVerificationFlowBody(
				ory.SubmitSelfServiceVerificationFlowBody{
					Email: email, Method: "link"}))
	pkg.SDKExitOnError(err, res)
	return flow
}

func main() {
	result := performVerification("someone@foobar.com")
	pkg.PrintJSONPretty(result)
}
