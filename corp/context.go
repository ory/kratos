package corp

import (
	"context"
	"net/http"

	"github.com/gofrs/uuid"

	"github.com/ory/kratos/driver/config"
)

type Contextualizer interface {
	ContextualizeTableName(ctx context.Context, name string) string
	ContextualizeMiddleware(ctx context.Context) func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc)
	ContextualizeConfig(ctx context.Context, fb *config.Config) *config.Config
	ContextualizeNID(ctx context.Context, fallback uuid.UUID) uuid.UUID
}

var DefaultContextualizer Contextualizer = nil

// These global functions call the respective method on DefaultContextualizer

func ContextualizeTableName(ctx context.Context, name string) string {
	return DefaultContextualizer.ContextualizeTableName(ctx, name)
}

func ContextualizeMiddleware(ctx context.Context) func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	return DefaultContextualizer.ContextualizeMiddleware(ctx)
}

func ContextualizeConfig(ctx context.Context, fb *config.Config) *config.Config {
	return DefaultContextualizer.ContextualizeConfig(ctx, fb)
}

func ContextualizeNID(ctx context.Context, fallback uuid.UUID) uuid.UUID {
	return DefaultContextualizer.ContextualizeNID(ctx, fallback)
}
