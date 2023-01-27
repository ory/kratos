// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

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

// These global functions call the respective method on Context

func ContextualizeConfig(ctx context.Context, fb *config.Config) *config.Config {
	return c.ContextualizeConfig(ctx, fb)
}
