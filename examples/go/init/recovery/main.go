package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	ory "github.com/ory/kratos-client-go"
)

func sdk(endpoint string) *ory.APIClient {
	conf := ory.NewConfiguration()
	conf.Servers = ory.ServerConfigurations{{URL: endpoint}}
	return ory.NewAPIClient(conf)
}

func initRecovery() *ory.RecoveryFlow {
	ctx := context.Background()

	// If you use Open Source this would be:
	//
	// c := sdk("https://127.0.0.1:4433")
	c := sdk("https://playground.projects.oryapis.com/api/kratos/public")

	flow, _, err := c.PublicApi.InitializeSelfServiceRecoveryWithoutBrowser(ctx).Execute()
	if err != nil {
		log.Fatalf("An error ocurred: %s\n", err)
	}

	return flow
}

func main() {
	flow := initRecovery()
	out, _ := json.MarshalIndent(flow, "", "  ")
	fmt.Println(string(out))
}
