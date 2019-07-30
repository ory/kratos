package selfservice

import (
	"context"

	"github.com/ory/hive/identity"
)

type RegistrationRequestManager interface {
	CreateRegistrationRequest(context.Context, *RegistrationRequest) error
	GetRegistrationRequest(ctx context.Context, id string) (*RegistrationRequest, error)
	UpdateRegistrationRequest(context.Context, string, identity.CredentialsType, RequestMethodConfig) error
}

type RegistrationRequestManagementProvider interface {
	RegistrationRequestManager() RegistrationRequestManager
}
