package selfservice

import (
	"net/http"

	"github.com/ory/kratos/x"
)

type Strategy interface {
	SetRoutes(*x.RouterPublic)
	PopulateLoginMethod(r *http.Request, sr *LoginRequest) error
	PopulateRegistrationMethod(r *http.Request, sr *RegistrationRequest) error
}

type StrategyProvider interface {
	SelfServiceStrategies() []Strategy
}

// func EnsureEnabled(c configuration.Provider, ct identity.CredentialsType, next httprouter.Handle) func (w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
// 	return func (w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
// 		if !c.LoginStrategy(string(ct)).Enabled {
//
// 			return
// 		}
// 	}
// }
