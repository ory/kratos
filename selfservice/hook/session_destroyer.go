// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package hook

import (
	"context"
	"net/http"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/selfservice/flow/recovery"
	"github.com/ory/kratos/selfservice/flow/settings"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/ui/node"
	"github.com/ory/x/otelx"
)

var _ login.PostHookExecutor = new(SessionDestroyer)
var _ recovery.PostHookExecutor = new(SessionDestroyer)
var _ settings.PostHookPostPersistExecutor = new(SessionDestroyer)

type (
	sessionDestroyerDependencies interface {
		session.ManagementProvider
		session.PersistenceProvider
	}
	SessionDestroyer struct {
		r sessionDestroyerDependencies
	}
)

func NewSessionDestroyer(r sessionDestroyerDependencies) *SessionDestroyer {
	return &SessionDestroyer{r: r}
}

func (e *SessionDestroyer) ExecuteLoginPostHook(_ http.ResponseWriter, r *http.Request, _ node.UiNodeGroup, _ *login.Flow, s *session.Session) error {
	return otelx.WithSpan(r.Context(), "selfservice.hook.SessionDestroyer.ExecuteLoginPostHook", func(ctx context.Context) error {
		if _, err := e.r.SessionPersister().RevokeSessionsIdentityExcept(ctx, s.Identity.ID, s.ID); err != nil {
			return err
		}
		return nil
	})
}

func (e *SessionDestroyer) ExecutePostRecoveryHook(_ http.ResponseWriter, r *http.Request, _ *recovery.Flow, s *session.Session) error {
	return otelx.WithSpan(r.Context(), "selfservice.hook.SessionDestroyer.ExecutePostRecoveryHook", func(ctx context.Context) error {
		if _, err := e.r.SessionPersister().RevokeSessionsIdentityExcept(ctx, s.Identity.ID, s.ID); err != nil {
			return err
		}
		return nil
	})
}

func (e *SessionDestroyer) ExecuteSettingsPostPersistHook(_ http.ResponseWriter, r *http.Request, _ *settings.Flow, i *identity.Identity, s *session.Session) error {
	return otelx.WithSpan(r.Context(), "selfservice.hook.SessionDestroyer.ExecuteSettingsPostPersistHook", func(ctx context.Context) error {
		if _, err := e.r.SessionPersister().RevokeSessionsIdentityExcept(ctx, i.ID, s.ID); err != nil {
			return err
		}
		return nil
	})
}
