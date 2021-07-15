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

var c Contextualizer = nil

func GetContextualizer() Contextualizer {
	return c
}

func SetContextualizer(cc Contextualizer) {
	if _, ok := cc.(*ContextNoOp); ok && c != nil {
		return
	}

	c = cc
}

// These global functions call the respective method on Context

func ContextualizeTableName(ctx context.Context, name string) string {
	return c.ContextualizeTableName(ctx, name)
}

func ContextualizeMiddleware(ctx context.Context) func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	return c.ContextualizeMiddleware(ctx)
}

func ContextualizeConfig(ctx context.Context, fb *config.Config) *config.Config {
	return c.ContextualizeConfig(ctx, fb)
}

func ContextualizeNID(ctx context.Context, fallback uuid.UUID) uuid.UUID {
	return c.ContextualizeNID(ctx, fallback)
}
