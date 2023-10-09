// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"

	ory "github.com/ory/client-go"
	"github.com/ory/kratos/examples/go/pkg"
	"github.com/ory/kratos/x"
)

// If you use Open Source this would be:
//
// var client = pkg.NewSDKForSelfHosted("http://127.0.0.1:4433")
var client = pkg.NewSDK("playground")

func createIdentity() *ory.Identity {
	ctx := context.Background()

	identity, res, err := client.IdentityApi.CreateIdentity(ctx).
		CreateIdentityBody(ory.CreateIdentityBody{
			SchemaId: "default",
			Traits: map[string]interface{}{
				"email": "dev+" + x.NewUUID().String() + "@ory.sh",
			},
		}).Execute()
	pkg.SDKExitOnError(err, res)

	return identity
}

func main() {
	pkg.PrintJSONPretty(createIdentity())
}
