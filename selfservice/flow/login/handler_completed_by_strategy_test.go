// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package login_test

import (
	"context"
	"io"
	"net/http"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"

	"github.com/ory/x/httprouterx"
	"github.com/ory/x/otelx"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/hydra"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/pkg"
	"github.com/ory/kratos/pkg/testhelpers"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/selfservice/strategy/code"
	"github.com/ory/kratos/session"
)

// completeByStrategyCodeStrategy embeds the real code strategy but forces its
// fast-login path to report that it has completed the flow. Completion is a
// control signal (the strategy has written the response itself), modelled as
// the `completed` return value rather than an error.
type completeByStrategyCodeStrategy struct {
	*code.Strategy
}

func (s *completeByStrategyCodeStrategy) FastLogin2FA(_ http.ResponseWriter, _ *http.Request, _ *login.Flow, _ *session.Session) (bool, error) {
	return true, nil
}

// TestCreateLoginFlowCompletedByStrategyTelemetry is a regression test for
// https://github.com/ory/kratos/issues/4109: when a strategy completes the flow
// during creation, the handler intentionally does not write an error to the
// client. Completion is now returned as a control value (completed bool) rather
// than a sentinel error, so the deferred otelx.End(span, &err) can no longer
// record it as a span error/exception. This test asserts that neither the
// client nor the span sees an error on the completed-by-strategy path.
func TestCreateLoginFlowCompletedByStrategyTelemetry(t *testing.T) {
	ctx := context.Background()
	conf, reg := pkg.NewFastRegistryWithMocks(t)
	reg.SetHydra(hydra.NewFake())

	// Install a span recorder as the registry tracer so we can inspect the span
	// that the handler produces.
	recorder := tracetest.NewSpanRecorder()
	tracer := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder)).Tracer("test")
	reg.SetTracer(otelx.NewNoop().WithOTLP(tracer))

	reg.WithSelfserviceStrategies(t, []any{&completeByStrategyCodeStrategy{code.NewStrategy(reg)}})
	conf.MustSet(ctx, config.ViperKeySelfServiceStrategyConfig+"."+string(identity.CredentialsTypeCodeAuth),
		map[string]interface{}{"passwordless_enabled": true})
	testhelpers.SetDefaultIdentitySchema(conf, "file://./stub/login.schema.json")

	routerPublic := httprouterx.NewRouterPublic()
	ts, _ := testhelpers.NewKratosServerWithRouters(t, reg, routerPublic, httprouterx.NewRouterAdminWithPrefix())

	req := testhelpers.NewTestHTTPRequest(t, "GET", ts.URL+login.RouteInitAPIFlow, nil)
	res, err := ts.Client().Do(req)
	require.NoError(t, err)
	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	require.NoError(t, res.Body.Close())

	// The client must not receive an error: the flow was completed by the
	// strategy, which is a control signal and not a failure.
	assert.Empty(t, gjson.GetBytes(body, "error").Raw, "the completed-by-strategy signal must not be surfaced to the client, body: %s", body)

	// The span for the handler must not be marked as errored, because the
	// control signal is expected behaviour, not a failure.
	ended := recorder.Ended()
	i := slices.IndexFunc(ended, func(sp sdktrace.ReadOnlySpan) bool {
		return sp.Name() == "selfservice.flow.login.createNativeLoginFlow"
	})
	require.GreaterOrEqual(t, i, 0, "expected the createNativeLoginFlow span to be recorded, got %v", spanNames(ended))

	handlerSpan := ended[i]
	assert.NotEqual(t, codes.Error, handlerSpan.Status().Code,
		"the createNativeLoginFlow span must not be marked as errored when a strategy completes the flow, but was: %q", handlerSpan.Status().Description)

	for _, e := range handlerSpan.Events() {
		assert.NotEqual(t, "exception", e.Name, "the createNativeLoginFlow span must not record an exception event when a strategy completes the flow")
	}
}

func spanNames(spans []sdktrace.ReadOnlySpan) []string {
	names := make([]string, len(spans))
	for i, sp := range spans {
		names[i] = sp.Name()
	}
	return names
}
