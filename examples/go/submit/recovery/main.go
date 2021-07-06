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

func performRecovery(email string) *ory.RecoveryFlow {
	flow, res, err := client.PublicApi.InitializeSelfServiceRecoveryWithoutBrowser(ctx).Execute()
	pkg.SDKExitOnError(err, res)

	flow, res, err = client.PublicApi.
		SubmitSelfServiceRecoveryFlowExecute(ory.PublicApiApiSubmitSelfServiceRecoveryFlowRequest{}.
			Flow(flow.Id).
			SubmitSelfServiceRecoveryFlowBody(
				ory.SubmitSelfServiceRecoveryFlowBody{
					Email: email, Method: "link"}))
	pkg.SDKExitOnError(err, res)
	return flow
}

func main() {
	flow := performRecovery("someone@foobar.com")
	out, _ := json.MarshalIndent(flow, "", "  ")
	fmt.Println(string(out))
}
