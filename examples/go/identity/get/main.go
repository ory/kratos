// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"

	ory "github.com/ory/client-go"
	"github.com/ory/kratos/examples/go/pkg"
)

// If you use Open Source this would be:
//
// var client = pkg.NewSDKForSelfHosted("http://127.0.0.1:4433")
var client = pkg.NewSDK("playground")

func getIdentity() *ory.Identity {
	ctx := context.Background()
	created := pkg.CreateIdentity(client)

	identity, res, err := client.IdentityApi.GetIdentity(ctx, created.Id).IncludeCredential([]string{"password"}).Execute()
	pkg.SDKExitOnError(err, res)

	return identity
}

func main() {
	pkg.PrintJSONPretty(getIdentity())
}
