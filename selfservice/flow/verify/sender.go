package verify

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

func (m *Sender) SendCode(ctx context.Context, via identity.VerifiableAddressType, value string) error {
	m.r.Logger().WithField("via", via).Debug("Sending out verification code.")

	address, err := m.r.IdentityPool().FindAddressByValue(ctx, via, value)
	if err != nil {
		if errorsx.Cause(err) == sqlcon.ErrNoRows {
			if err := m.sendToUnknownAddress(ctx, identity.VerifiableAddressTypeEmail, value); err != nil {
				return err
			}
			return nil
		}
		return err
	}

	if err := m.r.IdentityManager().RefreshVerifyAddress(ctx, address); err != nil {
		return err
	}

	return m.sendCodeToKnownAddress(ctx, address)
}

func (m *Sender) sendToUnknownAddress(ctx context.Context, via identity.VerifiableAddressType, address string) error {
	m.r.Logger().WithField("via", via).Debug("Sending out invalid verification email because address is unknown.")
	return m.run(via, func() error {
		_, err := m.r.Courier().QueueEmail(ctx,
			templates.NewVerifyInvalid(m.c, &templates.VerifyInvalidModel{To: address}))
		return err
	})
}

func (m *Sender) sendCodeToKnownAddress(ctx context.Context, address *identity.VerifiableAddress) error {
	m.r.Logger().WithField("via", address.Via).Debug("Sending out verification email.")
	return m.run(address.Via, func() error {
		_, err := m.r.Courier().QueueEmail(ctx, templates.NewVerifyValid(m.c,
			&templates.VerifyValidModel{
				To: address.Value,
				VerifyURL: urlx.AppendPaths(
					m.c.SelfPublicURL(),
					strings.Replace(
						PublicVerificationConfirmPath, ":code",
						address.Code, 1,
					)).String(),
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
