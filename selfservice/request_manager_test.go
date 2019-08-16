package selfservice_test

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/x/sqlcon/dockertest"

	"github.com/ory/hive/identity"
	"github.com/ory/hive/internal"
	. "github.com/ory/hive/selfservice"
	"github.com/ory/hive/selfservice/oidc"
	"github.com/ory/hive/selfservice/password"
)

func TestRequestManagerMemory(t *testing.T) {
	managers := map[string]RequestManager{
		"memory": NewRequestManagerMemory(),
	}

	if !testing.Short() {
		var l sync.Mutex
		dockertest.Parallel([]func(){
			func() {
				db, err := dockertest.ConnectToTestPostgreSQL()
				require.NoError(t, err)

				_, reg := internal.NewRegistrySQL(t, db)
				manager := reg.LoginRequestManager().(*RequestManagerSQL)

				l.Lock()
				managers["postgres"] = manager
				l.Unlock()
			},
		})
	}

	nbr := func() *Request {
		return &Request{
			ID:             uuid.New().String(),
			IssuedAt:       time.Now().UTC().Round(time.Second),
			ExpiresAt:      time.Now().Add(time.Hour).UTC().Round(time.Second),
			RequestURL:     "https://www.ory.sh/request",
			RequestHeaders: http.Header{"Content-Type": {"application/json"}},
			Active:         identity.CredentialsTypePassword,
			Methods: map[identity.CredentialsType]*DefaultRequestMethod{
				identity.CredentialsTypePassword: {
					Method: identity.CredentialsTypePassword,
					Config: password.NewRequestMethodConfig(),
				},
				identity.CredentialsTypeOIDC: {
					Method: identity.CredentialsTypeOIDC,
					Config: oidc.NewRequestMethodConfig(),
				},
			},
		}
	}

	assertUpdated := func(t *testing.T, expected, actual Request) {
		assert.EqualValues(t, identity.CredentialsTypePassword, actual.Active)
		assert.EqualValues(t, "bar", actual.Methods[identity.CredentialsTypeOIDC].Config.(*oidc.RequestMethodConfig).Action)
		assert.EqualValues(t, "foo", actual.Methods[identity.CredentialsTypePassword].Config.(*password.RequestMethodConfig).Action)
	}

	for name, m := range managers {
		t.Run(fmt.Sprintf("manager=%s", name), func(t *testing.T) {

			t.Run("suite=login", func(t *testing.T) {
				r := LoginRequest{Request: nbr()}
				require.NoError(t, m.CreateLoginRequest(context.Background(), &r))

				g, err := m.GetLoginRequest(context.Background(), r.ID)
				require.NoError(t, err)
				assert.EqualValues(t, r, *g)

				require.NoError(t, m.UpdateLoginRequest(context.Background(), r.ID, identity.CredentialsTypeOIDC, &oidc.RequestMethodConfig{Action: "bar"}))
				require.NoError(t, m.UpdateLoginRequest(context.Background(), r.ID, identity.CredentialsTypePassword, &password.RequestMethodConfig{Action: "foo"}))

				g, err = m.GetLoginRequest(context.Background(), r.ID)
				require.NoError(t, err)
				assertUpdated(t, *r.Request, *g.Request)
			})

			t.Run("suite=registration", func(t *testing.T) {
				r := RegistrationRequest{Request: nbr()}

				require.NoError(t, m.CreateRegistrationRequest(context.Background(), &r))
				g, err := m.GetRegistrationRequest(context.Background(), r.ID)
				require.NoError(t, err)
				assert.EqualValues(t, r, *g)

				require.NoError(t, m.UpdateRegistrationRequest(context.Background(), r.ID, identity.CredentialsTypeOIDC, &oidc.RequestMethodConfig{Action: "bar"}))
				require.NoError(t, m.UpdateRegistrationRequest(context.Background(), r.ID, identity.CredentialsTypePassword, &password.RequestMethodConfig{Action: "foo"}))

				g, err = m.GetRegistrationRequest(context.Background(), r.ID)
				require.NoError(t, err)
				assertUpdated(t, *r.Request, *g.Request)
			})
		})
	}
}
