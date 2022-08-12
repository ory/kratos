package courier_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
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

	"crypto/x509"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/ory/kratos/courier"
	templates "github.com/ory/kratos/courier/template/email"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/x"
	gomail "github.com/ory/mail/v3"
)

func TestNewSMTP(t *testing.T) {
	ctx := context.Background()
	conf, reg := internal.NewFastRegistryWithMocks(t)

	setupCourier := func(stringURL string) courier.Courier {
		conf.MustSet(config.ViperKeyCourierSMTPURL, stringURL)
		t.Logf("SMTP URL: %s", conf.CourierSMTPURL().String())

		return courier.NewCourier(ctx, reg)
	}

	if testing.Short() {
		t.SkipNow()
	}

	//Should enforce StartTLS => dialer.StartTLSPolicy = gomail.MandatoryStartTLS and dialer.SSL = false
	smtp := setupCourier("smtp://foo:bar@my-server:1234/")
	assert.Equal(t, smtp.SmtpDialer().StartTLSPolicy, gomail.MandatoryStartTLS, "StartTLS not enforced")
	assert.Equal(t, smtp.SmtpDialer().SSL, false, "Implicit TLS should not be enabled")

	//Should enforce TLS => dialer.SSL = true
	smtp = setupCourier("smtps://foo:bar@my-server:1234/")
	assert.Equal(t, smtp.SmtpDialer().SSL, true, "Implicit TLS should be enabled")

	//Should allow cleartext => dialer.StartTLSPolicy = gomail.OpportunisticStartTLS and dialer.SSL = false
	smtp = setupCourier("smtp://foo:bar@my-server:1234/?disable_starttls=true")
	assert.Equal(t, smtp.SmtpDialer().StartTLSPolicy, gomail.OpportunisticStartTLS, "StartTLS is enforced")
	assert.Equal(t, smtp.SmtpDialer().SSL, false, "Implicit TLS should not be enabled")

	//Test cert based SMTP client auth
	clientCert, clientKey, err := generateTestClientCert()
	require.NoError(t, err)
	defer os.Remove(clientCert.Name())
	defer os.Remove(clientKey.Name())

	conf.Set(config.ViperKeyCourierSMTPClientCertPath, clientCert.Name())
	conf.Set(config.ViperKeyCourierSMTPClientKeyPath, clientKey.Name())

	clientPEM, err := tls.LoadX509KeyPair(clientCert.Name(), clientKey.Name())
	require.NoError(t, err)

	smtpWithCert := setupCourier("smtps://subdomain.my-server:1234/?server_name=my-server")
	assert.Equal(t, smtpWithCert.SmtpDialer().SSL, true, "Implicit TLS should be enabled")
	assert.Equal(t, smtpWithCert.SmtpDialer().Host, "subdomain.my-server", "SMTP Dialer host should match")
	assert.Equal(t, smtpWithCert.SmtpDialer().TLSConfig.ServerName, "my-server", "TLS config server name should match")
	assert.Equal(t, smtpWithCert.SmtpDialer().TLSConfig.ServerName, "my-server", "TLS config server name should match")
	assert.Contains(t, smtpWithCert.SmtpDialer().TLSConfig.Certificates, clientPEM, "TLS config should contain client pem")

	//error case: invalid client key
	conf.Set(config.ViperKeyCourierSMTPClientKeyPath, clientCert.Name()) //mixup client key and client cert
	smtpWithCert = setupCourier("smtps://subdomain.my-server:1234/?server_name=my-server")
	assert.Equal(t, len(smtpWithCert.SmtpDialer().TLSConfig.Certificates), 0, "TLS config certificates should be empty")
}

func TestQueueEmail(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	smtp, api, err := x.RunTestSMTP()
	require.NoError(t, err)
	t.Logf("SMTP URL: %s", smtp)
	t.Logf("API URL: %s", api)

	ctx := context.Background()

	conf, reg := internal.NewRegistryDefaultWithDSN(t, "")
	conf.MustSet(config.ViperKeyCourierSMTPURL, smtp)
	conf.MustSet(config.ViperKeyCourierSMTPFrom, "test-stub@ory.sh")
	reg.Logger().Level = logrus.TraceLevel

	c := reg.Courier(ctx)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	id, err := c.QueueEmail(ctx, templates.NewTestStub(reg, &templates.TestStubModel{
		To:      "test-recipient-1@example.org",
		Subject: "test-subject-1",
		Body:    "test-body-1",
	}))
	require.NoError(t, err)
	require.NotEqual(t, uuid.Nil, id)

	id, err = c.QueueEmail(ctx, templates.NewTestStub(reg, &templates.TestStubModel{
		To:      "test-recipient-2@example.org",
		Subject: "test-subject-2",
		Body:    "test-body-2",
	}))
	require.NoError(t, err)
	require.NotEqual(t, uuid.Nil, id)

	// The third email contains a sender name and custom headers
	conf.MustSet(config.ViperKeyCourierSMTPFromName, "Bob")
	conf.MustSet(config.ViperKeyCourierSMTPHeaders+".test-stub-header1", "foo")
	conf.MustSet(config.ViperKeyCourierSMTPHeaders+".test-stub-header2", "bar")
	customerHeaders := conf.CourierSMTPHeaders()
	require.Len(t, customerHeaders, 2)
	id, err = c.QueueEmail(ctx, templates.NewTestStub(reg, &templates.TestStubModel{
		To:      "test-recipient-3@example.org",
		Subject: "test-subject-3",
		Body:    "test-body-3",
	}))
	require.NoError(t, err)
	require.NotEqual(t, uuid.Nil, id)

	go func() {
		require.NoError(t, c.Work(ctx))
	}()

	var body []byte
	for k := 0; k < 30; k++ {
		time.Sleep(time.Second)
		err = func() error {
			res, err := http.Get(api + "/api/v2/messages")
			if err != nil {
				return err
			}

			defer res.Body.Close()
			body, err = io.ReadAll(res.Body)
			if err != nil {
				return err
			}

			if http.StatusOK != res.StatusCode {
				return errors.Errorf("expected status code 200 but got %d with body: %s", res.StatusCode, body)
			}

			if total := gjson.GetBytes(body, "total").Int(); total != 3 {
				return errors.Errorf("expected to have delivered at least 3 messages but got count %d with body: %s", total, body)
			}

			return nil
		}()
		if err == nil {
			break
		}
	}
	require.NoError(t, err)

	for k := 1; k <= 3; k++ {
		assert.Contains(t, string(body), fmt.Sprintf("test-subject-%d", k))
		assert.Contains(t, string(body), fmt.Sprintf("test-body-%d", k))
		assert.Contains(t, string(body), fmt.Sprintf("test-recipient-%d@example.org", k))
		assert.Contains(t, string(body), "test-stub@ory.sh")
	}

	// Assertion for the third email with sender name and headers
	assert.Contains(t, string(body), "Bob")
	assert.Contains(t, string(body), `"test-stub-header1":["foo"]`)
	assert.Contains(t, string(body), `"test-stub-header2":["bar"]`)
}

func generateTestClientCert() (clientCert *os.File, clientKey *os.File, err error) {
	var hostName *string = flag.String("host", "127.0.0.1", "Hostname to certify")
	priv, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		return nil, nil, err
	}
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
	if err != nil {
		return nil, nil, err
	}
	clientCert, err = os.CreateTemp("./test", "testCert")
	if err != nil {
		return nil, nil, err
	}

	pem.Encode(clientCert, &pem.Block{Type: "CERTIFICATE", Bytes: cert})
	clientCert.Close()

	clientKey, err = os.CreateTemp("./test", "testKey")
	if err != nil {
		return nil, nil, err
	}
	pem.Encode(clientKey, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})
	clientKey.Close()

	return clientCert, clientKey, nil
}
