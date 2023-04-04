// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"

	"github.com/ory/kratos/examples/go/pkg"

	ory "github.com/ory/client-go"
)

// If you use Open Source this would be:
//
// var client = pkg.NewSDKForSelfHosted("http://127.0.0.1:4433")
var client = pkg.NewSDK("playground")

func performLogout() {
	// Create a temporary user
	email, password := pkg.RandomCredentials()
	_, sessionToken := pkg.CreateIdentityWithSession(client, email, password)

	// Log out using session token
	res, err := client.FrontendApi.PerformNativeLogout(context.Background()).
		PerformNativeLogoutBody(ory.PerformNativeLogoutBody{SessionToken: sessionToken}).Execute()
	pkg.SDKExitOnError(err, res)
}

func main() {
	performLogout()
	fmt.Println("Logout successful!")
}
