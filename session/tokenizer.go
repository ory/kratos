package session

import (
	"context"
	"github.com/ory/herodot"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/x"
	"github.com/ory/x/fetcher"
	"github.com/ory/x/otelx"
	"github.com/pkg/errors"
)

type tokenizerDependencies interface {
	x.TracingProvider
	config.Provider
	Fetcher() *fetcher.Fetcher
}

type Tokenizer struct {
	r tokenizerDependencies
}

func (s *Tokenizer) TokenizeSession(ctx context.Context, template string, session *Session) (err error) {
	ctx, span := s.r.Tracer(ctx).Tracer().Start(ctx, "sessions.ManagerHTTP.TokenizeSession")
	defer otelx.End(span, &err)

	tpl, err := s.r.Config().TokenizeTemplate(ctx, template)
	if err != nil {
		return err
	}

	if tpl.Type != "jwt" {
		return errors.WithStack(herodot.ErrInternalServerError.WithReasonf("Tokenize template type %s is not supported.", tpl.Type))
	}

	return nil
}
