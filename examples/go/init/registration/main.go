package main

import (
	"context"
	"encoding/json"
	"fmt"
	ory "github.com/ory/kratos-client-go"
	"log"
)

func sdk(endpoint string) *ory.APIClient {
	conf := ory.NewConfiguration()
	conf.Servers = ory.ServerConfigurations{{URL: endpoint}}
	return ory.NewAPIClient(conf)
}

func initRegistration() *ory.RegistrationFlow {
	ctx := context.Background()

	// If you use Open Source this would be:
	//
	// c := sdk("https://127.0.0.1:4433")
	c := sdk("https://playground.projects.oryapis.com/api/kratos/public")

	flow, _, err := c.PublicApi.InitializeSelfServiceRegistrationWithoutBrowser(ctx).Execute()
	if err != nil {
		log.Fatalf("An error ocurred: %s\n", err)
	}

	return flow
}

func main() {
	flow := initRegistration()
	out, _ := json.MarshalIndent(flow, "", "  ")
	fmt.Println(string(out))
}
