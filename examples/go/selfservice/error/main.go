package main

import (
	"github.com/ory/kratos/examples/go/pkg"

	ory "github.com/ory/kratos-client-go"
)

// If you use Open Source this would be:
//
//var client = pkg.NewSDKForSelfHosted("http://127.0.0.1:4433")
var client = pkg.NewSDK("playground")

func getError() *ory.SelfServiceError {
	e, res, err := client.V0alpha1Api.GetSelfServiceErrorExecute(ory.V0alpha1ApiApiGetSelfServiceErrorRequest{}.Id("stub:500"))
	pkg.SDKExitOnError(err, res)
	return e
}

func main() {
	pkg.PrintJSONPretty(
		getError(),
	)
}
