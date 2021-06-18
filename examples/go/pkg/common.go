package pkg

import (
	"encoding/json"
	"fmt"
	"os"

	ory "github.com/ory/kratos-client-go"
)

func NewSDK(project string) *ory.APIClient {
	conf := ory.NewConfiguration()
	conf.Servers = ory.ServerConfigurations{{URL: fmt.Sprintf("https://%s.projects.oryapis.com/api/kratos/public", project)}}
	return ory.NewAPIClient(conf)
}

func NewSDKForSelfHosted(endpoint string) *ory.APIClient {
	conf := ory.NewConfiguration()
	conf.Servers = ory.ServerConfigurations{{URL: endpoint}}
	return ory.NewAPIClient(conf)
}

func ExitOnError(err error) {
	if err == nil {
		return
	}
	out, _ := json.MarshalIndent(err, "", "  ")
	fmt.Printf("%s\n\nAn error occurred: %+v\n", out, err)
	os.Exit(1)
}
