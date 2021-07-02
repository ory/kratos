package tls_tests

import (
	"testing"

	"github.com/ory/kratos/internal/testhelpers"
)


func TestServeTLSPaths(t *testing.T) {
	certPath := "/tmp/e2e_test_cert.pem"
	keyPath := "/tmp/e2e_test_key.pem"

	testhelpers.GenerateTLSCertificateFilesForTests(t, certPath, keyPath)

	publicPort, adminPort := testhelpers.StartE2EServerOnly(t,
		"../../../contrib/quickstart/kratos/email-password/kratos.yml",
		testhelpers.ConfigOptions{
			"serve.public.tls.key.path":  keyPath,
			"serve.public.tls.cert.path": certPath,
			"serve.admin.tls.key.path":  keyPath,
			"serve.admin.tls.cert.path": certPath,
		},
	)
	testhelpers.CheckE2EServerOnHTTPS(t, publicPort, adminPort)
}
