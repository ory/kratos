package testhelpers

import (
	"context"
	"testing"

	db "github.com/gofrs/uuid"

	courier "github.com/ory/kratos/courier/test"
	"github.com/ory/kratos/internal/testhelpers"
	"github.com/ory/kratos/persistence"
)

func DefaultNetworkWrapper(p persistence.Persister) (courier.NetworkWrapper, courier.NetworkWrapper) {
	return func(t *testing.T, ctx context.Context) (db.UUID, courier.PersisterWrapper) {
			return testhelpers.NewNetworkUnlessExisting(t, ctx, p)
		}, func(t *testing.T, ctx context.Context) (db.UUID, courier.PersisterWrapper) {
			return testhelpers.NewNetwork(t, ctx, p)
		}
}
