package main

import (
	"context"
	"fmt"

	"github.com/ory/kratos/examples/go/pkg"
)

// If you use Open Source this would be:
//
// var client = pkg.NewSDKForSelfHosted("http://127.0.0.1:4433")
var client = pkg.NewSDK("playground")

func deleteIdentity() {
	ctx := context.Background()

	identity := pkg.CreateIdentity(client)

	res, err := client.V0alpha1Api.AdminDeleteIdentity(ctx, identity.Id).Execute()
	pkg.SDKExitOnError(err, res)
}

func main() {
	deleteIdentity()
	fmt.Println("Success")
}
