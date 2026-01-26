// Copyright © 2026 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package daemon

import (
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/phayes/freeport"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"

	"github.com/ory/x/configx"

	"github.com/ory/kratos/internal"
)

func TestMetricsRouterPaths(t *testing.T) {
	// TODO refactor this test to be parallelizable once we rewrite the server setup to be properly testable
	ports, err := freeport.GetFreePorts(2)
	require.NoError(t, err)
	adminPort, publicPort := ports[0], ports[1]
	cmd := &cobra.Command{}
	cmd.Flags().Bool("sqa-opt-out", true, "")
	_, reg := internal.NewFastRegistryWithMocks(t, configx.WithValues(map[string]any{
		"serve.admin.port":  adminPort,
		"serve.public.port": publicPort,
	}))

	startAdmin, err := serveAdmin(t.Context(), reg, cmd)
	require.NoError(t, err)
	startPublic, err := servePublic(t.Context(), reg, cmd)
	require.NoError(t, err)
	eg, _ := errgroup.WithContext(t.Context())
	eg.Go(startAdmin)
	eg.Go(startPublic)

	require.EventuallyWithT(t, func(t *assert.CollectT) {
		resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/health/ready", publicPort))
		require.NoError(t, err)
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		assert.Equalf(t, http.StatusOK, resp.StatusCode, "%s", body)
	}, 2*time.Second, 10*time.Millisecond)

	// Make some requests that should be recorded in the metrics
	req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("http://127.0.0.1:%d/sessions/session-id", publicPort), nil)
	_, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	_, err = http.Get(fmt.Sprintf("http://127.0.0.1:%d/admin/identities/some-id/sessions", adminPort))
	require.NoError(t, err)

	res, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/admin/metrics/prometheus", adminPort))
	require.NoError(t, err)
	require.EqualValues(t, http.StatusOK, res.StatusCode)
	respBody, err := io.ReadAll(res.Body)
	body := string(respBody)

	require.NoError(t, err)
	assert.Contains(t, body, `endpoint="/sessions/{param}"`)
	assert.Contains(t, body, `endpoint="/admin/identities/{param}/sessions"`)
}
