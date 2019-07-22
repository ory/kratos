package selfservice

import (
	"context"

	"github.com/ory/hive-cloud/hive/identity"
)

type LoginRequestManager interface {
	CreateLoginRequest(context.Context, *LoginRequest) error
	GetLoginRequest(ctx context.Context, id string) (*LoginRequest, error)
	UpdateLoginRequest(context.Context, string, identity.CredentialsType, interface{}) error
}

type LoginRequestManagementProvider interface {
	LoginRequestManager() LoginRequestManager
}
