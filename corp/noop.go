package corp

import (
	"context"
	"net/http"

	"github.com/gofrs/uuid"
	"github.com/ory/kratos/driver/config"
)

type noopContextualizer struct{}

func (noopContextualizer) ContextualizeTableName(_ context.Context, name string) string {
	return name
}

func (noopContextualizer) ContextualizeMiddleware(_ context.Context) func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	return func(w http.ResponseWriter, r *http.Request, n http.HandlerFunc) {
		n(w, r)
	}
}

func (noopContextualizer) ContextualizeConfig(ctx context.Context, fb *config.Config) *config.Config {
	return fb
}

func (noopContextualizer) ContextualizeNID(_ context.Context, fallback uuid.UUID) uuid.UUID {
	return fallback
}
