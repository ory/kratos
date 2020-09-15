package password

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTidyForm(t *testing.T) {
	assert.EqualValues(t, url.Values{"foobar": {"foo"}}, tidyForm(url.Values{
		"password":   {"some-value"},
		"csrf_token": {"some-value"},
		"flow":       {"some-value"},
		"foobar":     {"foo"},
	}))
}
