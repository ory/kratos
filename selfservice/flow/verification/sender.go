package verification

import (
	"context"
	"strings"

	"github.com/pkg/errors"

	"github.com/ory/go-convenience/urlx"
	"github.com/ory/x/errorsx"
	"github.com/ory/x/sqlcon"

	"github.com/ory/kratos/courier"
	templates "github.com/ory/kratos/courier/template"
	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/x"
)

var ErrUnknownAddress = errors.New("verification requested for unknown address")

type (
	senderDependencies interface {
		courier.Provider
		identity.PoolProvider
		identity.ManagementProvider
		x.LoggingProvider
	}
	SenderProvider interface {
		VerificationSender() *Sender
	}
	Sender struct {
		r senderDependencies
		c configuration.Provider
	}
)

func NewSender(r senderDependencies, c configuration.Provider) *Sender {
	return &Sender{r: r, c: c}
}

// SendCode sends a code to the specified address. If the address does not exist in the store, an email is
// still being sent to prevent account enumeration attacks. In that case, this function returns the ErrUnknownAddress
// error.
func (m *Sender) SendCode(ctx context.Context, via identity.VerifiableAddressType, value string) (*identity.VerifiableAddress, error) {
	m.r.Logger().
		WithField("via", via).
		WithSensitiveField("address", value).
		Debug("Preparing verification code.")

	address, err := m.r.IdentityPool().FindVerifiableAddressByValue(ctx, via, value)
	if err != nil {
		if errorsx.Cause(err) == sqlcon.ErrNoRows {
			if err := m.sendToUnknownAddress(ctx, identity.VerifiableAddressTypeEmail, value); err != nil {
				return nil, err
			}
			return nil, errors.Cause(ErrUnknownAddress)
		}
		return nil, err
	}

	if err := m.r.IdentityManager().RefreshVerifyAddress(ctx, address); err != nil {
		return nil, err
	}

	if err := m.sendCodeToKnownAddress(ctx, address); err != nil {
		return nil, err
	}
	return address, nil
}

func (m *Sender) sendToUnknownAddress(ctx context.Context, via identity.VerifiableAddressType, address string) error {
	m.r.Audit().
		WithSensitiveField("email_address", address).
		Info("Sending out invalid verification email because address is unknown.")

	return m.run(via, func() error {
		_, err := m.r.Courier().QueueEmail(ctx,
			templates.NewVerificationInvalid(m.c, &templates.VerificationInvalidModel{To: address}))
		return err
	})
}

func (m *Sender) sendCodeToKnownAddress(ctx context.Context, address *identity.VerifiableAddress) error {
	m.r.Audit().
		WithField("identity_id", address.IdentityID).
		WithSensitiveField("email_address", address.Value).
		Info("Sending out verification code via email.")

	return m.run(address.Via, func() error {
		_, err := m.r.Courier().QueueEmail(ctx, templates.NewVerificationValid(m.c,
			&templates.VerificationValidModel{
				To: address.Value,
				VerificationURL: urlx.AppendPaths(
					m.c.SelfPublicURL(),
					strings.ReplaceAll(
						strings.ReplaceAll(PublicVerificationConfirmPath, ":via", string(address.Via)),
						":code", address.Code)).
					String(),
			},
		))
		return err
	})
}

func (m *Sender) run(via identity.VerifiableAddressType, emailFunc func() error) error {
	switch via {
	case identity.VerifiableAddressTypeEmail:
		return emailFunc()
	default:
		return errors.Errorf("received unexpected via type: %s", via)
	}
}
