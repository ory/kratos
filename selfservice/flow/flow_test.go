package flow

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/herodot"
	"github.com/ory/kratos/x"
)

func TestGetFlowID(t *testing.T) {
	_, err := GetFlowID(&http.Request{URL: &url.URL{RawQuery: ""}})
	require.Error(t, err)
	assert.Contains(t,
		errors.Cause(err).(*herodot.DefaultError).ReasonField,
		"flow query parameter is missing or malformed")

	_, err = GetFlowID(&http.Request{URL: &url.URL{RawQuery: "flow=not-uuid"}})
	require.Error(t, err)
	assert.Contains(t,
		errors.Cause(err).(*herodot.DefaultError).ReasonField,
		"flow query parameter is missing or malformed")

	expected := x.NewUUID()
	actual, err := GetFlowID(&http.Request{URL: &url.URL{RawQuery: "flow=" + expected.String()}})
	require.NoError(t, err)
	assert.Equal(t, expected.String(), actual.String())
}
