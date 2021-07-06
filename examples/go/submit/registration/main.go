package main

import (
	"context"

	"github.com/google/uuid"

	ory "github.com/ory/kratos-client-go"
	"github.com/ory/kratos/examples/go/pkg"
)

// If you use Open Source this would be:
//
//	var client = pkg.NewSDKForSelfHosted("http://127.0.0.1:4433")
var client = pkg.NewSDK("playground")

var ctx = context.Background()

func performRegistration() *ory.RegistrationViaApiResponse {
	flow, _, err := client.PublicApi.InitializeSelfServiceRegistrationWithoutBrowser(ctx).Execute()
	pkg.ExitOnError(err)

	result, res, err := client.PublicApi.SubmitSelfServiceRegistrationFlow(ctx).Flow(flow.Id).SubmitSelfServiceRegistrationFlow(ory.SubmitSelfServiceRegistrationFlow{
		&ory.SubmitSelfServiceRegistrationFlowWithPasswordMethod{
			Method:   "password",
			Password: ory.PtrString(uuid.New().String() + uuid.New().String()),
			Traits: map[string]interface{}{
				"email": "dev+" + uuid.New().String() + "@ory.sh",
			},
		},
	}).Execute()
	pkg.SDKExitOnError(err, res)
	return result
}

func main() {
	result := performRegistration()
	pkg.PrintJSONPretty(result)
}
