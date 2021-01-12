package corp

import (
	"context"
	"net/http"

	"github.com/ory/kratos/driver/config"
)

func ContextualizeTableName(_ context.Context, name string) string {
	return name
}

func ContextualizeMiddleware(_ context.Context) func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	return func(w http.ResponseWriter, r *http.Request, n http.HandlerFunc) {
		n(w, r)
	}
}

func ContextualizeConfig(ctx context.Context, fb *config.Config) *config.Config {
	return fb
}
