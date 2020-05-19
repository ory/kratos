package recovery

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

var ErrUnknownAddress = errors.New("recovery requested for unknown address")

type (
	senderDependencies interface {
		courier.Provider
		identity.PoolProvider
		identity.ManagementProvider
		x.LoggingProvider
	}
	SenderProvider interface {
		RecoverySender() *Sender
	}
	Sender struct {
		r senderDependencies
		c configuration.Provider
	}
)

func NewSender(r senderDependencies, c configuration.Provider) *Sender {
	return &Sender{r: r, c: c}
}

func (m *Sender) IssueAndSendRecoveryToken(ctx context.Context, address, via string) (*identity.VerifiableAddress, error) {
	m.r.Logger().WithField("via", via).Debug("Sending out verification code.")

	a, err := m.r.IdentityPool().FindAddressByValue(ctx, identity.VerifiableAddressTypeEmail, address)
	if err != nil {
		if errorsx.Cause(err) == sqlcon.ErrNoRows {
			if err := m.sendToUnknownAddress(ctx, identity.VerifiableAddressTypeEmail, address); err != nil {
				return nil, err
			}
			return nil, errors.Cause(ErrUnknownAddress)
		}
		return nil, err
	}

	if err := m.r.IdentityManager().RefreshVerifyAddress(ctx, a); err != nil {
		return nil, err
	}

	if err := m.sendCodeToKnownAddress(ctx, a); err != nil {
		return nil, err
	}
	return a, nil
}

func (m *Sender) sendToUnknownAddress(ctx context.Context, via identity.VerifiableAddressType, address string) error {
	m.r.Logger().WithField("via", via).Debug("Sending out unsuccessful recovery message because address is unknown.")
	return m.run(via, func() error {
		_, err := m.r.Courier().QueueEmail(ctx,
			templates.NewVerifyInvalid(m.c, &templates.VerifyInvalidModel{To: address}))
		return err
	})
}

func (m *Sender) sendCodeToKnownAddress(ctx context.Context, address *identity.VerifiableAddress) error {
	m.r.Logger().WithField("via", address.Via).Debug("Sending out recovery message.")
	return m.run(address.Via, func() error {
		_, err := m.r.Courier().QueueEmail(ctx, templates.NewVerifyValid(m.c,
			&templates.VerifyValidModel{
				To: address.Value,
				VerifyURL: urlx.AppendPaths(
					m.c.SelfPublicURL(),
					strings.ReplaceAll(
						strings.ReplaceAll(PublicRecoveryConfirmPath, ":via", string(address.Via)),
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
