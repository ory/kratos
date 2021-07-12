package testhelpers

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net/http"
	"os"
	"testing"

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

func StartE2EServerOnly(t *testing.T, configFile string, configOptions ConfigOptions) (publicPort, adminPort int) {
	adminPort, err := freeport.GetFreePort()
	require.NoError(t, err)

	publicPort, err = freeport.GetFreePort()
	require.NoError(t, err)

	publicUrl = fmt.Sprintf("http://127.0.0.1:%d", publicPort)
	adminUrl = fmt.Sprintf("http://127.0.0.1:%d", adminPort)

	ctx := configx.ContextWithConfigOptions(context.Background(),
		configx.WithValue("dsn", "memory"),
		configx.WithValue("dev", true),
		configx.WithValue("log.level", "trace"),
		configx.WithValue("log.leak_sensitive_values", true),
		configx.WithValue("serve.public.port", publicPort),
		configx.WithValue("serve.admin.port", adminPort),
		configx.WithValue("serve.public.base_url", publicUrl),
		configx.WithValue("serve.admin.base_url", adminUrl),
		configx.WithValues(configOptions),
	)

	//nolint:staticcheck
	ctx = context.WithValue(ctx, "dsn", "memory")
	ctx, cancel := context.WithCancel(ctx)
	t.Cleanup(cancel)

	executor := &cmdx.CommandExecuter{New: func() *cobra.Command {
		return cmd.NewRootCmd()
	}, Ctx: ctx}

	go func() {
		t.Log("Starting server...")
		_ = executor.ExecNoErr(t, "serve", "--config", configFile, "--watch-courier")
	}()

	require.NoError(t, retry.Do(func() error {
		res, err := http.Get(publicUrl + "/health/ready")
		if err != nil {
			return err
		}
		defer res.Body.Close()
		body := x.MustReadAll(res.Body)
		if res.StatusCode != http.StatusOK {
			t.Logf("%s", body)
			return fmt.Errorf("expected status code 200 but got: %d", res.StatusCode)
		}
		return nil
	}), err)


	return publicPort, adminPort
}

func StartE2EServer(t *testing.T, configFile string, configOptions ConfigOptions) (publicUrl, adminUrl string) {
	publicPort, adminPort := StartE2EServerOnly(t, configFile, configOptions)
	return CheckE2EServerOnHTTP(t, publicPort, adminPort)
}

func CheckE2EServerOnHTTP(t *testing.T, publicPort, adminPort int) (publicUrl, adminUrl string) {
	publicUrl = fmt.Sprintf("http://127.0.0.1:%d", publicPort)
	adminUrl = fmt.Sprintf("http://127.0.0.1:%d", adminPort)
}

func CheckE2EServerOnHTTPS(t *testing.T, publicPort, adminPort int) (publicUrl, adminUrl string) {
	publicUrl = fmt.Sprintf("https://127.0.0.1:%d", publicPort)
	adminUrl = fmt.Sprintf("https://127.0.0.1:%d", adminPort)

	require.NoError(t, retry.Do(func() error {
		/* #nosec G402: TLS InsecureSkipVerify set true. */
		tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
		client := &http.Client{Transport: tr}

		for _, url := range []string{publicUrl + "/health/ready", adminUrl + "/health/ready"} {
			res, err := client.Get(url)
			if err != nil {
				return err
			}

			body := x.MustReadAll(res.Body)

			err = res.Body.Close()
			if err != nil {
				return err
			}

			if res.StatusCode != http.StatusOK {
				t.Logf("%s", body)
				return fmt.Errorf("expected status code 200 but got: %d", res.StatusCode)
			}
		}
		return nil
	}))

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
