package main

import (
	"context"
	"encoding/json"
	"fmt"

	ory "github.com/ory/kratos-client-go"
	"github.com/ory/kratos/examples/go/pkg"
)

// If you use Open Source this would be:
//
//var client = pkg.NewSDKForSelfHosted("http://127.0.0.1:4433")
var client = pkg.NewSDK("playground")

var ctx = context.Background()

func performVerification(email string) *ory.VerificationFlow {
	flow, _, err := client.PublicApi.InitializeSelfServiceVerificationWithoutBrowser(ctx).Execute()
	pkg.ExitOnError(err)

	flow, _, err = client.PublicApi.
		SubmitSelfServiceVerificationFlowExecute(ory.PublicApiApiSubmitSelfServiceVerificationFlowRequest{}.
			Flow(flow.Id).
			SubmitSelfServiceVerificationFlowBody(
				ory.SubmitSelfServiceVerificationFlowBody{
					Email: email, Method: "link"}))
	pkg.ExitOnError(err)
	return flow
}

func main() {
	flow := performVerification("someone@foobar.com")
	out, _ := json.MarshalIndent(flow, "", "  ")
	fmt.Println(string(out))
}
