package main

import (
	"fmt"
	"log"

	"github.com/ory/kratos/examples/go/pkg"

	ory "github.com/ory/kratos-client-go"
)

// If you use Open Source this would be:
//
//var client = pkg.NewSDKForSelfHosted("http://127.0.0.1:4433")
var client = pkg.NewSDK("playground")

func performLogout() {
	// Create a temporary user
	_, sessionToken := pkg.CreateIdentityWithSession(client)
	_, err := client.PublicApi.SubmitSelfServiceLogoutFlowWithoutBrowserExecute(ory.PublicApiApiSubmitSelfServiceLogoutFlowWithoutBrowserRequest{}.
		SubmitSelfServiceLogoutFlowWithoutBrowserBody(ory.SubmitSelfServiceLogoutFlowWithoutBrowserBody{SessionToken: sessionToken}))
	if err != nil {
		log.Fatalf("An error occurred: %s\n", err)
	}
}

func main() {
	performLogout()
	fmt.Println("Logout successful!")
}
