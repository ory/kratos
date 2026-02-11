// Copyright © 2024 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package testhelpers

import (
	"net/http"
	"time"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/session"
)

func NewActiveSession(r *http.Request, reg interface {
	session.ManagementProvider
}, i *identity.Identity, authenticatedAt time.Time, completedLoginFor identity.CredentialsType, completedLoginAAL identity.AuthenticatorAssuranceLevel) (*session.Session, error) {
	s := session.NewInactiveSession()
	s.CompletedLoginFor(completedLoginFor, completedLoginAAL)
	if err := reg.SessionManager().ActivateSession(r, s, i, authenticatedAt); err != nil {
		return nil, err
	}
	return s, nil
}
