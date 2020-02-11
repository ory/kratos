package verify

import (
	"context"
	"strings"
	"time"

	"github.com/ory/x/errorsx"
	"github.com/ory/x/sqlcon"
	"github.com/pkg/errors"

	"github.com/ory/x/urlx"

	"github.com/ory/kratos/courier"
	templates "github.com/ory/kratos/courier/template"
	"github.com/ory/kratos/driver/configuration"
)

type (
	managerDependencies interface {
		PersistenceProvider
		courier.Provider
	}
	ManagementProvider interface {
		VerificationManager() *Manager
	}
	Manager struct {
		r managerDependencies
		c configuration.Provider
	}
)

func NewManager(r managerDependencies, c configuration.Provider) *Manager {
	return &Manager{r: r, c: c}
}

func (m *Manager) sendToUnknownAddress(ctx context.Context, via Via, address string) error {
	return m.run(via, func() error {
		_, err := m.r.Courier().QueueEmail(ctx,
			templates.NewVerifyInvalid(m.c, &templates.VerifyInvalidModel{To: address}))
		return err
	})
}

func (m *Manager) sendCodeToKnownAddress(ctx context.Context, address *Address) error {
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

func (m *Manager) SendCode(ctx context.Context, via Via, value string) error {
	address, err := m.r.VerificationPersister().FindAddressByValue(ctx, via, value)
	if err != nil {
		if errorsx.Cause(err) == sqlcon.ErrNoRows {
			if err := m.sendToUnknownAddress(ctx, ViaEmail, value); err != nil {
				return err
			}
			return nil
		}
		return err
	}

	code, err := NewVerifyCode()
	if err != nil {
		return err
	}

	address.Code = code
	address.ExpiresAt = time.Now().UTC().Add(m.c.SelfServiceVerificationLinkLifespan())
	if err := m.r.VerificationPersister().UpdateAddress(ctx, address); err != nil {
		return err
	}

	return m.sendCodeToKnownAddress(ctx, address)
}

func (m *Manager) TrackAndSend(ctx context.Context, addresses []Address) error {
	if err := m.r.VerificationPersister().TrackAddresses(ctx, addresses); err != nil {
		return err
	}

	for _, address := range addresses {
		if err := m.SendCode(ctx, address.Via, address.Value); err != nil {
			return err
		}
	}
	return nil
}

func (m *Manager) run(via Via, emailFunc func() error) error {
	switch via {
	case ViaEmail:
		return emailFunc()
	default:
		return errors.Errorf("received unexpected via type: %s", via)
	}
}
