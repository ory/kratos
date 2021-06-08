package corp

import (
	"context"
	"net/http"

	"github.com/gofrs/uuid"

	"github.com/ory/kratos/driver/config"
)

type NoopContextualizer struct{}

func (*NoopContextualizer) ContextualizeTableName(_ context.Context, name string) string {
	return name
}

func (*NoopContextualizer) ContextualizeMiddleware(_ context.Context) func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	return func(w http.ResponseWriter, r *http.Request, n http.HandlerFunc) {
		n(w, r)
	}
}

func (*NoopContextualizer) ContextualizeConfig(ctx context.Context, fb *config.Config) *config.Config {
	return fb
}

func (*NoopContextualizer) ContextualizeNID(_ context.Context, fallback uuid.UUID) uuid.UUID {
	return fallback
}
