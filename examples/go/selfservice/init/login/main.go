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

func initLogin() *ory.LoginFlow {
	ctx := context.Background()

	flow, res, err := client.PublicApi.InitializeSelfServiceLoginWithoutBrowser(ctx).Execute()
	pkg.SDKExitOnError(err, res)

	return flow
}

func main() {
	flow := initLogin()
	pkg.PrintJSONPretty(flow)
}
