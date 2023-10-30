package passkey

import (
	"github.com/ory/kratos/x"
)

func (s *Strategy) RegisterSettingsRoutes(_ *x.RouterPublic) {
}

func (s *Strategy) SettingsStrategyID() string {
	return s.ID().String()
}

const (
	InternalContextKeySessionData = "session_data"
)
