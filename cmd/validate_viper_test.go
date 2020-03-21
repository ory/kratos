package cmd

import (
	"os"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/viper"
)

func setRequiredKeys(t *testing.T) {
	require.NoError(t, os.Setenv("DSN", "memory"))
	require.NoError(t, os.Setenv("IDENTITY_TRAITS_DEFAULT_SCHEMA_URL", "file://./stub.schema.json"))
	require.NoError(t, os.Setenv("SELFSERVICE_LOGOUT_REDIRECT_TO", "https://example.com"))
	require.NoError(t, os.Setenv("COURIER_SMTP_CONNECTION_URI", "https://example.com"))
	require.NoError(t, os.Setenv("URLS_PROFILE_UI", "https://example.com"))
	require.NoError(t, os.Setenv("URLS_MFA_UI", "https://example.com"))
	require.NoError(t, os.Setenv("URLS_LOGIN_UI", "https://example.com"))
	require.NoError(t, os.Setenv("URLS_REGISTRATION_UI", "https://example.com"))
	require.NoError(t, os.Setenv("URLS_ERROR_UI", "https://example.com"))
	require.NoError(t, os.Setenv("URLS_VERIFY_UI", "https://example.com"))
	require.NoError(t, os.Setenv("URLS_DEFAULT_RETURN_TO", "https://example.com"))
}

func TestWatchAndValidateViper(t *testing.T) {
	setRequiredKeys(t)
	log := logrus.New()
	logger = log
	log.ExitFunc = func(i int) {
		t.Errorf("unexpectedly exited with code %d", i)
	}

	t.Run("case=reads string slice from env var", func(t *testing.T) {
		hook := test.NewLocal(log)
		require.NoError(t, os.Setenv("URLS_WHITELISTED_RETURN_TO_DOMAINS", "https://expample.com/one https://expample.com/two"))
		watchAndValidateViper()
		assert.Equal(t, []*logrus.Entry{}, hook.AllEntries())
		assert.Equal(t, []string{"https://expample.com/one", "https://expample.com/two"}, viper.Get("urls.whitelisted_return_to_domains"))
	})
}
