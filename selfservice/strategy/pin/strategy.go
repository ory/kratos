package pin

import (
	"github.com/go-playground/validator/v10"
	"github.com/ory/kratos/continuity"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/hash"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/ui/node"
	"github.com/ory/kratos/x"
	"github.com/ory/x/decoderx"
)

type loginStrategyDependencies interface {
	x.LoggingProvider
	x.CSRFTokenGeneratorProvider

	config.Provider

	continuity.ManagementProvider

	hash.HashProvider

	identity.PrivilegedPoolProvider

	login.FlowPersistenceProvider

	session.ManagementProvider
}

type Strategy struct {
	d  loginStrategyDependencies
	v  *validator.Validate
	hd *decoderx.HTTP
}

func NewStrategy(d loginStrategyDependencies) *Strategy {
	return &Strategy{
		d:  d,
		v:  validator.New(),
		hd: decoderx.NewHTTP(),
	}
}

func (s *Strategy) ID() identity.CredentialsType {
	return identity.CredentialsTypePin
}

func (s *Strategy) NodeGroup() node.UiNodeGroup {
	return node.PinGroup
}
