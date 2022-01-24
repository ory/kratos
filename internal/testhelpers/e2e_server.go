package testhelpers

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ory/kratos/corp"
	"github.com/ory/kratos/driver"
	"github.com/ory/x/dbal"

	"golang.org/x/sync/errgroup"

	"github.com/ory/x/tlsx"

	"github.com/avast/retry-go/v3"
	"github.com/phayes/freeport"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/cmd"
	"github.com/ory/kratos/x"
	"github.com/ory/x/cmdx"
	"github.com/ory/x/configx"
)

type ConfigOptions map[string]interface{}

func init() {
	corp.SetContextualizer(new(corp.ContextNoOp))
	dbal.RegisterDriver(func() dbal.Driver {
		return driver.NewRegistryDefault()
	})
}

func StartE2EServerOnly(t *testing.T, configFile string, isTLS bool, configOptions ConfigOptions) (publicPort, adminPort int) {
	return startE2EServerOnly(t, configFile, isTLS, configOptions, 0)
}

func startE2EServerOnly(t *testing.T, configFile string, isTLS bool, configOptions ConfigOptions, tries int) (publicPort, adminPort int) {
	adminPort, err := freeport.GetFreePort()
	require.NoError(t, err)

	publicPort, err = freeport.GetFreePort()
	require.NoError(t, err)

	publicUrl := fmt.Sprintf("http://127.0.0.1:%d", publicPort)
	adminUrl := fmt.Sprintf("http://127.0.0.1:%d", adminPort)

	if isTLS {
		publicUrl = fmt.Sprintf("https://127.0.0.1:%d", publicPort)
		adminUrl = fmt.Sprintf("https://127.0.0.1:%d", adminPort)
	}

	dbt, err := os.MkdirTemp(os.TempDir(), "ory-kratos-e2e-examples-*")
	require.NoError(t, err)
	dsn := "sqlite://" + filepath.Join(dbt, "db.sqlite") + "?_fk=true&mode=rwc"

	ctx := configx.ContextWithConfigOptions(
		context.Background(),
		configx.WithValue("dsn", dsn),
		configx.WithValue("dev", true),
		configx.WithValue("log.level", "error"),
		configx.WithValue("log.leak_sensitive_values", true),
		configx.WithValue("serve.public.port", publicPort),
		configx.WithValue("serve.admin.port", adminPort),
		configx.WithValue("serve.public.base_url", publicUrl),
		configx.WithValue("serve.admin.base_url", adminUrl),
		configx.WithValues(configOptions),
	)

	//nolint:staticcheck
	ctx = context.WithValue(ctx, "dsn", dsn)
	ctx, cancel := context.WithCancel(ctx)
	executor := &cmdx.CommandExecuter{
		New: func() *cobra.Command {
			return cmd.NewRootCmd()
		},
		Ctx: ctx,
	}

	t.Log("Starting migrations...")
	_ = executor.ExecNoErr(t, "migrate", "sql", dsn, "--yes")
	t.Logf("Migration done")

	t.Log("Starting server...")
	stdOut, stdErr := &bytes.Buffer{}, &bytes.Buffer{}
	eg := executor.ExecBackground(nil, stdErr, stdOut, "serve", "--config", configFile, "--watch-courier")

	err = waitTimeout(t, eg, time.Second)
	if err != nil && tries < 5 {
		if !errors.Is(err, context.Canceled) || strings.Contains(err.Error(), "address already in use") {
			t.Logf("Detected an instance with port reuse, retrying #%d...", tries)
			time.Sleep(time.Millisecond * 500)
			cancel()
			return startE2EServerOnly(t, configFile, isTLS, configOptions, tries+1)
		}
	}
	require.NoError(t, err)

	t.Cleanup(cancel)
	return publicPort, adminPort
}

// waitTimeout waits for the waitgroup for the specified max timeout.
// Returns true if waiting timed out.
func waitTimeout(t *testing.T, wg *errgroup.Group, timeout time.Duration) (err error) {
	c := make(chan struct{})
	go func() {
		defer close(c)
		err = wg.Wait()
	}()
	select {
	case <-c:
		return
	case <-time.After(timeout):
		return
	}
}

func StartE2EServer(t *testing.T, configFile string, configOptions ConfigOptions) (publicUrl, adminUrl string) {
	publicPort, adminPort := StartE2EServerOnly(t, configFile, false, configOptions)
	return CheckE2EServerOnHTTP(t, publicPort, adminPort)
}

func CheckE2EServerOnHTTP(t *testing.T, publicPort, adminPort int) (publicUrl, adminUrl string) {
	publicUrl = fmt.Sprintf("http://127.0.0.1:%d", publicPort)
	adminUrl = fmt.Sprintf("http://127.0.0.1:%d", adminPort)
	waitToComeAlive(t, publicUrl, adminUrl)
	return
}

func waitToComeAlive(t *testing.T, publicUrl, adminUrl string) {
	require.NoError(t, retry.Do(func() error {
		/* #nosec G402: TLS InsecureSkipVerify set true. */
		tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
		client := &http.Client{Transport: tr}

		for _, url := range []string{
			publicUrl + "/health/ready",
			adminUrl + "/health/ready",
			publicUrl + "/health/alive",
			adminUrl + "/health/alive",
		} {
			res, err := client.Get(url)
			if err != nil {
				return err
			}

			body := x.MustReadAll(res.Body)
			if err := res.Body.Close(); err != nil {
				return err
			}
			t.Logf("%s", body)

			if res.StatusCode != http.StatusOK {
				return fmt.Errorf("expected status code 200 but got: %d", res.StatusCode)
			}
		}
		return nil
	},
		retry.MaxDelay(time.Second),
		retry.Attempts(60)),
	)
}

func CheckE2EServerOnHTTPS(t *testing.T, publicPort, adminPort int) (publicUrl, adminUrl string) {
	publicUrl = fmt.Sprintf("https://127.0.0.1:%d", publicPort)
	adminUrl = fmt.Sprintf("https://127.0.0.1:%d", adminPort)
	waitToComeAlive(t, publicUrl, adminUrl)
	return
}

func GenerateTLSCertificateFilesForTests(t *testing.T, certPath, keyPath string) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	cert, err := tlsx.CreateSelfSignedCertificate(privateKey)
	require.NoError(t, err)

	certOut, err := os.Create(certPath)
	require.NoError(t, err, "Failed to open cert.pem for writing: %v", err)

	err = pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: cert.Raw})
	require.NoError(t, err, "Failed to write data to cert.pem: %v", err)

	err = certOut.Close()
	require.NoError(t, err, "Error closing cert.pem: %v", err)

	t.Logf("wrote cert.pem")

	keyOut, err := os.OpenFile(keyPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	require.NoError(t, err, "Failed to open key.pem for writing: %v", err)

	privBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	require.NoError(t, err, "Unable to marshal private key: %v", err)

	err = pem.Encode(keyOut, &pem.Block{Type: "PRIVATE KEY", Bytes: privBytes})
	require.NoError(t, err, "Failed to write data to key.pem: %v", err)

	err = keyOut.Close()
	require.NoError(t, err, "Error closing key.pem: %v", err)

	t.Logf("wrote key.pem")
}
