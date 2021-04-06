package verification

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/x/configx"
	"github.com/ory/x/logrusx"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/x/urlx"

	"github.com/ory/kratos/selfservice/flow"
)

func TestFlow(t *testing.T) {
	conf, err := config.New(context.Background(), logrusx.New("", ""), configx.SkipValidation())
	require.NoError(t, err)

	must := func(r *Flow, err error) *Flow {
		require.NoError(t, err)
		return r
	}

	u := &http.Request{URL: urlx.ParseOrPanic("http://foo/bar/baz"), Host: "foo"}
	for k, tc := range []struct {
		r         *Flow
		expectErr bool
	}{
		{r: must(NewFlow(conf, time.Hour, "", u, nil, flow.TypeBrowser))},
		{r: must(NewFlow(conf, -time.Hour, "", u, nil, flow.TypeBrowser)), expectErr: true},
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

	assert.EqualValues(t, StateChooseMethod,
		must(NewFlow(conf, time.Hour, "", u, nil, flow.TypeBrowser)).State)
}
