package pkg

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/google/uuid"

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

// CreateIdentityWithSession creates an identity and an Ory Session Token for it.
func CreateIdentityWithSession(c *ory.APIClient) (*ory.Session, string) {
	ctx := context.Background()

	// Initialize a registration flow
	flow, _, err := c.PublicApi.InitializeSelfServiceRegistrationWithoutBrowser(ctx).Execute()
	if err != nil {
		log.Fatalf("An error ocurred during registration initialization: %s\n", err)
	}

	// Submit the registration flow
	result, _, err := c.PublicApi.SubmitSelfServiceRegistrationFlow(ctx).Flow(flow.Id).SubmitSelfServiceRegistrationFlow(ory.SubmitSelfServiceRegistrationFlow{
		&ory.SubmitSelfServiceRegistrationFlowWithPasswordMethod{
			Method:   "password",
			Password: ory.PtrString(uuid.New().String() + uuid.New().String()),
			Traits: map[string]interface{}{
				"email": "dev+" + uuid.New().String() + "@ory.sh",
			},
		},
	}).Execute()
	if err != nil {
		log.Fatalf("An error ocurred during registration: %s\n", err)
	}

	if result.Session == nil {
		log.Fatalf("The server is expected to create sessions for new registrations.")
	}

	return result.Session, *result.SessionToken
}
