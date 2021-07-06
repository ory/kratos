package serve_test

import (
	"encoding/base64"
	"github.com/ory/kratos/x"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/ory/kratos/internal/testhelpers"
)

func TestServe(t *testing.T) {
	_, _ = testhelpers.StartE2EServer(t, "../../contrib/quickstart/kratos/email-password/kratos.yml", nil)
}

func TestServeTLSBase64(t *testing.T) {
	certPath := filepath.Join(os.TempDir(), "e2e_test_cert_"+x.NewUUID().String()+".pem")
	keyPath := filepath.Join(os.TempDir(), "e2e_test_key_"+x.NewUUID().String()+".pem")

	testhelpers.GenerateTLSCertificateFilesForTests(t, certPath, keyPath)

	certRaw, err := ioutil.ReadFile(certPath)
	require.NoError(t, err)

	keyRaw, err := ioutil.ReadFile(keyPath)
	require.NoError(t, err)

	certBase64 := base64.StdEncoding.EncodeToString(certRaw)
	keyBase64 := base64.StdEncoding.EncodeToString(keyRaw)
	publicPort, adminPort := testhelpers.StartE2EServerOnly(t,
		"../../contrib/quickstart/kratos/email-password/kratos.yml",
		testhelpers.ConfigOptions{
			"serve.public.tls.key.base64":  keyBase64,
			"serve.public.tls.cert.base64": certBase64,
			"serve.admin.tls.key.base64":   keyBase64,
			"serve.admin.tls.cert.base64":  certBase64,
		},
	)
	testhelpers.CheckE2EServerOnHTTPS(t, publicPort, adminPort)
}

func TestServeTLSPaths(t *testing.T) {
	certPath := filepath.Join(os.TempDir(), "e2e_test_cert_"+x.NewUUID().String()+".pem")
	keyPath := filepath.Join(os.TempDir(), "e2e_test_key_"+x.NewUUID().String()+".pem")

	testhelpers.GenerateTLSCertificateFilesForTests(t, certPath, keyPath)

	publicPort, adminPort := testhelpers.StartE2EServerOnly(t,
		"../../contrib/quickstart/kratos/email-password/kratos.yml",
		testhelpers.ConfigOptions{
			"serve.public.tls.key.path":  keyPath,
			"serve.public.tls.cert.path": certPath,
			"serve.admin.tls.key.path":   keyPath,
			"serve.admin.tls.cert.path":  certPath,
		},
	)
	testhelpers.CheckE2EServerOnHTTPS(t, publicPort, adminPort)
}
