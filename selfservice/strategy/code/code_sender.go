// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package code

import (
	"context"
	"net/url"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"

	"github.com/ory/herodot"
	"github.com/ory/kratos/courier/template/email"
	"github.com/ory/kratos/courier/template/sms"

	"github.com/ory/x/sqlcon"
	"github.com/ory/x/stringsx"
	"github.com/ory/x/urlx"

	"github.com/ory/kratos/courier"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/flow"
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

		RecoveryCodePersistenceProvider
		VerificationCodePersistenceProvider
		RegistrationCodePersistenceProvider
		LoginCodePersistenceProvider

		x.HTTPClientProvider
	}
	SenderProvider interface {
		CodeSender() *Sender
	}

	Sender struct {
		deps senderDependencies
	}
	Address struct {
		To  string
		Via identity.CodeChannel
	}
)

var ErrUnknownAddress = herodot.ErrNotFound.WithReason("recovery requested for unknown address")

func NewSender(deps senderDependencies) *Sender {
	return &Sender{deps: deps}
}

func (s *Sender) SendCode(ctx context.Context, f flow.Flow, id *identity.Identity, addresses ...Address) error {
	s.deps.Logger().
		WithSensitiveField("address", addresses).
		Debugf("Preparing %s code", f.GetFlowName())

	transientPayload, err := x.ParseRawMessageOrEmpty(f.GetTransientPayload())
	if err != nil {
		return errors.WithStack(err)
	}

	// send to all addresses
	for _, address := range addresses {
		// We have to generate a unique code per address, or otherwise it is not possible to link which
		// address was used to verify the code.
		//
		// See also [this discussion](https://github.com/ory/kratos/pull/3456#discussion_r1307560988).
		rawCode := GenerateCode()

		switch f.GetFlowName() {
		case flow.RegistrationFlow:
			code, err := s.deps.
				RegistrationCodePersister().
				CreateRegistrationCode(ctx, &CreateRegistrationCodeParams{
					AddressType: address.Via,
					RawCode:     rawCode,
					ExpiresIn:   s.deps.Config().SelfServiceCodeMethodLifespan(ctx),
					FlowID:      f.GetID(),
					Address:     address.To,
				})
			if err != nil {
				return err
			}
			model, err := x.StructToMap(id.Traits)
			if err != nil {
				return err
			}

			s.deps.Audit().
				WithField("registration_flow_id", code.FlowID).
				WithField("registration_code_id", code.ID).
				WithSensitiveField("registration_code", rawCode).
				Info("Sending out registration email with code.")

			var t courier.Template
			switch address.Via {
			case identity.ChannelTypeEmail:
				t = email.NewRegistrationCodeValid(s.deps, &email.RegistrationCodeValidModel{
					To:               address.To,
					RegistrationCode: rawCode,
					Traits:           model,
					RequestURL:       f.GetRequestURL(),
					TransientPayload: transientPayload,
					ExpiresInMinutes: int(s.deps.Config().SelfServiceCodeMethodLifespan(ctx).Minutes()),
				})
			case identity.ChannelTypeSMS:
				t = sms.NewRegistrationCodeValid(s.deps, &sms.RegistrationCodeValidModel{
					To:               address.To,
					RegistrationCode: rawCode,
					Identity:         model,
					RequestURL:       f.GetRequestURL(),
					TransientPayload: transientPayload,
					ExpiresInMinutes: int(s.deps.Config().SelfServiceCodeMethodLifespan(ctx).Minutes()),
				})
			}

			if err := s.send(ctx, string(address.Via), t); err != nil {
				return errors.WithStack(err)
			}

		case flow.LoginFlow:
			code, err := s.deps.
				LoginCodePersister().
				CreateLoginCode(ctx, &CreateLoginCodeParams{
					AddressType: address.Via,
					Address:     address.To,
					RawCode:     rawCode,
					ExpiresIn:   s.deps.Config().SelfServiceCodeMethodLifespan(ctx),
					FlowID:      f.GetID(),
					IdentityID:  id.ID,
				})
			if err != nil {
				return err
			}

			model, err := x.StructToMap(id)
			if err != nil {
				return err
			}
			s.deps.Audit().
				WithField("login_flow_id", code.FlowID).
				WithField("login_code_id", code.ID).
				WithSensitiveField("login_code", rawCode).
				Info("Sending out login email with code.")

			var t courier.Template
			switch address.Via {
			case identity.ChannelTypeEmail:
				t = email.NewLoginCodeValid(s.deps, &email.LoginCodeValidModel{
					To:               address.To,
					LoginCode:        rawCode,
					Identity:         model,
					RequestURL:       f.GetRequestURL(),
					TransientPayload: transientPayload,
					ExpiresInMinutes: int(s.deps.Config().SelfServiceCodeMethodLifespan(ctx).Minutes()),
				})
			case identity.ChannelTypeSMS:
				t = sms.NewLoginCodeValid(s.deps, &sms.LoginCodeValidModel{
					To:               address.To,
					LoginCode:        rawCode,
					Identity:         model,
					RequestURL:       f.GetRequestURL(),
					TransientPayload: transientPayload,
					ExpiresInMinutes: int(s.deps.Config().SelfServiceCodeMethodLifespan(ctx).Minutes()),
				})
			}

			if err := s.send(ctx, string(address.Via), t); err != nil {
				return errors.WithStack(err)
			}

		default:
			return errors.WithStack(errors.New("received unknown flow type"))

		}
	}
	return nil
}

// SendRecoveryCode sends a recovery code to the specified address
//
// If the address does not exist in the store and dispatching invalid emails is enabled (CourierEnableInvalidDispatch is
// true), an email is still being sent to prevent account enumeration attacks. In that case, this function returns the
// ErrUnknownAddress error.
func (s *Sender) SendRecoveryCode(ctx context.Context, f *recovery.Flow, via identity.VerifiableAddressType, to string) error {
	s.deps.Logger().
		WithField("via", via).
		WithSensitiveField("address", to).
		Debug("Preparing recovery code.")

	address, err := s.deps.IdentityPool().FindRecoveryAddressByValue(ctx, identity.RecoveryAddressTypeEmail, to)
	if errors.Is(err, sqlcon.ErrNoRows) {
		notifyUnknownRecipients := s.deps.Config().SelfServiceFlowRecoveryNotifyUnknownRecipients(ctx)
		s.deps.Audit().
			WithField("via", via).
			WithSensitiveField("email_address", address).
			WithField("strategy", "code").
			WithField("was_notified", notifyUnknownRecipients).
			Info("Account recovery was requested for an unknown address.")

		transientPayload, err := x.ParseRawMessageOrEmpty(f.GetTransientPayload())
		if err != nil {
			return errors.WithStack(err)
		}
		if !notifyUnknownRecipients {
			// do nothing
		} else if err := s.send(ctx, string(via), email.NewRecoveryCodeInvalid(s.deps, &email.RecoveryCodeInvalidModel{
			To:               to,
			RequestURL:       f.RequestURL,
			TransientPayload: transientPayload,
		})); err != nil {
			return err
		}
		return errors.WithStack(ErrUnknownAddress)
	} else if err != nil {
		// DB error
		return err
	}

	// Get the identity associated with the recovery address
	i, err := s.deps.IdentityPool().GetIdentity(ctx, address.IdentityID, identity.ExpandDefault)
	if err != nil {
		return err
	}

	rawCode := GenerateCode()

	var code *RecoveryCode
	if code, err = s.deps.
		RecoveryCodePersister().
		CreateRecoveryCode(ctx, &CreateRecoveryCodeParams{
			RawCode:         rawCode,
			CodeType:        RecoveryCodeTypeSelfService,
			ExpiresIn:       s.deps.Config().SelfServiceCodeMethodLifespan(ctx),
			RecoveryAddress: address,
			FlowID:          f.ID,
			IdentityID:      i.ID,
		}); err != nil {
		return err
	}

	return s.SendRecoveryCodeTo(ctx, i, rawCode, code, f)
}

func (s *Sender) SendRecoveryCodeTo(ctx context.Context, i *identity.Identity, codeString string, code *RecoveryCode, f *recovery.Flow) error {
	s.deps.Audit().
		WithField("via", code.RecoveryAddress.Via).
		WithField("identity_id", code.RecoveryAddress.IdentityID).
		WithField("recovery_code_id", code.ID).
		WithSensitiveField("email_address", code.RecoveryAddress.Value).
		WithSensitiveField("recovery_code", codeString).
		Info("Sending out recovery email with recovery code.")

	model, err := x.StructToMap(i)
	if err != nil {
		return err
	}

	transientPayload, err := x.ParseRawMessageOrEmpty(f.GetTransientPayload())
	if err != nil {
		return errors.WithStack(err)
	}

	emailModel := email.RecoveryCodeValidModel{
		To:               code.RecoveryAddress.Value,
		RecoveryCode:     codeString,
		Identity:         model,
		RequestURL:       f.GetRequestURL(),
		TransientPayload: transientPayload,
		ExpiresInMinutes: int(s.deps.Config().SelfServiceCodeMethodLifespan(ctx).Minutes()),
	}

	return s.send(ctx, string(code.RecoveryAddress.Via), email.NewRecoveryCodeValid(s.deps, &emailModel))
}

// SendVerificationCode sends a verification code & link to the specified address
//
// If the address does not exist in the store and dispatching invalid emails is enabled (CourierEnableInvalidDispatch is
// true), an email is still being sent to prevent account enumeration attacks. In that case, this function returns the
// ErrUnknownAddress error.
func (s *Sender) SendVerificationCode(ctx context.Context, f *verification.Flow, via string, to string) error {
	s.deps.Logger().
		WithField("via", via).
		WithSensitiveField("address", to).
		Debug("Preparing verification code.")

	address, err := s.deps.IdentityPool().FindVerifiableAddressByValue(ctx, via, to)
	if errors.Is(err, sqlcon.ErrNoRows) {
		notifyUnknownRecipients := s.deps.Config().SelfServiceFlowVerificationNotifyUnknownRecipients(ctx)
		s.deps.Audit().
			WithField("via", via).
			WithField("strategy", "code").
			WithSensitiveField("email_address", to).
			WithField("was_notified", notifyUnknownRecipients).
			Info("Address verification was requested for an unknown address.")

		transientPayload, err := x.ParseRawMessageOrEmpty(f.GetTransientPayload())
		if err != nil {
			return errors.WithStack(err)
		}
		if !notifyUnknownRecipients {
			// do nothing
		} else if err := s.send(ctx, string(via), email.NewVerificationCodeInvalid(s.deps, &email.VerificationCodeInvalidModel{
			To:               to,
			RequestURL:       f.GetRequestURL(),
			TransientPayload: transientPayload,
		})); err != nil {
			return err
		}
		return errors.WithStack(ErrUnknownAddress)

	} else if err != nil {
		return err
	}

	rawCode := GenerateCode()
	var code *VerificationCode
	if code, err = s.deps.VerificationCodePersister().CreateVerificationCode(ctx, &CreateVerificationCodeParams{
		RawCode:           rawCode,
		ExpiresIn:         s.deps.Config().SelfServiceCodeMethodLifespan(ctx),
		VerifiableAddress: address,
		FlowID:            f.ID,
	}); err != nil {
		return err
	}

	// Get the identity associated with the recovery address
	i, err := s.deps.IdentityPool().GetIdentity(ctx, address.IdentityID, identity.ExpandDefault)
	if err != nil {
		return err
	}

	return s.SendVerificationCodeTo(ctx, f, i, rawCode, code)
}

func (s *Sender) constructVerificationLink(ctx context.Context, fID uuid.UUID, codeStr string) string {
	return urlx.CopyWithQuery(
		urlx.AppendPaths(s.deps.Config().SelfServiceLinkMethodBaseURL(ctx), verification.RouteSubmitFlow),
		url.Values{
			"flow": {fID.String()},
			"code": {codeStr},
		}).String()
}

func (s *Sender) SendVerificationCodeTo(ctx context.Context, f *verification.Flow, i *identity.Identity, codeString string, code *VerificationCode) error {
	s.deps.Audit().
		WithField("via", code.VerifiableAddress.Via).
		WithField("identity_id", i.ID).
		WithField("verification_code_id", code.ID).
		WithSensitiveField("email_address", code.VerifiableAddress.Value).
		WithSensitiveField("verification_link_token", codeString).
		Info("Sending out verification email with verification code.")

	model, err := x.StructToMap(i)
	if err != nil {
		return err
	}

	transientPayload, err := x.ParseRawMessageOrEmpty(f.GetTransientPayload())
	if err != nil {
		return errors.WithStack(err)
	}

	var t courier.Template

	// TODO: this can likely be abstracted by making templates not specific to the channel they're using
	switch code.VerifiableAddress.Via {
	case identity.ChannelTypeEmail:
		t = email.NewVerificationCodeValid(s.deps, &email.VerificationCodeValidModel{
			To:               code.VerifiableAddress.Value,
			VerificationURL:  s.constructVerificationLink(ctx, f.ID, codeString),
			Identity:         model,
			VerificationCode: codeString,
			RequestURL:       f.GetRequestURL(),
			TransientPayload: transientPayload,
			ExpiresInMinutes: int(s.deps.Config().SelfServiceCodeMethodLifespan(ctx).Minutes()),
		})
	case identity.ChannelTypeSMS:
		t = sms.NewVerificationCodeValid(s.deps, &sms.VerificationCodeValidModel{
			To:               code.VerifiableAddress.Value,
			VerificationCode: codeString,
			Identity:         model,
			RequestURL:       f.GetRequestURL(),
			TransientPayload: transientPayload,
			ExpiresInMinutes: int(s.deps.Config().SelfServiceCodeMethodLifespan(ctx).Minutes()),
		})
	default:
		return errors.WithStack(herodot.ErrInternalServerError.WithReasonf("Expected email or sms but got %s", code.VerifiableAddress.Via))
	}

	if err := s.send(ctx, string(code.VerifiableAddress.Via), t); err != nil {
		return err
	}
	code.VerifiableAddress.Status = identity.VerifiableAddressStatusSent
	return s.deps.PrivilegedIdentityPool().UpdateVerifiableAddress(ctx, code.VerifiableAddress)
}

func (s *Sender) send(ctx context.Context, via string, t courier.Template) error {
	switch f := stringsx.SwitchExact(via); {
	case f.AddCase(identity.ChannelTypeEmail):
		c, err := s.deps.Courier(ctx)
		if err != nil {
			return err
		}

		t, ok := t.(courier.EmailTemplate)
		if !ok {
			return errors.WithStack(herodot.ErrInternalServerError.WithReasonf("Expected email template but got %T", t))
		}

		_, err = c.QueueEmail(ctx, t)
		return err
	case f.AddCase(identity.ChannelTypeSMS):
		c, err := s.deps.Courier(ctx)
		if err != nil {
			return err
		}

		t, ok := t.(courier.SMSTemplate)
		if !ok {
			return errors.WithStack(herodot.ErrInternalServerError.WithReasonf("Expected sms template but got %T", t))
		}

		_, err = c.QueueSMS(ctx, t)
		return err
	default:
		return f.ToUnknownCaseErr()
	}
}
