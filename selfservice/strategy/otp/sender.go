package otp

import (
	"context"
	"net/http"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/pkg/errors"

	"github.com/ory/kratos/courier"
	templates "github.com/ory/kratos/courier/template/sms"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/flow/recovery"
	"github.com/ory/kratos/selfservice/flow/verification"
	"github.com/ory/kratos/selfservice/token"
	"github.com/ory/kratos/x"
	"github.com/ory/x/errorsx"
	"github.com/ory/x/httpx"
	"github.com/ory/x/sqlcon"
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

		token.VerificationTokenPersistenceProvider
		token.RecoveryTokenPersistenceProvider

		HTTPClient(ctx context.Context, opts ...httpx.ResilientOptions) *retryablehttp.Client
	}

	SenderProvider interface {
		OTPSender() *Sender
	}

	Sender struct {
		r senderDependencies
	}
)

var ErrUnknownPhone = errors.New("verification requested for unknown phone")

func NewSender(r senderDependencies) *Sender {
	return &Sender{r: r}
}

// SendRecoveryOTP sends a one time password to the specified phone number. If the address does not exist in the store, a sms is
// still being sent to prevent account enumeration attacks. In that case, this function returns the ErrUnknownAddress
// error.
func (s *Sender) SendRecoveryOTP(ctx context.Context, r *http.Request, f *recovery.Flow, via identity.VerifiableAddressType, to string) error {
	s.r.Logger().
		WithField("via", via).
		WithSensitiveField("phone_number", to).
		Debug("Preparing verification OTP.")

	address, err := s.r.IdentityPool().FindRecoveryAddressByValue(ctx, identity.RecoveryAddressTypePhone, to)
	if err != nil {
		if err := s.send(ctx, string(via), templates.NewOTPMessage(s.r, &templates.OTPMessageModel{To: to})); err != nil {
			return err
		}
		return errors.Cause(ErrUnknownPhone)
	}

	tkn := token.NewOTPRecovery(address, f, s.r.Config(r.Context()).SelfServiceLinkMethodLifespan())
	if err := s.r.RecoveryTokenPersister().CreateRecoveryToken(ctx, tkn); err != nil {
		return err
	}

	if err := s.sendRecoveryTokenTo(ctx, address, tkn); err != nil {
		return err
	}

	return nil
}

// SendVerificationOTP sends a verification link to the specified address. If the address does not exist in the store, an email is
// still being sent to prevent account enumeration attacks. In that case, this function returns the ErrUnknownAddress
// error.
func (s *Sender) SendVerificationOTP(ctx context.Context, f *verification.Flow, via identity.VerifiableAddressType, to string) error {
	s.r.Logger().
		WithField("via", via).
		WithSensitiveField("phone_number", to).
		Debug("Preparing verification code.")

	address, err := s.r.IdentityPool().FindVerifiableAddressByValue(ctx, via, to)
	if err != nil {
		if errorsx.Cause(err) == sqlcon.ErrNoRows {
			s.r.Audit().
				WithField("via", via).
				WithSensitiveField("phone_number", address).
				Info("Sending out invalid verification email because address is unknown.")
			if err := s.send(ctx, string(via), templates.NewOTPMessage(s.r, &templates.OTPMessageModel{To: to})); err != nil {
				return err
			}
			return errors.Cause(ErrUnknownPhone)
		}
		return err
	}

	tkn := token.NewOTPVerification(address, f, s.r.Config(ctx).SelfServiceLinkMethodLifespan())
	if err := s.r.VerificationTokenPersister().CreateVerificationToken(ctx, tkn); err != nil {
		return err
	}

	if err := s.SendVerificationTokenTo(ctx, address, tkn); err != nil {
		return err
	}
	return nil
}

func (s *Sender) sendRecoveryTokenTo(ctx context.Context, address *identity.RecoveryAddress, tkn *token.RecoveryToken) error {
	s.r.Audit().
		WithField("via", address.Via).
		WithField("identity_id", address.IdentityID).
		WithField("recovery_token_id", tkn.ID).
		WithSensitiveField("phone_number", address.Value).
		WithSensitiveField("recovery_otp_token", tkn.Token).
		Info("Sending out recovery sms with OTP.")

	otpMessage := templates.NewOTPMessage(s.r, &templates.OTPMessageModel{To: address.Value, Code: tkn.Token})

	if err := s.send(ctx, string(address.Via), otpMessage); err != nil {
		return err
	}

	return nil
}

func (s *Sender) SendVerificationTokenTo(ctx context.Context, address *identity.VerifiableAddress, tkn *token.VerificationToken) error {
	s.r.Audit().
		WithField("via", address.Via).
		WithField("identity_id", address.IdentityID).
		WithField("verification_token_id", tkn.ID).
		WithSensitiveField("phone_number", address.Value).
		WithSensitiveField("verification_otp_token", tkn.Token).
		Info("Sending out verification sms with OTP.")

	otpMessage := templates.NewOTPMessage(s.r, &templates.OTPMessageModel{To: address.Value, Code: tkn.Token})

	if err := s.send(ctx, string(address.Via), otpMessage); err != nil {
		return err
	}

	address.Status = identity.VerifiableAddressStatusSent

	if err := s.r.PrivilegedIdentityPool().UpdateVerifiableAddress(ctx, address); err != nil {
		return err
	}

	return nil
}

func (s *Sender) send(ctx context.Context, via string, t courier.SMSTemplate) error {
	switch via {
	case identity.AddressTypePhone:
		_, err := s.r.Courier(ctx).QueueSMS(ctx, t)
		return err
	default:
		return errors.Errorf("received unexpected via type: %s", via)
	}
}
