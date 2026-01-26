// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package courier_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"flag"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/ory/kratos/courier"
	templates "github.com/ory/kratos/courier/template/email"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/x"
	gomail "github.com/ory/mail/v3"
	"github.com/ory/x/configx"
)

func TestNewSMTPClientPreventLeak(t *testing.T) {
	// Test for https://hackerone.com/reports/2384028

	invalidURL := "sm<>t>p://f%oo::bar:baz@my-server:1234:122/"
	conf, reg := internal.NewFastRegistryWithMocks(t, configx.WithValue(config.ViperKeyCourierSMTPURL, invalidURL))

	channels, err := conf.CourierChannels(t.Context())
	require.NoError(t, err)
	require.Len(t, channels, 1)

	_, err = courier.NewSMTPClient(reg, channels[0].SMTPConfig)
	require.Error(t, err)
	assert.NotContains(t, err.Error(), invalidURL)
}

func TestNewSMTP(t *testing.T) {
	conf, reg := internal.NewFastRegistryWithMocks(t)

	setupSMTPClient := func(stringURL string) *courier.SMTPClient {
		conf.MustSet(t.Context(), config.ViperKeyCourierSMTPURL, stringURL)

		channels, err := conf.CourierChannels(t.Context())
		require.NoError(t, err)
		require.Len(t, channels, 1)
		c, err := courier.NewSMTPClient(reg, channels[0].SMTPConfig)
		require.NoError(t, err)
		return c
	}

	// Should enforce StartTLS => dialer.StartTLSPolicy = gomail.MandatoryStartTLS and dialer.SSL = false
	smtp := setupSMTPClient("smtp://foo:bar@my-server:1234/")
	assert.Equal(t, smtp.StartTLSPolicy, gomail.MandatoryStartTLS, "StartTLS not enforced")
	assert.Equal(t, smtp.SSL, false, "Implicit TLS should not be enabled")

	// Should enforce TLS => dialer.SSL = true
	smtp = setupSMTPClient("smtps://foo:bar@my-server:1234/")
	assert.Equal(t, smtp.SSL, true, "Implicit TLS should be enabled")

	// Should disable StartTLS completely => dialer.StartTLSPolicy = gomail.NoStartTLS and dialer.SSL = false
	smtp = setupSMTPClient("smtp://foo:bar@my-server:1234/?disable_starttls=true")
	assert.Equal(t, int(smtp.StartTLSPolicy), int(gomail.NoStartTLS), "StartTLS should be completely disabled")
	assert.Equal(t, smtp.SSL, false, "Implicit TLS should not be enabled")

	// Test cert based SMTP client auth
	clientCert, clientKey, err := generateTestClientCert(t)
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.Remove(clientCert.Name()) })
	t.Cleanup(func() { _ = os.Remove(clientKey.Name()) })

	conf.MustSet(t.Context(), config.ViperKeyCourierSMTPClientCertPath, clientCert.Name())
	conf.MustSet(t.Context(), config.ViperKeyCourierSMTPClientKeyPath, clientKey.Name())

	clientPEM, err := tls.LoadX509KeyPair(clientCert.Name(), clientKey.Name())
	require.NoError(t, err)

	smtpWithCert := setupSMTPClient("smtps://subdomain.my-server:1234/?server_name=my-server")
	assert.Equal(t, smtpWithCert.SSL, true, "Implicit TLS should be enabled")
	assert.Equal(t, smtpWithCert.Host, "subdomain.my-server", "SMTP Dialer host should match")
	assert.Equal(t, smtpWithCert.TLSConfig.ServerName, "my-server", "TLS config server name should match")
	assert.Equal(t, smtpWithCert.TLSConfig.ServerName, "my-server", "TLS config server name should match")
	assert.Contains(t, smtpWithCert.TLSConfig.Certificates, clientPEM, "TLS config should contain client pem")

	// error case: invalid client key
	require.NoError(t, conf.Set(t.Context(), config.ViperKeyCourierSMTPClientKeyPath, clientCert.Name())) // mixup client key and client cert
	smtpWithCert = setupSMTPClient("smtps://subdomain.my-server:1234/?server_name=my-server")
	assert.Equal(t, len(smtpWithCert.TLSConfig.Certificates), 0, "TLS config certificates should be empty")
}

func TestQueueEmail(t *testing.T) {
	smtp, api := x.StartMailhog(t, true)

	_, reg := internal.NewRegistryDefaultWithDSN(t, "", configx.WithValues(map[string]any{
		config.ViperKeyCourierSMTPURL:                            smtp,
		config.ViperKeyCourierSMTPFrom:                           "test-stub@ory.sh",
		config.ViperKeyCourierSMTPFromName:                       "Bob",
		config.ViperKeyCourierSMTPHeaders + ".test-stub-header1": "foo",
		config.ViperKeyCourierSMTPHeaders + ".test-stub-header2": "bar",
		config.ViperKeyCourierMessageRetries:                     50,
	}))

	c, err := reg.Courier(t.Context())
	require.NoError(t, err)
	c.FailOnDispatchError()

	_, err = c.QueueEmail(t.Context(), templates.NewTestStub(&templates.TestStubModel{
		To:      "invalid-email",
		Subject: "test-subject-1",
		Body:    "test-body-1",
	}))
	require.Error(t, err)

	id, err := c.QueueEmail(t.Context(), templates.NewTestStub(&templates.TestStubModel{
		To:      "test-recipient-1@example.org",
		Subject: "test-subject-1",
		Body:    "test-body-1",
	}))
	require.NoError(t, err)
	require.NotZero(t, id)

	id, err = c.QueueEmail(t.Context(), templates.NewTestStub(&templates.TestStubModel{
		To:      "test-recipient-2@example.org",
		Subject: "test-subject-2",
		Body:    "test-body-2",
	}))
	require.NoError(t, err)
	require.NotZero(t, id)

	id, err = c.QueueEmail(t.Context(), templates.NewTestStub(&templates.TestStubModel{
		To:      "test-recipient-3@example.org",
		Subject: "test-subject-3",
		Body:    "test-body-3",
	}))
	require.NoError(t, err)
	require.NotZero(t, id)

	require.EventuallyWithT(t, func(t *assert.CollectT) {
		require.NoError(t, c.DispatchQueue(context.Background()))
	}, time.Second, 10*time.Millisecond)

	res, err := http.Get(api + "/api/v2/messages")
	require.NoError(t, err)
	defer func() { _ = res.Body.Close() }()

	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	require.Equalf(t, http.StatusOK, res.StatusCode, "%s", body)
	require.EqualValues(t, 3, gjson.GetBytes(body, "total").Int())

	for k := 1; k <= 3; k++ {
		assert.Contains(t, string(body), fmt.Sprintf("test-subject-%d", k))
		assert.Contains(t, string(body), fmt.Sprintf("test-body-%d", k))
		assert.Contains(t, string(body), fmt.Sprintf("test-recipient-%d@example.org", k))
		assert.Contains(t, string(body), "test-stub@ory.sh")
	}

	assert.Contains(t, string(body), "Bob")
	assert.Contains(t, string(body), `"test-stub-header1":["foo"]`)
	assert.Contains(t, string(body), `"test-stub-header2":["bar"]`)
}

func generateTestClientCert(t *testing.T) (clientCert *os.File, clientKey *os.File, err error) {
	hostName := flag.String("host", "127.0.0.1", "Hostname to certify")
	priv, err := rsa.GenerateKey(rand.Reader, 1024) // #nosec G403 -- test code
	require.NoError(t, err)
	now := time.Now()
	certTemplate := x509.Certificate{
		SerialNumber: big.NewInt(1234),
		Subject: pkix.Name{
			CommonName:   *hostName,
			Organization: []string{"myorg"},
		},
		NotBefore:    now.Add(-300 * time.Second),
		NotAfter:     now.Add(24 * time.Hour),
		SubjectKeyId: []byte{1, 2, 3, 4},
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
	}
	cert, err := x509.CreateCertificate(rand.Reader, &certTemplate, &certTemplate, &priv.PublicKey, priv)
	require.NoError(t, err)
	clientCert, err = os.CreateTemp("./test", "testCert")
	require.NoError(t, err)
	defer func() { _ = clientCert.Close() }()

	require.NoError(t, pem.Encode(clientCert, &pem.Block{Type: "CERTIFICATE", Bytes: cert}))

	clientKey, err = os.CreateTemp("./test", "testKey")
	require.NoError(t, err)
	defer func() { _ = clientKey.Close() }()
	require.NoError(t, pem.Encode(clientKey, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)}))

	return clientCert, clientKey, nil
}
