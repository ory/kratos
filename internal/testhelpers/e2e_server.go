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

	"github.com/stretchr/testify/assert"

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

	ctx := configx.ContextWithConfigOptions(context.Background(),
		configx.WithValue("dsn", "memory"),
		configx.WithValue("dev", true),
		configx.WithValue("log.level", "trace"),
		configx.WithValue("log.leak_sensitive_values", true),
		configx.WithValue("serve.public.port", publicPort),
		configx.WithValue("serve.admin.port", adminPort),
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

	return publicPort, adminPort
}

func StartE2EServer(t *testing.T, configFile string, configOptions ConfigOptions) (publicUrl, adminUrl string) {
	publicPort, adminPort := StartE2EServerOnly(t, configFile, configOptions)
	return CheckE2EServerOnHTTP(t, publicPort, adminPort)
}

func CheckE2EServerOnHTTP(t *testing.T, publicPort, adminPort int) (publicUrl, adminUrl string) {
	publicUrl = fmt.Sprintf("http://127.0.0.1:%d", publicPort)
	adminUrl = fmt.Sprintf("http://127.0.0.1:%d", adminPort)

	require.NoError(t, retry.Do(func() error {
		res, err := http.Get(publicUrl + "/health/alive")
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
	}))

	return
}

func CheckE2EServerOnHTTPS(t *testing.T, publicPort, adminPort int) (publicUrl, adminUrl string) {
	publicUrl = fmt.Sprintf("https://127.0.0.1:%d", publicPort)
	adminUrl = fmt.Sprintf("https://127.0.0.1:%d", adminPort)

	require.NoError(t, retry.Do(func() error {
		/* #nosec G402: TLS InsecureSkipVerify set true. */
		tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
		client := &http.Client{Transport: tr}
		res, err := client.Get(publicUrl + "/health/alive")
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
	}))

	return
}

func GenerateTLSCertificateFilesForTests(t *testing.T, certPath, keyPath string) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.Nil(t, err)

	cert, err := tlsx.CreateSelfSignedCertificate(privateKey)
	assert.Nil(t, err)

	derBytes := cert.Raw

	certOut, err := os.Create(certPath)
	if err != nil {
		t.Errorf("Failed to open cert.pem for writing: %v", err)
	}
	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		t.Errorf("Failed to write data to cert.pem: %v", err)
	}
	if err := certOut.Close(); err != nil {
		t.Errorf("Error closing cert.pem: %v", err)
	}
	t.Logf("wrote cert.pem")

	keyOut, err := os.OpenFile(keyPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		t.Errorf("Failed to open key.pem for writing: %v", err)
	}
	privBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		t.Errorf("Unable to marshal private key: %v", err)
	}
	if err := pem.Encode(keyOut, &pem.Block{Type: "PRIVATE KEY", Bytes: privBytes}); err != nil {
		t.Errorf("Failed to write data to key.pem: %v", err)
	}
	if err := keyOut.Close(); err != nil {
		t.Errorf("Error closing key.pem: %v", err)
	}
	t.Logf("wrote key.pem")
}
