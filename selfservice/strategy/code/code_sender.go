// Copyright © 2022 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package code

import (
	"context"
	"encoding/json"
	"net/http"
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

		HTTPClient(ctx context.Context, opts ...httpx.ResilientOptions) *retryablehttp.Client
	}
	SenderProvider interface {
		CodeSender() *Sender
	}

	Sender struct {
		deps senderDependencies
	}
)

var ErrUnknownAddress = herodot.ErrNotFound.WithReason("recovery requested for unknown address")

func NewSender(deps senderDependencies) *Sender {
	return &Sender{deps: deps}
}

// SendRecoveryCode sends a recovery code to the specified address.
// If the address does not exist in the store, an email is still being sent to prevent account
// enumeration attacks. In that case, this function returns the ErrUnknownAddress error.
func (s *Sender) SendRecoveryCode(ctx context.Context, r *http.Request, f *recovery.Flow, via identity.VerifiableAddressType, to string,
	transientPayload json.RawMessage, branding string) error {
	s.deps.Logger().
		WithField("via", via).
		WithSensitiveField("address", to).
		Debug("Preparing recovery code.")

	address, err := s.deps.IdentityPool().FindRecoveryAddressByValue(ctx, identity.RecoveryAddressTypeEmail, to)
	if err != nil {
		if err := s.send(ctx, string(via), email.NewRecoveryCodeInvalid(s.deps, &email.RecoveryCodeInvalidModel{To: to})); err != nil {
			return err
		}
		return ErrUnknownAddress
	}

	// Get the identity associated with the recovery address
	i, err := s.deps.IdentityPool().GetIdentity(ctx, address.IdentityID)
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
			ExpiresIn:       s.deps.Config().SelfServiceCodeMethodLifespan(r.Context()),
			RecoveryAddress: address,
			FlowID:          f.ID,
			IdentityID:      i.ID,
		}); err != nil {
		return err
	}

	return s.SendRecoveryCodeTo(ctx, i, rawCode, code, transientPayload, branding)
}

func (s *Sender) SendRecoveryCodeTo(ctx context.Context, i *identity.Identity, codeString string, code *RecoveryCode,
	transientPayload json.RawMessage, branding string) error {
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
		To:               code.RecoveryAddress.Value,
		RecoveryCode:     codeString,
		Identity:         model,
		TransientPayload: transientPayload,
		Branding:         branding,
	}

	return s.send(ctx, string(code.RecoveryAddress.Via), email.NewRecoveryCodeValid(s.deps, &emailModel))
}

// SendVerificationCode sends a verification link to the specified address. If the address does not exist in the store, an email is
// still being sent to prevent account enumeration attacks. In that case, this function returns the ErrUnknownAddress
// error.
func (s *Sender) SendVerificationCode(ctx context.Context, f *verification.Flow, via identity.VerifiableAddressType, to string,
	transientPayload json.RawMessage, branding string) error {
	s.deps.Logger().
		WithField("via", via).
		WithSensitiveField("address", to).
		Debug("Preparing verification code.")

	address, err := s.deps.IdentityPool().FindVerifiableAddressByValue(ctx, via, to)
	if err != nil {
		if errors.Is(err, sqlcon.ErrNoRows) {
			s.deps.Audit().
				WithField("via", via).
				WithSensitiveField("email_address", address).
				Info("Sending out invalid verification via code email because address is unknown.")
			if err := s.send(ctx, string(via), email.NewVerificationCodeInvalid(s.deps, &email.VerificationCodeInvalidModel{To: to})); err != nil {
				return err
			}
			return errors.Cause(ErrUnknownAddress)
		}
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
	i, err := s.deps.IdentityPool().GetIdentity(ctx, address.IdentityID)
	if err != nil {
		return err
	}

	if err := s.SendVerificationCodeTo(ctx, f, i, rawCode, code, transientPayload, branding); err != nil {
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

func (s *Sender) SendVerificationCodeTo(ctx context.Context, f *verification.Flow, i *identity.Identity, codeString string,
	code *VerificationCode, transientPayload json.RawMessage, branding string) error {
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
			TransientPayload: transientPayload,
			Branding:         branding,
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
