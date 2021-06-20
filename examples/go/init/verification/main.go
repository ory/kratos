package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/ory/kratos/examples/go/pkg"

	ory "github.com/ory/kratos-client-go"
)

// If you use Open Source this would be:
//
//var client = pkg.NewSDKForSelfHosted("http://127.0.0.1:4433")
var client = pkg.NewSDK("playground")

func initVerification() *ory.VerificationFlow {
	ctx := context.Background()

	flow, _, err := client.PublicApi.InitializeSelfServiceVerificationWithoutBrowser(ctx).Execute()
	if err != nil {
		log.Fatalf("An error ocurred: %s\n", err)
	}

	return flow
}

func main() {
	flow := initVerification()
	out, _ := json.MarshalIndent(flow, "", "  ")
	fmt.Println(string(out))
}
