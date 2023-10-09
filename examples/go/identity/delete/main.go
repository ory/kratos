// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

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

	res, err := client.IdentityApi.DeleteIdentity(ctx, identity.Id).Execute()
	pkg.SDKExitOnError(err, res)
}

func main() {
	deleteIdentity()
	fmt.Println("Success")
}
