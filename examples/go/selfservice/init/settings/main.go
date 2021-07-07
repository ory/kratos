package main

import (
	"github.com/ory/kratos/examples/go/pkg"

	ory "github.com/ory/kratos-client-go"
)

// If you use Open Source this would be:
//
//var client = pkg.NewSDKForSelfHosted("http://127.0.0.1:4433")
var client = pkg.NewSDK("playground")

func initSettings() *ory.SettingsFlow {
	// Create a temporary user
	_, sessionToken := pkg.CreateIdentityWithSession(client)
	flow, res, err := client.PublicApi.
		InitializeSelfServiceSettingsWithoutBrowserExecute(ory.
			PublicApiApiInitializeSelfServiceSettingsWithoutBrowserRequest{}.
			XSessionToken(sessionToken))
	pkg.SDKExitOnError(err, res)

	return flow
}

func main() {
	flow := initSettings()
	pkg.PrintJSONPretty(flow)
}
