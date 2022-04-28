package otp

import (
	"context"
	"net/http"
	"reflect"
	"strings"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/pkg/errors"

	"github.com/ory/kratos/courier"
	"github.com/ory/kratos/courier/template/email"
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

var ErrUnknownIdentifier = errors.New("operation is requested for unknown identifier")

func NewSender(r senderDependencies) *Sender {
	return &Sender{r: r}
}

// SendRecoveryOTP sends a recovery OTP to the specified address. If the address does not exist in the store,
// this function returns the ErrUnknownAddress error.
func (s *Sender) SendRecoveryOTP(ctx context.Context, r *http.Request, f *recovery.Flow, to string) error {
	via := identifierType(to)

	s.r.Logger().
		WithField("via", via).
		WithSensitiveField("identifier", to).
		Debug("Preparing recovery code.")

	address, err := s.r.IdentityPool().FindRecoveryAddress(ctx, to)
	switch errorsx.Cause(err) {
	case sqlcon.ErrNoRows:
		if via == identity.AddressTypeEmail {
			s.r.Audit().
				WithField("via", via).
				WithSensitiveField("identifier", address).
				Info("Sending out invalid recovery message because address is unknown.")

			if err := s.send(ctx, email.NewRecoveryInvalid(s.r, &email.RecoveryInvalidModel{To: to})); err != nil {
				return err
			}
		}
		return errors.Cause(ErrUnknownIdentifier)
	case nil:
	default:
		return err
	}

	tkn := token.NewOTPRecovery(address, f, s.r.Config(r.Context()).SelfServiceLinkMethodLifespan())
	if err := s.r.RecoveryTokenPersister().CreateRecoveryToken(ctx, tkn); err != nil {
		return err
	}

	if err := s.sendRecoveryOtpTo(ctx, address, tkn); err != nil {
		return err
	}

	return nil
}

// SendVerificationOTP sends a verification OTP to the specified address. If the address does not exist in the store,
// this function returns the ErrUnknownAddress error.
func (s *Sender) SendVerificationOTP(ctx context.Context, f *verification.Flow, to string) error {
	via := identifierType(to)

	s.r.Logger().
		WithField("via", via).
		WithSensitiveField("identifier", to).
		Debug("Preparing verification code.")

	address, err := s.r.IdentityPool().FindVerifiableAddress(ctx, to)
	switch errorsx.Cause(err) {
	case sqlcon.ErrNoRows:
		if via == identity.AddressTypeEmail {
			s.r.Audit().
				WithField("via", via).
				WithSensitiveField("identifier", address).
				Info("Sending out invalid verification message because address is unknown.")

			if err := s.send(ctx, email.NewVerificationInvalid(s.r, &email.VerificationInvalidModel{To: to})); err != nil {
				return err
			}
		}
		return errors.Cause(ErrUnknownIdentifier)
	case nil:
	default:
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

func (s *Sender) sendRecoveryOtpTo(ctx context.Context, address *identity.RecoveryAddress, tkn *token.RecoveryToken) error {
	s.r.Audit().
		WithField("via", address.Via).
		WithField("identity_id", address.IdentityID).
		WithField("recovery_token_id", tkn.ID).
		WithSensitiveField("identifier", address.Value).
		WithSensitiveField("recovery_otp_token", tkn.Token).
		Info("Sending out recovery message with OTP.")

	otpMessage, err := s.newRecoveryOtpMessage(ctx, address, tkn)
	if err != nil {
		return err
	}

	if err := s.send(ctx, otpMessage); err != nil {
		return err
	}

	return nil
}

func (s *Sender) newRecoveryOtpMessage(ctx context.Context, address *identity.RecoveryAddress, tkn *token.RecoveryToken) (interface{}, error) {
	var msg interface{}

	switch address.Via {
	case identity.RecoveryAddressTypePhone:
		msg = templates.NewRecoveryOTPMessage(s.r, &templates.RecoveryMessageModel{To: address.Value, Code: tkn.Token})
	case identity.AddressTypeEmail:
		fallthrough
	default:
		i, err := s.r.IdentityPool().GetIdentity(ctx, address.IdentityID)
		if err != nil {
			return nil, err
		}

		model, err := x.StructToMap(i)
		if err != nil {
			return nil, err
		}

		msg = email.NewRecoveryValidOTP(s.r, &email.RecoveryValidOTPModel{To: address.Value,
			Code: tkn.Token, Identity: model})
	}

	return msg, nil
}

func (s *Sender) SendVerificationTokenTo(ctx context.Context, address *identity.VerifiableAddress, tkn *token.VerificationToken) error {
	s.r.Audit().
		WithField("via", address.Via).
		WithField("identity_id", address.IdentityID).
		WithField("verification_token_id", tkn.ID).
		WithSensitiveField("identifier", address.Value).
		WithSensitiveField("verification_otp_token", tkn.Token).
		Info("Sending out verification message with OTP.")

	otpMessage, err := s.newVerificationTokenMessage(ctx, address, tkn)
	if err != nil {
		return err
	}

	if err := s.send(ctx, otpMessage); err != nil {
		return err
	}

	address.Status = identity.VerifiableAddressStatusSent

	if err := s.r.PrivilegedIdentityPool().UpdateVerifiableAddress(ctx, address); err != nil {
		return err
	}

	return nil
}

func (s *Sender) newVerificationTokenMessage(ctx context.Context, address *identity.VerifiableAddress, tkn *token.VerificationToken) (interface{}, error) {
	var msg interface{}

	switch address.Via {
	case identity.VerifiableAddressTypePhone:
		msg = templates.NewVerificationOTPMessage(s.r, &templates.VerificationMessageModel{To: address.Value, Code: tkn.Token})
	case identity.AddressTypeEmail:
		fallthrough
	default:
		i, err := s.r.IdentityPool().GetIdentity(ctx, address.IdentityID)
		if err != nil {
			return nil, err
		}

		model, err := x.StructToMap(i)
		if err != nil {
			return nil, err
		}

		msg = email.NewVerificationValidOTP(s.r, &email.VerificationValidOTPModel{To: address.Value,
			Code: tkn.Token, Identity: model})
	}

	return msg, nil
}

func (s *Sender) send(ctx context.Context, t interface{}) error {
	switch tt := t.(type) {
	case courier.SMSTemplate:
		_, err := s.r.Courier(ctx).QueueSMS(ctx, tt)
		return err
	case courier.EmailTemplate:
		_, err := s.r.Courier(ctx).QueueEmail(ctx, tt)
		return err
	default:
		return errors.Errorf("received unexpected template type: %s", reflect.TypeOf(t).String())
	}
}

func identifierType(identifier string) identity.VerifiableAddressType {
	if strings.Contains(identifier, "@") {
		return identity.AddressTypeEmail
	}

	return identity.AddressTypePhone
}
