package code

import (
	"context"
	"net/http"

	"github.com/hashicorp/go-retryablehttp"

	"github.com/ory/herodot"
	"github.com/ory/kratos/courier/template/email"

	"github.com/ory/x/httpx"
	"github.com/ory/x/stringsx"

	"github.com/ory/kratos/courier"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/flow/recovery"
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

		HTTPClient(ctx context.Context, opts ...httpx.ResilientOptions) *retryablehttp.Client
	}
	RecoveryCodeSenderProvider interface {
		RecoveryCodeSender() *RecoveryCodeSender
	}

	RecoveryCodeSender struct {
		deps senderDependencies
	}
)

var ErrUnknownAddress = herodot.ErrNotFound.WithReason("recovery requested for unknown address")

func NewSender(deps senderDependencies) *RecoveryCodeSender {
	return &RecoveryCodeSender{deps: deps}
}

// SendRecoveryCode sends a recovery code to the specified address.
// If the address does not exist in the store, an email is still being sent to prevent account
// enumeration attacks. In that case, this function returns the ErrUnknownAddress error.
func (s *RecoveryCodeSender) SendRecoveryCode(ctx context.Context, r *http.Request, f *recovery.Flow, via identity.VerifiableAddressType, to string) error {
	s.deps.Logger().
		WithField("via", via).
		WithSensitiveField("address", to).
		Debug("Preparing recovery code.")

	address, err := s.deps.IdentityPool().FindRecoveryAddressByValue(ctx, identity.RecoveryAddressTypeEmail, to)
	if err != nil {
		if err := s.send(ctx, string(via), email.NewRecoveryInvalid(s.deps, &email.RecoveryInvalidModel{To: to})); err != nil {
			return err
		}
		return ErrUnknownAddress
	}

	// Get the identity associated with the recovery address
	i, err := s.deps.IdentityPool().GetIdentity(ctx, address.IdentityID)
	if err != nil {
		return err
	}

	rawCode := GenerateRecoveryCode()

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

	return s.SendRecoveryCodeTo(ctx, i, rawCode, code)
}

func (s *RecoveryCodeSender) SendRecoveryCodeTo(ctx context.Context, i *identity.Identity, codeString string, code *RecoveryCode) error {
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

func (s *RecoveryCodeSender) send(ctx context.Context, via string, t courier.EmailTemplate) error {
	switch f := stringsx.SwitchExact(via); {
	case f.AddCase(identity.AddressTypeEmail):
		_, err := s.deps.Courier(ctx).QueueEmail(ctx, t)
		return err
	default:
		return f.ToUnknownCaseErr()
	}
}
