package selfservice

import (
	"context"

	"github.com/ory/kratos/identity"
)

type RequestManager interface {
	RegistrationRequestManager
	LoginRequestManager
}

type ProfileRequestManager interface {
	CreateProfileRequest(context.Context, *LoginRequest) error
	GetProfileRequest(ctx context.Context, id string) (*LoginRequest, error)
	UpdateProfileRequest(context.Context, string, identity.CredentialsType, RequestMethodConfig) error
}

type ProfileRequestManagementProvider interface {
	ProfileRequestManager() ProfileRequestManager
}

type LoginRequestManager interface {
	CreateLoginRequest(context.Context, *LoginRequest) error
	GetLoginRequest(ctx context.Context, id string) (*LoginRequest, error)
	UpdateLoginRequest(context.Context, string, identity.CredentialsType, RequestMethodConfig) error
}

type LoginRequestManagementProvider interface {
	LoginRequestManager() LoginRequestManager
}

type RegistrationRequestManager interface {
	CreateRegistrationRequest(context.Context, *RegistrationRequest) error
	GetRegistrationRequest(ctx context.Context, id string) (*RegistrationRequest, error)
	UpdateRegistrationRequest(context.Context, string, identity.CredentialsType, RequestMethodConfig) error
}

type RegistrationRequestManagementProvider interface {
	RegistrationRequestManager() RegistrationRequestManager
}
