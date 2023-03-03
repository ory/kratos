// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package devices

import (
	"context"

	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"

	"github.com/ory/kratos/session"
	"github.com/ory/x/contextx"
	"github.com/ory/x/popx"
	"github.com/ory/x/sqlcon"
)

var _ session.DevicePersister = (*DevicePersister)(nil)

type DevicePersister struct {
	ctxer contextx.Provider
	c     *pop.Connection
	nid   uuid.UUID
}

func NewPersister(r contextx.Provider, c *pop.Connection) *DevicePersister {
	return &DevicePersister{
		ctxer: r,
		c:     c,
	}
}

func (p *DevicePersister) NetworkID(ctx context.Context) uuid.UUID {
	return p.ctxer.Contextualizer().Network(ctx, p.nid)
}

func (p DevicePersister) WithNetworkID(nid uuid.UUID) session.DevicePersister {
	p.nid = nid
	return &p
}

func (p *DevicePersister) CreateDevice(ctx context.Context, d *session.Device) error {
	d.NID = p.NetworkID(ctx)
	return sqlcon.HandleError(popx.GetConnection(ctx, p.c.WithContext(ctx)).Create(d))
}
