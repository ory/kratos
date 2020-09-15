package link

import (
	"context"
	"net/url"

	"github.com/pkg/errors"

	"github.com/ory/x/errorsx"
	"github.com/ory/x/sqlcon"
	"github.com/ory/x/urlx"

	"github.com/ory/kratos/courier"
	templates "github.com/ory/kratos/courier/template"
	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/flow/recovery"
	"github.com/ory/kratos/selfservice/flow/verification"
	"github.com/ory/kratos/x"
)

type (
	senderDependencies interface {
		courier.Provider
		identity.PoolProvider
		identity.ManagementProvider
		x.LoggingProvider

		VerificationTokenPersistenceProvider
		RecoveryTokenPersistenceProvider
	}

	SenderProvider interface {
		LinkSender() *Sender
	}

	Sender struct {
		r senderDependencies
		c configuration.Provider
	}
)

var ErrUnknownAddress = errors.New("verification requested for unknown address")

func NewSender(r senderDependencies, c configuration.Provider) *Sender {
	return &Sender{r: r, c: c}
}

// SendRecoveryLink sends a recovery link to the specified address. If the address does not exist in the store, an email is
// still being sent to prevent account enumeration attacks. In that case, this function returns the ErrUnknownAddress
// error.
func (s *Sender) SendRecoveryLink(ctx context.Context, f *recovery.Flow, via identity.VerifiableAddressType, to string) error {
	s.r.Logger().
		WithField("via", via).
		WithSensitiveField("address", to).
		Debug("Preparing verification code.")

	address, err := s.r.IdentityPool().FindRecoveryAddressByValue(ctx, identity.RecoveryAddressTypeEmail, to)
	if err != nil {
		if err := s.send(ctx, string(via), templates.NewRecoveryInvalid(s.c, &templates.RecoveryInvalidModel{To: to})); err != nil {
			return err
		}
		return errors.Cause(ErrUnknownAddress)
	}

	token := NewSelfServiceRecoveryToken(address, f)
	if err := s.r.RecoveryTokenPersister().CreateRecoveryToken(ctx, token); err != nil {
		return err
	}

	if err := s.SendRecoveryTokenTo(ctx, address, token); err != nil {
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
			if err := s.send(ctx, string(via), templates.NewVerificationInvalid(s.c, &templates.VerificationInvalidModel{To: to})); err != nil {
				return err
			}
			return errors.Cause(ErrUnknownAddress)
		}
		return err
	}

	token := NewSelfServiceVerificationToken(address, f)
	if err := s.r.VerificationTokenPersister().CreateVerificationToken(ctx, token); err != nil {
		return err
	}

	if err := s.SendVerificationTokenTo(ctx, address, token); err != nil {
		return err
	}
	return nil
}

func (s *Sender) SendRecoveryTokenTo(ctx context.Context, address *identity.RecoveryAddress, token *RecoveryToken) error {
	s.r.Audit().
		WithField("via", address.Via).
		WithField("identity_id", address.IdentityID).
		WithField("recovery_link_id", token.ID).
		WithSensitiveField("email_address", address.Value).
		WithSensitiveField("recovery_link_token", token.Token).
		Info("Sending out recovery email with recovery link.")
	return s.send(ctx, string(address.Via), templates.NewRecoveryValid(s.c,
		&templates.RecoveryValidModel{To: address.Value, RecoveryURL: urlx.CopyWithQuery(
			urlx.AppendPaths(s.c.SelfPublicURL(), RouteRecovery),
			url.Values{"token": {token.Token}}).String()}))
}

func (s *Sender) SendVerificationTokenTo(ctx context.Context, address *identity.VerifiableAddress, token *VerificationToken) error {
	s.r.Audit().
		WithField("via", address.Via).
		WithField("identity_id", address.IdentityID).
		WithField("verification_link_id", token.ID).
		WithSensitiveField("email_address", address.Value).
		WithSensitiveField("verification_link_token", token.Token).
		Info("Sending out verification email with verification link.")

	return s.send(ctx, string(address.Via), templates.NewVerificationValid(s.c,
		&templates.VerificationValidModel{To: address.Value, VerificationURL: urlx.CopyWithQuery(
			urlx.AppendPaths(s.c.SelfPublicURL(), RouteVerification),
			url.Values{"token": {token.Token}}).String()}))
}

func (s *Sender) send(ctx context.Context, via string, t courier.EmailTemplate) error {
	switch via {
	case identity.AddressTypeEmail:
		_, err := s.r.Courier().QueueEmail(ctx, t)
		return err
	default:
		return errors.Errorf("received unexpected via type: %s", via)
	}
}
