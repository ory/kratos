package tls_tests

import (
	"encoding/base64"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"testing"

	"github.com/ory/kratos/internal/testhelpers"
)


func TestServeTLSBase64(t *testing.T) {
	certPath := "/tmp/e2e_test_cert.pem"
	keyPath := "/tmp/e2e_test_key.pem"

	testhelpers.GenerateTLSCertificateFilesForTests(t, certPath, keyPath)

	certRaw, err := ioutil.ReadFile(certPath)
	require.NoError(t, err)

	keyRaw, err := ioutil.ReadFile(keyPath)
	require.NoError(t, err)

	certBase64 := base64.StdEncoding.EncodeToString(certRaw)
	keyBase64 := base64.StdEncoding.EncodeToString(keyRaw)
	publicPort, adminPort := testhelpers.StartE2EServerOnly(t,
		"../../../contrib/quickstart/kratos/email-password/kratos.yml",
		testhelpers.ConfigOptions{
			"serve.public.tls.key.base64":  keyBase64,
			"serve.public.tls.cert.base64": certBase64,
			"serve.admin.tls.key.base64":  keyBase64,
			"serve.admin.tls.cert.base64": certBase64,
		},
	)
	testhelpers.CheckE2EServerOnHTTPS(t, publicPort, adminPort)
}
