package testhelpers

import (
	"context"
	"github.com/gofrs/uuid"
	"github.com/ory/kratos/persistence"
	"github.com/ory/x/networkx"
	"github.com/stretchr/testify/require"
	"testing"
)

func NewNetwork(t *testing.T, p persistence.Persister) (uuid.UUID, persistence.Persister) {
	n := networkx.NewNetwork()
	require.NoError(t, p.GetConnection(context.Background()).Create(n))
	return n.ID, p.WithNetworkID(n.ID)
}

func ExistingNetwork(t *testing.T, p persistence.Persister, id uuid.UUID) persistence.Persister {
	require.NoError(t, p.GetConnection(context.Background()).Save(&networkx.Network{ID: id}))
	return p.WithNetworkID(id)
}
