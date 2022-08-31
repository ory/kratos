package code

import (
	"context"
	"net/http"

	"github.com/hashicorp/go-retryablehttp"

	"github.com/ory/kratos/courier/template/email"

	"github.com/ory/x/httpx"

	"github.com/pkg/errors"

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

var ErrUnknownAddress = errors.New("verification requested for unknown address")

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
		Debug("Preparing verification code.")

	address, err := s.deps.IdentityPool().FindRecoveryAddressByValue(ctx, identity.RecoveryAddressTypeEmail, to)
	if err != nil {
		if err := s.send(ctx, string(via), email.NewRecoveryInvalid(s.deps, &email.RecoveryInvalidModel{To: to})); err != nil {
			return err
		}
		return errors.Cause(ErrUnknownAddress)
	}

	// Get the identity associated with the recovery address
	i, err := s.deps.IdentityPool().GetIdentity(ctx, address.IdentityID)
	if err != nil {
		return err
	}

	code := NewSelfServiceRecoveryCode(i.ID, address, f, s.deps.Config().SelfServiceCodeMethodLifespan(r.Context()))
	if err := s.deps.RecoveryCodePersister().CreateRecoveryCode(ctx, code); err != nil {
		return err
	}

	if err := s.SendRecoveryCodeTo(ctx, f, i, address, code); err != nil {
		return err
	}

	return nil
}

func (s *RecoveryCodeSender) SendRecoveryCodeTo(ctx context.Context, f *recovery.Flow, i *identity.Identity, address *identity.RecoveryAddress, code *RecoveryCode) error {
	s.deps.Audit().
		WithField("via", address.Via).
		WithField("identity_id", address.IdentityID).
		WithField("recovery_code_id", code.ID).
		WithSensitiveField("email_address", address.Value).
		WithSensitiveField("recovery_code", code.Code).
		Info("Sending out recovery email with recovery link.")

	model, err := x.StructToMap(i)
	if err != nil {
		return err
	}

	emailModel := email.RecoveryCodeValidModel{
		To:           address.Value,
		RecoveryCode: code.Code,
		Identity:     model,
	}

	return s.send(ctx, string(address.Via), email.NewRecoveryCodeValid(s.deps, &emailModel))
}

func (s *RecoveryCodeSender) send(ctx context.Context, via string, t courier.EmailTemplate) error {
	switch via {
	case identity.AddressTypeEmail:
		_, err := s.deps.Courier(ctx).QueueEmail(ctx, t)
		return err
	default:
		return errors.Errorf("received unexpected via type: %s", via)
	}
}
