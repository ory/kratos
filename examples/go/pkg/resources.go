package pkg

import (
	"context"
	"log"

	"github.com/google/uuid"

	ory "github.com/ory/kratos-client-go"
)

// CreateIdentityWithSession creates an identity and an Ory Session Token for it.
func CreateIdentityWithSession(c *ory.APIClient) (*ory.Session, string) {
	ctx := context.Background()

	// Initialize a registration flow
	flow, _, err := c.PublicApi.InitializeSelfServiceRegistrationWithoutBrowser(ctx).Execute()
	ExitOnError(err)

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
	ExitOnError(err)

	if result.Session == nil {
		log.Fatalf("The server is expected to create sessions for new registrations.")
	}

	return result.Session, *result.SessionToken
}

func CreateIdentity(c *ory.APIClient) *ory.Identity {
	ctx := context.Background()

	identity, _, err := c.V0alpha0Api.AdminCreateIdentity(ctx).AdminCreateIdentityBody(ory.AdminCreateIdentityBody{
		SchemaId: "default",
		Traits: map[string]interface{}{
			"email": "dev+" + uuid.New().String() + "@ory.sh",
		}}).Execute()
	ExitOnError(err)
	return identity
}
