package poc

import (
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/selfservice/flow/profile"
	"github.com/ory/kratos/selfservice/flow/registration"
)

type Persister interface {
	// identity.Pool
	registration.RequestPersister
	login.RequestPersister
	profile.RequestPersister
}