package password

import (
	"net/http"

	"github.com/ory/hive/identity"
)

const CredentialsType identity.CredentialsType = "password"

const csrfTokenName = "csrf_token"

type CredentialsConfig struct {
	HashedPassword string `json:"hashed_password"`
}

type csrfGenerator func(r *http.Request) string
