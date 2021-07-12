package pkg

import (
	"context"
	"log"
	"strings"

	"github.com/google/uuid"

	ory "github.com/ory/kratos-client-go"
)

func RandomCredentials() (email, password string) {
	email = "dev+" + uuid.New().String() + "@ory.sh"
	password = strings.ReplaceAll(uuid.New().String(), "-", "")
	return
}

// CreateIdentityWithSession creates an identity and an Ory Session Token for it.
func CreateIdentityWithSession(c *ory.APIClient, email, password string) (*ory.Session, string) {
	ctx := context.Background()

	if email == "" {
		email, _ = RandomCredentials()
	}

	if password == "" {
		_, password = RandomCredentials()
	}

	// Initialize a registration flow
	flow, _, err := c.V0alpha1Api.InitializeSelfServiceRegistrationFlowWithoutBrowser(ctx).Execute()
	ExitOnError(err)

	// Submit the registration flow
	result, res, err := c.V0alpha1Api.SubmitSelfServiceRegistrationFlow(ctx).Flow(flow.Id).SubmitSelfServiceRegistrationFlowBody(
		ory.SubmitSelfServiceRegistrationFlowWithPasswordMethodBodyAsSubmitSelfServiceRegistrationFlowBody(&ory.SubmitSelfServiceRegistrationFlowWithPasswordMethodBody{
			Method:   "password",
			Password: password,
			Traits:   map[string]interface{}{"email": email},
		}),
	).Execute()
	SDKExitOnError(err, res)

	if result.Session == nil {
		log.Fatalf("The server is expected to create sessions for new registrations.")
	}

	return result.Session, *result.SessionToken
}

func CreateIdentity(c *ory.APIClient) *ory.Identity {
	ctx := context.Background()

	email, _ := RandomCredentials()
	identity, _, err := c.V0alpha1Api.AdminCreateIdentity(ctx).AdminCreateIdentityBody(ory.AdminCreateIdentityBody{
		SchemaId: "default",
		Traits: map[string]interface{}{
			"email": email,
		}}).Execute()
	ExitOnError(err)
	return identity
}
