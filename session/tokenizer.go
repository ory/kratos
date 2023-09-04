// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package session

import (
	"context"
	"encoding/json"
	"time"

	"github.com/gofrs/uuid"
	"github.com/golang-jwt/jwt/v5"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"

	"github.com/ory/herodot"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/x"
	"github.com/ory/x/fetcher"
	"github.com/ory/x/jsonnetsecure"
	"github.com/ory/x/jwksx"
	"github.com/ory/x/otelx"
)

type (
	tokenizerDependencies interface {
		jsonnetsecure.VMProvider
		x.TracingProvider
		x.HTTPClientProvider
		config.Provider
		x.JWKFetchProvider
	}
	Tokenizer struct {
		r       tokenizerDependencies
		nowFunc func() time.Time
	}
	TokenizerProvider interface {
		SessionTokenizer() *Tokenizer
	}
)

func NewTokenizer(r tokenizerDependencies) *Tokenizer {
	return &Tokenizer{r: r, nowFunc: time.Now().UTC}
}

func (s *Tokenizer) SetNowFunc(t func() time.Time) {
	s.nowFunc = t
}

func (s *Tokenizer) TokenizeSession(ctx context.Context, template string, session *Session) (err error) {
	ctx, span := s.r.Tracer(ctx).Tracer().Start(ctx, "sessions.ManagerHTTP.TokenizeSession")
	defer otelx.End(span, &err)

	tpl, err := s.r.Config().TokenizeTemplate(ctx, template)
	if err != nil {
		return err
	}

	httpClient := s.r.HTTPClient(ctx)
	if tpl.Type != "jwt" {
		return errors.WithStack(herodot.ErrInternalServerError.WithReasonf("Tokenize template type \"%s\" is not supported.", tpl.Type))
	}

	key, err := s.r.Fetcher().ResolveKey(
		ctx,
		tpl.Config.JWKSURL,
		jwksx.WithCacheEnabled(),
		jwksx.WithCacheTTL(time.Hour),
		jwksx.WithHTTPClient(httpClient))
	if err != nil {
		if errors.Is(err, jwksx.ErrUnableToFindKeyID) {
			return errors.WithStack(herodot.ErrBadRequest.WithReasonf("Could not find key a suitable key for tokenization in the JWKS url."))
		}
		return err
	}

	alg := jwt.GetSigningMethod(key.Algorithm())
	if alg == nil {
		return errors.WithStack(herodot.ErrBadRequest.WithReasonf("The JSON Web Key must include a valid \"alg\" parameter but \"%s\" was given.", key.Algorithm()))
	}

	vm, err := s.r.JsonnetVM(ctx)
	if err != nil {
		return err
	}

	fetch := fetcher.NewFetcher(fetcher.WithClient(httpClient))

	now := s.nowFunc()
	token := jwt.New(alg)
	token.Header["kid"] = key.KeyID()
	claims := jwt.MapClaims{
		"jti": uuid.Must(uuid.NewV4()).String(),
		"iss": s.r.Config().SelfPublicURL(ctx).String(),
		"exp": now.Add(tpl.Config.TTL).Unix(),
		"sub": session.IdentityID.String(),
		"sid": session.ID.String(),
		"nbf": now.Unix(),
		"iat": now.Unix(),
	}

	if mapper := tpl.Config.ClaimsMapperURL; len(mapper) > 0 {
		jn, err := fetch.FetchContext(ctx, mapper)
		if err != nil {
			return err
		}

		sessionRaw, err := json.Marshal(session)
		if err != nil {
			return errors.WithStack(herodot.ErrInternalServerError.WithWrap(err).WithReasonf("Unable to encode session to JSON."))
		}

		claimsRaw, err := json.Marshal(&claims)
		if err != nil {
			return errors.WithStack(herodot.ErrInternalServerError.WithWrap(err).WithReasonf("Unable to encode session to JSON."))
		}

		vm.ExtCode("session", string(sessionRaw))
		vm.ExtCode("claims", string(claimsRaw))

		evaluated, err := vm.EvaluateAnonymousSnippet(tpl.Config.ClaimsMapperURL, jn.String())
		if err != nil {
			return errors.WithStack(herodot.ErrBadRequest.WithWrap(err).WithDebug(err.Error()).WithReasonf("Unable to execute tokenizer JsonNet."))
		}

		evaluatedClaims := gjson.Get(evaluated, "claims")
		if !evaluatedClaims.IsObject() {
			return errors.WithStack(herodot.ErrBadRequest.WithWrap(err).WithReasonf("Expected tokenizer JsonNet to return a claims object but it did not."))
		}

		if err := json.Unmarshal([]byte(evaluatedClaims.Raw), &claims); err != nil {
			return errors.WithStack(herodot.ErrBadRequest.WithWrap(err).WithReasonf("Unable to encode tokenized claims."))
		}

		claims["sub"] = session.IdentityID.String()
	}

	var privateKey interface{}
	if err := key.Raw(&privateKey); err != nil {
		return errors.WithStack(herodot.ErrBadRequest.WithWrap(err).WithReasonf("Unable to decode the given private key."))
	}

	token.Claims = claims
	result, err := token.SignedString(privateKey)
	if err != nil {
		return errors.WithStack(herodot.ErrBadRequest.WithWrap(err).WithReasonf("Unable to sign JSON Web Token."))
	}

	session.Tokenized = result
	return nil
}
