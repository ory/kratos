// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package code

import (
	"context"
	"net/url"

	"github.com/gofrs/uuid"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/pkg/errors"

	"github.com/ory/herodot"
	"github.com/ory/kratos/courier/template/email"

	"github.com/ory/x/httpx"
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

		HTTPClient(ctx context.Context, opts ...httpx.ResilientOptions) *retryablehttp.Client
	}
	SenderProvider interface {
		CodeSender() *Sender
	}

	Sender struct {
		deps senderDependencies
	}
	Address struct {
		To  string
		Via identity.CodeAddressType
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

	// send to all addresses
	for _, address := range addresses {
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

			emailModel := email.RegistrationCodeValidModel{
				To:               address.To,
				RegistrationCode: rawCode,
				Traits:           model,
			}

			s.deps.Audit().
				WithField("registration_flow_id", code.FlowID).
				WithField("registration_code_id", code.ID).
				WithSensitiveField("registration_code", rawCode).
				Info("Sending out registration email with code.")

			if err := s.send(ctx, string(address.Via), email.NewRegistrationCodeValid(s.deps, &emailModel)); err != nil {
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

			emailModel := email.LoginCodeValidModel{
				To:        address.To,
				LoginCode: rawCode,
				Identity:  model,
			}
			s.deps.Audit().
				WithField("login_flow_id", code.FlowID).
				WithField("login_code_id", code.ID).
				WithSensitiveField("login_code", rawCode).
				Info("Sending out login email with code.")

			if err := s.send(ctx, string(address.Via), email.NewLoginCodeValid(s.deps, &emailModel)); err != nil {
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
		if !notifyUnknownRecipients {
			// do nothing
		} else if err := s.send(ctx, string(via), email.NewRecoveryCodeInvalid(s.deps, &email.RecoveryCodeInvalidModel{To: to})); err != nil {
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

	return s.SendRecoveryCodeTo(ctx, i, rawCode, code)
}

func (s *Sender) SendRecoveryCodeTo(ctx context.Context, i *identity.Identity, codeString string, code *RecoveryCode) error {
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

	emailModel := email.RecoveryCodeValidModel{
		To:           code.RecoveryAddress.Value,
		RecoveryCode: codeString,
		Identity:     model,
	}

	return s.send(ctx, string(code.RecoveryAddress.Via), email.NewRecoveryCodeValid(s.deps, &emailModel))
}

// SendVerificationCode sends a verification code & link to the specified address
//
// If the address does not exist in the store and dispatching invalid emails is enabled (CourierEnableInvalidDispatch is
// true), an email is still being sent to prevent account enumeration attacks. In that case, this function returns the
// ErrUnknownAddress error.
func (s *Sender) SendVerificationCode(ctx context.Context, f *verification.Flow, via identity.VerifiableAddressType, to string) error {
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
			WithSensitiveField("email_address", address).
			WithField("was_notified", notifyUnknownRecipients).
			Info("Address verification was requested for an unknown address.")
		if !notifyUnknownRecipients {
			// do nothing
		} else if err := s.send(ctx, string(via), email.NewVerificationCodeInvalid(s.deps, &email.VerificationCodeInvalidModel{To: to})); err != nil {
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

	if err := s.SendVerificationCodeTo(ctx, f, i, rawCode, code); err != nil {
		return err
	}
	return nil
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

	if err := s.send(ctx, string(code.VerifiableAddress.Via), email.NewVerificationCodeValid(s.deps,
		&email.VerificationCodeValidModel{
			To:               code.VerifiableAddress.Value,
			VerificationURL:  s.constructVerificationLink(ctx, f.ID, codeString),
			Identity:         model,
			VerificationCode: codeString,
		})); err != nil {
		return err
	}
	code.VerifiableAddress.Status = identity.VerifiableAddressStatusSent
	return s.deps.PrivilegedIdentityPool().UpdateVerifiableAddress(ctx, code.VerifiableAddress)
}

func (s *Sender) send(ctx context.Context, via string, t courier.EmailTemplate) error {
	switch f := stringsx.SwitchExact(via); {
	case f.AddCase(identity.AddressTypeEmail):
		c, err := s.deps.Courier(ctx)
		if err != nil {
			return err
		}

		_, err = c.QueueEmail(ctx, t)
		return err
	default:
		return f.ToUnknownCaseErr()
	}
}
