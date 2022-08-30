package saml_test

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"encoding/xml"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/beevik/etree"
	crewjamsaml "github.com/crewjam/saml"
	"github.com/crewjam/saml/samlsp"
	"github.com/crewjam/saml/xmlenc"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
	"gotest.tools/golden"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/internal/testhelpers"
	"github.com/ory/kratos/selfservice/strategy/saml"
	"github.com/ory/kratos/x"
)

var TimeNow = func() time.Time { return time.Now().UTC() }
var RandReader = rand.Reader

func ViperSetProviderConfig(t *testing.T, conf *config.Config, SAMLProvider ...saml.Configuration) {
	conf.MustSet(context.Background(), config.ViperKeySelfServiceStrategyConfig+"."+string(identity.CredentialsTypeSAML)+".config", &saml.ConfigurationCollection{SAMLProviders: SAMLProvider})
	conf.MustSet(context.Background(), config.ViperKeySelfServiceStrategyConfig+"."+string(identity.CredentialsTypeSAML)+".enabled", true)
}

func NewTestClient(t *testing.T, jar *cookiejar.Jar) *http.Client {
	if jar == nil {
		j, err := cookiejar.New(nil)
		jar = j
		require.NoError(t, err)
	}
	return &http.Client{
		Jar: jar,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 20 {
				for k, v := range via {
					t.Logf("Failed with redirect (%d): %s", k, v.URL.String())
				}
				return errors.New("stopped after 20 redirects")
			}
			return nil
		},
	}
}

// AssertSystemError asserts an error ui response
func AssertSystemError(t *testing.T, errTS *httptest.Server, res *http.Response, body []byte, code int, reason string) {
	require.Contains(t, res.Request.URL.String(), errTS.URL, "%s", body)

	assert.Equal(t, int64(code), gjson.GetBytes(body, "code").Int(), "%s", body)
	assert.Contains(t, gjson.GetBytes(body, "reason").String(), reason, "%s", body)
}

func mustParseCertificate(pemStr []byte) *x509.Certificate {
	b, _ := pem.Decode(pemStr)
	if b == nil {
		panic("cannot parse PEM")
	}
	cert, err := x509.ParseCertificate(b.Bytes)
	if err != nil {
		panic(err)
	}
	return cert
}

func mustParsePrivateKey(pemStr []byte) crypto.PrivateKey {
	b, _ := pem.Decode(pemStr)
	if b == nil {
		panic("cannot parse PEM")
	}
	k, err := x509.ParsePKCS1PrivateKey(b.Bytes)
	if err != nil {
		panic(err)
	}
	return k
}

func InitTestMiddleware(t *testing.T, idpInformation map[string]string) (*samlsp.Middleware, *saml.Strategy, *httptest.Server, error) {
	conf, reg := internal.NewFastRegistryWithMocks(t)

	strategy := saml.NewStrategy(reg)
	routerP := x.NewRouterPublic()
	routerA := x.NewRouterAdmin()
	ts, _ := testhelpers.NewKratosServerWithRouters(t, reg, routerP, routerA)

	attributesMap := make(map[string]string)
	attributesMap["id"] = "mail"
	attributesMap["firstname"] = "givenName"
	attributesMap["lastname"] = "sn"
	attributesMap["email"] = "mail"

	// Initiates the service provider
	ViperSetProviderConfig(
		t,
		conf,
		saml.Configuration{
			ID:             "samlProvider",
			Label:          "samlProviderLabel",
			Provider:       "generic",
			PublicCertPath: "file://testdata/myservice.cert",
			PrivateKeyPath: "file://testdata/myservice.key",
			Mapper:         "file://testdata/saml.jsonnet",
			AttributesMap:  attributesMap,
			IDPInformation: idpInformation,
		},
	)

	conf.MustSet(context.Background(), config.ViperKeySelfServiceRegistrationEnabled, true)
	testhelpers.SetDefaultIdentitySchema(conf, "file://testdata/registration.schema.json")
	conf.MustSet(context.Background(), config.HookStrategyKey(config.ViperKeySelfServiceRegistrationAfter,
		identity.CredentialsTypeSAML.String()), []config.SelfServiceHook{{Name: "session"}})

	t.Logf("Kratos Public URL: %s", ts.URL)

	// Instantiates the MiddleWare
	_, err := NewTestClient(t, nil).Get(ts.URL + "/self-service/methods/saml/metadata/samlProvider")
	require.NoError(t, err)
	middleware, err := saml.GetMiddleware("samlProvider")
	require.NoError(t, err)
	middleware.ServiceProvider.Key = mustParsePrivateKey(golden.Get(t, "key.pem")).(*rsa.PrivateKey)
	middleware.ServiceProvider.Certificate = mustParseCertificate(golden.Get(t, "cert.pem"))

	return middleware, strategy, ts, err
}

func InitTestMiddlewareWithMetadata(t *testing.T, metadataURL string) (*samlsp.Middleware, *saml.Strategy, *httptest.Server, error) {
	idpInformation := make(map[string]string)
	idpInformation["idp_metadata_url"] = metadataURL

	return InitTestMiddleware(t, idpInformation)
}

func InitTestMiddlewareWithoutMetadata(t *testing.T, idpSsoUrl string, idpEntityId string,
	idpCertifiatePath string, idpLogoutUrl string) (*samlsp.Middleware, *saml.Strategy, *httptest.Server, error) {

	idpInformation := make(map[string]string)
	idpInformation["idp_sso_url"] = idpSsoUrl
	idpInformation["idp_entity_id"] = idpEntityId
	idpInformation["idp_certificate_path"] = idpCertifiatePath
	idpInformation["idp_logout_url"] = idpLogoutUrl

	return InitTestMiddleware(t, idpInformation)
}

func GetAndDecryptAssertion(t *testing.T, samlResponseFile string, key *rsa.PrivateKey) (*crewjamsaml.Assertion, error) {
	// Load saml response test file
	samlResponse, err := ioutil.ReadFile(samlResponseFile)
	require.NoError(t, err)

	// Decrypt saml response assertion
	doc := etree.NewDocument()
	err = doc.ReadFromBytes(samlResponse)
	require.NoError(t, err)
	responseEl := doc.Root()
	el := responseEl.FindElement("//EncryptedAssertion/EncryptedData")
	plaintextAssertion, err := xmlenc.Decrypt(key, el)
	require.NoError(t, err)

	assertion := &crewjamsaml.Assertion{}
	err = xml.Unmarshal(plaintextAssertion, assertion)
	require.NoError(t, err)

	return assertion, err
}
