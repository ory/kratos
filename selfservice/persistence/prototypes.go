package persistence

import (
	"fmt"
	"sync"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/selfservice/flow/registration"
)

var loginRequestPrototypeFactories = map[identity.CredentialsType]func() login.RequestMethodConfig{}
var registrationRequestPrototypeFactories = map[identity.CredentialsType]func() registration.RequestMethodConfig{}
var pl sync.RWMutex

// RegisterRegistrationRequestPrototype a login.RequestMethodConfig prototype for a specific identity.CredentialsType.
func RegisterLoginRequestPrototypeFactory(t identity.CredentialsType, f func() login.RequestMethodConfig) {
	pl.Lock()
	defer pl.Unlock()

	loginRequestPrototypeFactories[t] = f
}

// RegisterRegistrationRequestPrototype a registration.RequestMethodConfig prototype for a specific identity.CredentialsType.
func RegisterRegistrationRequestPrototypeFactory(t identity.CredentialsType, f func() registration.RequestMethodConfig) {
	pl.Lock()
	defer pl.Unlock()

	registrationRequestPrototypeFactories[t] = f
}

// registrationRequestMethodConfigFor returns the registration.RequestMethodConfig for the given identity.CredentialsType.
func registrationRequestMethodConfigFor(t identity.CredentialsType) registration.RequestMethodConfig {
	pl.RLock()
	defer pl.RUnlock()

	f, ok := registrationRequestPrototypeFactories[t]
	if !ok {
		panic(fmt.Sprintf("registration.RequestMethodConfig for identity.CredentialsType (%s) was not registered", t))
	}

	return f()
}

// loginRequestMethodConfigFor returns the login.RequestMethodConfig for the given identity.CredentialsType.
func loginRequestMethodConfigFor(t identity.CredentialsType) login.RequestMethodConfig {
	pl.RLock()
	defer pl.RUnlock()

	f, ok := loginRequestPrototypeFactories[t]
	if !ok {
		panic(fmt.Sprintf("login.RequestMethodConfig for identity.CredentialsType (%s) was not registered", t))
	}

	return f()
}
