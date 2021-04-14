package testhelpers

import (
	"context"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/persistence"
	"github.com/ory/x/networkx"
)

func NewNetworkUnlessExisting(t *testing.T, ctx context.Context, p persistence.Persister) (uuid.UUID, persistence.Persister) {
	if n, ok := ctx.Value("network").(*networkx.Network); ok && n != nil {
		return n.ID, p.WithNetworkID(n.ID)
	}

	n := networkx.NewNetwork()
	require.NoError(t, p.GetConnection(context.Background()).Create(n))
	return n.ID, p.WithNetworkID(n.ID)
}

func NewNetwork(t *testing.T, ctx context.Context, p persistence.Persister) (uuid.UUID, persistence.Persister) {
	SkipIfNetworkContext(t, ctx)

	n := networkx.NewNetwork()
	require.NoError(t, p.GetConnection(context.Background()).Create(n))
	return n.ID, p.WithNetworkID(n.ID)
}

func ExistingNetwork(t *testing.T, p persistence.Persister, id uuid.UUID) persistence.Persister {
	require.NoError(t, p.GetConnection(context.Background()).Save(&networkx.Network{ID: id}))
	return p.WithNetworkID(id)
}

func SkipIfNetworkContext(t *testing.T, ctx context.Context) {
	a, b := ctx.Value("network").(*networkx.Network)
	if a != nil && b {
		t.Skip("Network was set in context")
	}
}
