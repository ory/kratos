package oidc

import "github.com/ory/hive-cloud/hive/identity"

const (
	CredentialsType identity.CredentialsType = "oidc"

	sessionName      = "oidc_session"
	sessionRequestID = "request_id"
	sessionKeyState  = "state"
)
