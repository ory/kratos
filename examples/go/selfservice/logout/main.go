package main

import (
	"fmt"

	"github.com/ory/kratos/examples/go/pkg"

	ory "github.com/ory/kratos-client-go"
)

// If you use Open Source this would be:
//
//var client = pkg.NewSDKForSelfHosted("http://127.0.0.1:4433")
var client = pkg.NewSDK("playground")

func performLogout() {
	// Create a temporary user
	email, password := pkg.RandomCredentials()
	_, sessionToken := pkg.CreateIdentityWithSession(client, email, password)

	// Log out using session token
	res, err := client.V0alpha1Api.SubmitSelfServiceLogoutFlowWithoutBrowserExecute(ory.V0alpha1ApiApiSubmitSelfServiceLogoutFlowWithoutBrowserRequest{}.
		SubmitSelfServiceLogoutFlowWithoutBrowserBody(ory.SubmitSelfServiceLogoutFlowWithoutBrowserBody{SessionToken: sessionToken}))
	pkg.SDKExitOnError(err, res)
}

func main() {
	performLogout()
	fmt.Println("Logout successful!")
}
