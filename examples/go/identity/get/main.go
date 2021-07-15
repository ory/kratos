package main

import (
	"context"

	ory "github.com/ory/kratos-client-go"
	"github.com/ory/kratos/examples/go/pkg"
)

// If you use Open Source this would be:
//
// var client = pkg.NewSDKForSelfHosted("http://127.0.0.1:4433")
var client = pkg.NewSDK("playground")

func getIdentity() *ory.Identity {
	ctx := context.Background()
	created := pkg.CreateIdentity(client)

	identity, res, err := client.V0alpha1Api.AdminGetIdentity(ctx, created.Id).Execute()
	pkg.SDKExitOnError(err, res)

	return identity
}

func main() {
	pkg.PrintJSONPretty(getIdentity())
}
