// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package session

import (
	"context"
	"encoding/json"
	"time"

	"github.com/dgraph-io/ristretto/v2"
	"github.com/gofrs/uuid"
	"github.com/golang-jwt/jwt/v5"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
	"go.opentelemetry.io/otel/trace"

	"github.com/ory/herodot"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/x"
	"github.com/ory/kratos/x/events"
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
		x.JWKSFetchProvider
	}
	Tokenizer struct {
		r       tokenizerDependencies
		nowFunc func() time.Time
		cache   *ristretto.Cache[[]byte, []byte]
	}
	TokenizerProvider interface {
		SessionTokenizer() *Tokenizer
	}
)

func NewTokenizer(r tokenizerDependencies) *Tokenizer {
	cache, _ := ristretto.NewCache(&ristretto.Config[[]byte, []byte]{
		MaxCost:     50 << 20, // 50MB,
		NumCounters: 500_000,  // 1kB per snippet -> 50k snippets -> 500k counters
		BufferItems: 64,
	})
	return &Tokenizer{r: r, nowFunc: time.Now, cache: cache}
}

func (s *Tokenizer) SetNowFunc(t func() time.Time) {
	s.nowFunc = t
}

func SetSubjectClaim(claims jwt.MapClaims, session *Session, subjectSource string) error {
	switch subjectSource {
	case "", "id":
		claims["sub"] = session.IdentityID.String()
	case "external_id":
		if session.Identity.ExternalID == "" {
			return errors.WithStack(herodot.ErrBadRequest.WithReasonf("The session's identity does not have an external ID set, but it is required for the subject claim."))
		}
		claims["sub"] = session.Identity.ExternalID.String()
	default:
		return errors.WithStack(herodot.ErrBadRequest.WithReasonf("Unknown subject source %q", subjectSource))
	}
	return nil
}

func (s *Tokenizer) TokenizeSession(ctx context.Context, template string, session *Session) (err error) {
	ctx, span := s.r.Tracer(ctx).Tracer().Start(ctx, "sessions.ManagerHTTP.TokenizeSession")
	defer otelx.End(span, &err)

	tpl, err := s.r.Config().TokenizeTemplate(ctx, template)
	if err != nil {
		return err
	}

	httpClient := s.r.HTTPClient(ctx)
	key, err := s.r.JWKSFetcher().ResolveKey(
		ctx,
		tpl.JWKSURL,
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

	now := s.nowFunc()
	token := jwt.New(alg)
	token.Header["kid"] = key.KeyID()
	claims := jwt.MapClaims{
		"jti": uuid.Must(uuid.NewV4()).String(),
		"iss": s.r.Config().SelfPublicURL(ctx).String(),
		"exp": now.Add(tpl.TTL).Unix(),
		"sid": session.ID.String(),
		"nbf": now.Unix(),
		"iat": now.Unix(),
	}

	if err = SetSubjectClaim(claims, session, tpl.SubjectSource); err != nil {
		return err
	}

	if mapper := tpl.ClaimsMapperURL; len(mapper) > 0 {
		sessionRaw, err := json.Marshal(session)
		if err != nil {
			return errors.WithStack(herodot.ErrInternalServerError.WithWrap(err).WithReasonf("Unable to encode session to JSON."))
		}

		claimsRaw, err := json.Marshal(&claims)
		if err != nil {
			return errors.WithStack(herodot.ErrInternalServerError.WithWrap(err).WithReasonf("Unable to encode claims to JSON."))
		}

		vm.ExtCode("session", string(sessionRaw))
		vm.ExtCode("claims", string(claimsRaw))

		fetcher := fetcher.NewFetcher(fetcher.WithClient(httpClient), fetcher.WithCache(s.cache, 60*time.Minute))
		jsonnet, err := fetcher.FetchContext(ctx, mapper)
		if err != nil {
			return err
		}
		evaluated, err := vm.EvaluateAnonymousSnippet(tpl.ClaimsMapperURL, jsonnet.String())
		if err != nil {
			trace.SpanFromContext(ctx).AddEvent(events.NewJsonnetMappingFailed(
				ctx, err, jsonnet.Bytes(), evaluated, "", "",
			))
			return errors.WithStack(herodot.ErrBadRequest.WithWrap(err).WithDebug(err.Error()).WithReasonf("Unable to execute tokenizer JsonNet."))
		}

		evaluatedClaims := gjson.Get(evaluated, "claims")
		if !evaluatedClaims.IsObject() {
			trace.SpanFromContext(ctx).AddEvent(events.NewJsonnetMappingFailed(
				ctx, err, jsonnet.Bytes(), evaluated, "", "",
			))
			return errors.WithStack(herodot.ErrBadRequest.WithWrap(err).WithReasonf("Expected tokenizer JsonNet to return a claims object but it did not."))
		}

		if err := json.Unmarshal([]byte(evaluatedClaims.Raw), &claims); err != nil {
			return errors.WithStack(herodot.ErrBadRequest.WithWrap(err).WithReasonf("Unable to encode tokenized claims."))
		}
	}
	if err = SetSubjectClaim(claims, session, tpl.SubjectSource); err != nil {
		return err
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

	trace.SpanFromContext(ctx).AddEvent(events.NewSessionJWTIssued(ctx, session.ID, session.IdentityID, tpl.TTL))
	session.Tokenized = result
	return nil
}
