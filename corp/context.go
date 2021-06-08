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

var Context Contextualizer = nil

func SetContextualizer(c Contextualizer) {
	if _, ok := c.(*ContextNoOp); ok && Context != nil {
		panic("contextualizer was already set")
	}

	Context = c
}

// These global functions call the respective method on Context

func ContextualizeTableName(ctx context.Context, name string) string {
	return Context.ContextualizeTableName(ctx, name)
}

func ContextualizeMiddleware(ctx context.Context) func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	return Context.ContextualizeMiddleware(ctx)
}

func ContextualizeConfig(ctx context.Context, fb *config.Config) *config.Config {
	return Context.ContextualizeConfig(ctx, fb)
}

func ContextualizeNID(ctx context.Context, fallback uuid.UUID) uuid.UUID {
	return Context.ContextualizeNID(ctx, fallback)
}
