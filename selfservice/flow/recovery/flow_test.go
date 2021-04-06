package recovery_test

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/ory/kratos/internal"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/x/urlx"

	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/recovery"
)

func TestFlow(t *testing.T) {
	conf := internal.NewConfigurationWithDefaults(t)
	must := func(r *recovery.Flow, err error) *recovery.Flow {
		require.NoError(t, err)
		return r
	}

	u := &http.Request{URL: urlx.ParseOrPanic("http://foo/bar/baz"), Host: "foo"}
	for k, tc := range []struct {
		r         *recovery.Flow
		expectErr bool
	}{
		{r: must(recovery.NewFlow(conf, time.Hour, "", u, nil, flow.TypeBrowser))},
		{r: must(recovery.NewFlow(conf, -time.Hour, "", u, nil, flow.TypeBrowser)), expectErr: true},
	} {
		t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
			err := tc.r.Valid()
			if tc.expectErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}

	assert.EqualValues(t, recovery.StateChooseMethod,
		must(recovery.NewFlow(conf, time.Hour, "", u, nil, flow.TypeBrowser)).State)
}
