// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package link

import (
	"context"
	"net/http"
	"net/url"

	"github.com/hashicorp/go-retryablehttp"

	"github.com/ory/kratos/courier/template/email"

	"github.com/ory/x/httpx"

	"github.com/pkg/errors"

	"github.com/ory/x/errorsx"
	"github.com/ory/x/sqlcon"
	"github.com/ory/x/urlx"

	"github.com/ory/kratos/courier"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/flow/recovery"
	"github.com/ory/kratos/selfservice/flow/verification"
	"github.com/ory/kratos/x"
)

type (
	senderDependencies interface {
		courier.Provider
		courier.ConfigProvider

		identity.PoolProvider
		identity.ManagementProvider
		identity.PrivilegedPoolProvider
		x.LoggingProvider
		config.Provider

		VerificationTokenPersistenceProvider
		RecoveryTokenPersistenceProvider

		HTTPClient(ctx context.Context, opts ...httpx.ResilientOptions) *retryablehttp.Client
	}
	SenderProvider interface {
		LinkSender() *Sender
	}

	Sender struct {
		r senderDependencies
	}
)

var ErrUnknownAddress = errors.New("verification requested for unknown address")

func NewSender(r senderDependencies) *Sender {
	return &Sender{r: r}
}

// SendRecoveryLink sends a recovery link to the specified address. If the address does not exist in the store, an email is
// still being sent to prevent account enumeration attacks. In that case, this function returns the ErrUnknownAddress
// error.
func (s *Sender) SendRecoveryLink(ctx context.Context, r *http.Request, f *recovery.Flow, via identity.VerifiableAddressType, to string) error {
	s.r.Logger().
		WithField("via", via).
		WithSensitiveField("address", to).
		Debug("Preparing verification code.")

	address, err := s.r.IdentityPool().FindRecoveryAddressByValue(ctx, identity.RecoveryAddressTypeEmail, to)
	if err != nil {
		if err := s.send(ctx, string(via), email.NewRecoveryInvalid(s.r, &email.RecoveryInvalidModel{To: to})); err != nil {
			return err
		}
		return errors.Cause(ErrUnknownAddress)
	}

	// Get the identity associated with the recovery address
	i, err := s.r.IdentityPool().GetIdentity(ctx, address.IdentityID)
	if err != nil {
		return err
	}

	token := NewSelfServiceRecoveryToken(address, f, s.r.Config().SelfServiceLinkMethodLifespan(r.Context()))
	if err := s.r.RecoveryTokenPersister().CreateRecoveryToken(ctx, token); err != nil {
		return err
	}

	if err := s.SendRecoveryTokenTo(ctx, f, i, address, token); err != nil {
		return err
	}

	return nil
}

// SendVerificationLink sends a verification link to the specified address. If the address does not exist in the store, an email is
// still being sent to prevent account enumeration attacks. In that case, this function returns the ErrUnknownAddress
// error.
func (s *Sender) SendVerificationLink(ctx context.Context, f *verification.Flow, via identity.VerifiableAddressType, to string) error {
	s.r.Logger().
		WithField("via", via).
		WithSensitiveField("address", to).
		Debug("Preparing verification code.")

	address, err := s.r.IdentityPool().FindVerifiableAddressByValue(ctx, via, to)
	if err != nil {
		if errorsx.Cause(err) == sqlcon.ErrNoRows {
			s.r.Audit().
				WithField("via", via).
				WithSensitiveField("email_address", address).
				Info("Sending out invalid verification email because address is unknown.")
			if err := s.send(ctx, string(via), email.NewVerificationInvalid(s.r, &email.VerificationInvalidModel{To: to})); err != nil {
				return err
			}
			return errors.Cause(ErrUnknownAddress)
		}
		return err
	}

	// Get the identity associated with the recovery address
	i, err := s.r.IdentityPool().GetIdentity(ctx, address.IdentityID)
	if err != nil {
		return err
	}

	token := NewSelfServiceVerificationToken(address, f, s.r.Config().SelfServiceLinkMethodLifespan(ctx))
	if err := s.r.VerificationTokenPersister().CreateVerificationToken(ctx, token); err != nil {
		return err
	}

	if err := s.SendVerificationTokenTo(ctx, f, i, address, token); err != nil {
		return err
	}
	return nil
}

func (s *Sender) SendRecoveryTokenTo(ctx context.Context, f *recovery.Flow, i *identity.Identity, address *identity.RecoveryAddress, token *RecoveryToken) error {
	s.r.Audit().
		WithField("via", address.Via).
		WithField("identity_id", address.IdentityID).
		WithField("recovery_link_id", token.ID).
		WithSensitiveField("email_address", address.Value).
		WithSensitiveField("recovery_link_token", token.Token).
		Info("Sending out recovery email with recovery link.")

	model, err := x.StructToMap(i)
	if err != nil {
		return err
	}

	return s.send(ctx, string(address.Via), email.NewRecoveryValid(s.r,
		&email.RecoveryValidModel{To: address.Value, RecoveryURL: urlx.CopyWithQuery(
			urlx.AppendPaths(s.r.Config().SelfServiceLinkMethodBaseURL(ctx), recovery.RouteSubmitFlow),
			url.Values{
				"token": {token.Token},
				"flow":  {f.ID.String()},
			}).String(), Identity: model}))
}

func (s *Sender) SendVerificationTokenTo(ctx context.Context, f *verification.Flow, i *identity.Identity, address *identity.VerifiableAddress, token *VerificationToken) error {
	s.r.Audit().
		WithField("via", address.Via).
		WithField("identity_id", address.IdentityID).
		WithField("verification_link_id", token.ID).
		WithSensitiveField("email_address", address.Value).
		WithSensitiveField("verification_link_token", token.Token).
		Info("Sending out verification email with verification link.")

	model, err := x.StructToMap(i)
	if err != nil {
		return err
	}

	if err := s.send(ctx, string(address.Via), email.NewVerificationValid(s.r,
		&email.VerificationValidModel{To: address.Value, VerificationURL: urlx.CopyWithQuery(
			urlx.AppendPaths(s.r.Config().SelfServiceLinkMethodBaseURL(ctx), verification.RouteSubmitFlow),
			url.Values{
				"flow":  {f.ID.String()},
				"token": {token.Token},
			}).String(), Identity: model})); err != nil {
		return err
	}
	address.Status = identity.VerifiableAddressStatusSent
	if err := s.r.PrivilegedIdentityPool().UpdateVerifiableAddress(ctx, address); err != nil {
		return err
	}
	return nil
}

func (s *Sender) send(ctx context.Context, via string, t courier.EmailTemplate) error {
	switch via {
	case identity.AddressTypeEmail:
		c, err := s.r.Courier(ctx)
		if err != nil {
			return err
		}
		_, err = c.QueueEmail(ctx, t)
		return err
	default:
		return errors.Errorf("received unexpected via type: %s", via)
	}
}
