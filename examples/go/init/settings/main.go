package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/google/uuid"

	ory "github.com/ory/kratos-client-go"
)

func sdk(endpoint string) *ory.APIClient {
	conf := ory.NewConfiguration()
	conf.Servers = ory.ServerConfigurations{{URL: endpoint}}
	return ory.NewAPIClient(conf)
}

// If you use Open Source this would be:
//
// 	var c = sdk("https://127.0.0.1:4433")
var c = sdk("https://playground.projects.oryapis.com/api/kratos/public")
var ctx = context.Background()

func initSettings() *ory.SettingsFlow {
	// Create a temporary user
	sessionToken := createUserAndGetSession()
	flow, _, err := c.PublicApi.InitializeSelfServiceSettingsWithoutBrowserExecute(ory.PublicApiApiInitializeSelfServiceSettingsWithoutBrowserRequest{}.
		XSessionToken(sessionToken))
	if err != nil {
		log.Fatalf("An error ocurred: %s\n", err)
	}

	return flow
}

func main() {
	flow := initSettings()
	out, _ := json.MarshalIndent(flow, "", "  ")
	fmt.Println(string(out))
}

// createUserAndGetSession creates a user that we use for in settings flow.
func createUserAndGetSession() string {
	flow, _, err := c.PublicApi.InitializeSelfServiceRegistrationWithoutBrowser(ctx).Execute()
	if err != nil {
		log.Fatalf("An error ocurred: %s\n", err)
	}

	result, _, err := c.PublicApi.SubmitSelfServiceRegistrationFlow(ctx).Flow(flow.Id).SubmitSelfServiceRegistrationFlow(ory.SubmitSelfServiceRegistrationFlow{
		&ory.SubmitSelfServiceRegistrationFlowWithPasswordMethod{
			Method:   "password",
			Password: ory.PtrString(uuid.New().String()),
			Traits:   map[string]interface{}{"email": "dev+" + uuid.New().String() + "@ory.sh"},
		},
	}).Execute()
	if err != nil {
		log.Fatalf("An error ocurred: %s\n", err)
	}

	return *result.SessionToken
}
