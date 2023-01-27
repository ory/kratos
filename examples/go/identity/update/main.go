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

func updateIdentity() *ory.Identity {
	ctx := context.Background()
	toUpdate := pkg.CreateIdentity(client)

	identity, res, err := client.IdentityApi.UpdateIdentity(ctx, toUpdate.Id).UpdateIdentityBody(ory.UpdateIdentityBody{
		Traits: map[string]interface{}{
			"email": "dev+not-" + x.NewUUID().String() + "@ory.sh",
		},
	}).Execute()
	pkg.SDKExitOnError(err, res)

	return identity
}

func main() {
	pkg.PrintJSONPretty(updateIdentity())
}
