package persistence

import (
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/selfservice/flow/profile"
	"github.com/ory/kratos/selfservice/flow/registration"
)

type RequestPersister interface {
	registration.RequestPersister
	login.RequestPersister
	profile.RequestPersister
}
