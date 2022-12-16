package saml_test

import (
	"bytes"
	"crypto"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/xml"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/beevik/etree"
	"github.com/crewjam/saml"
	"github.com/crewjam/saml/logger"
	"github.com/crewjam/saml/samlsp"
	"github.com/crewjam/saml/xmlenc"
	"github.com/julienschmidt/httprouter"
	"github.com/ory/kratos/continuity"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/login"
	samlhandler "github.com/ory/kratos/selfservice/strategy/saml"
	"github.com/ory/kratos/x"

	"github.com/gofrs/uuid"
	"github.com/golang-jwt/jwt/v4"
	dsig "github.com/russellhaering/goxmldsig"
	"github.com/stretchr/testify/require"
	"gotest.tools/assert"
	"gotest.tools/golden"
)

type MiddlewareTest struct {
	AuthnRequest []byte
	SamlResponse []byte
	Key          *rsa.PrivateKey
	Certificate  *x509.Certificate
	IDPMetadata  []byte
	Middleware   *samlsp.Middleware
}

type IdentityProviderTest struct {
	SPKey         *rsa.PrivateKey
	SPCertificate *x509.Certificate
	SP            saml.ServiceProvider

	Key         crypto.PrivateKey
	Certificate *x509.Certificate
	IDP         saml.IdentityProvider
}

type mockServiceProviderProvider struct {
	GetServiceProviderFunc func(r *http.Request, serviceProviderID string) (*saml.EntityDescriptor, error)
}

func (mspp *mockServiceProviderProvider) GetServiceProvider(r *http.Request, serviceProviderID string) (*saml.EntityDescriptor, error) {
	return mspp.GetServiceProviderFunc(r, serviceProviderID)
}

func mustParseURL(s string) url.URL {
	rv, err := url.Parse(s)
	if err != nil {
		panic(err)
	}
	return *rv
}

func setSAMLTimeNow(timeStr string) {
	TimeNow = func() time.Time {
		rv, _ := time.Parse("Mon Jan 2 15:04:05.999999999 MST 2006", timeStr)
		return rv
	}

	saml.TimeNow = TimeNow
	jwt.TimeFunc = TimeNow
	saml.Clock = dsig.NewFakeClockAt(TimeNow())
}

func (test *MiddlewareTest) makeTrackedRequest(id string) (string, string) {
	uuid, _ := uuid.NewV4()

	codec := test.Middleware.RequestTracker.(samlsp.CookieRequestTracker).Codec
	index := uuid.String()
	token, err := codec.Encode(samlsp.TrackedRequest{
		Index:         index,
		SAMLRequestID: id,
		URI:           "/frob",
	})
	if err != nil {
		panic(err)
	}
	return token, index
}

func NewMiddlewareTest(t *testing.T) (*MiddlewareTest, *samlhandler.Strategy, *httptest.Server) {
	middlewareTest := MiddlewareTest{}

	samlhandler.DestroyMiddlewareIfExists("samlProvider")

	middleWare, strategy, ts, err := InitTestMiddlewareWithMetadata(t, "file://testdata/idp_metadata.xml")
	if err != nil {
		return nil, nil, nil
	}

	middlewareTest.Middleware = middleWare

	middlewareTest.Key = middlewareTest.Middleware.ServiceProvider.Key
	middlewareTest.Certificate = middlewareTest.Middleware.ServiceProvider.Certificate
	middlewareTest.IDPMetadata = golden.Get(t, "idp_metadata.xml")

	var metadata saml.EntityDescriptor
	if err := xml.Unmarshal(middlewareTest.IDPMetadata, &metadata); err != nil {
		panic(err)
	}

	return &middlewareTest, strategy, ts
}

func NewIdentifyProviderTest(t *testing.T, serviceProvider saml.ServiceProvider, tsURL string) *IdentityProviderTest {
	IDPtest := IdentityProviderTest{}

	IDPtest.SP = serviceProvider
	IDPtest.SPKey = IDPtest.SP.Key
	IDPtest.SPCertificate = IDPtest.SP.Certificate

	IDPtest.Key = mustParsePrivateKey(golden.Get(t, "idp_key.pem"))
	IDPtest.Certificate = mustParseCertificate(golden.Get(t, "idp_cert.pem"))

	IDPtest.IDP = saml.IdentityProvider{
		Key:         IDPtest.Key,
		Certificate: IDPtest.Certificate,
		Logger:      logger.DefaultLogger,
		MetadataURL: mustParseURL("https://idp.example.com/saml/metadata"),
		SSOURL:      mustParseURL("https://idp.example.com/saml/sso"),
		ServiceProviderProvider: &mockServiceProviderProvider{
			GetServiceProviderFunc: func(r *http.Request, serviceProviderID string) (*saml.EntityDescriptor, error) {
				if serviceProviderID == IDPtest.SP.MetadataURL.String() {
					return IDPtest.SP.Metadata(), nil
				}
				return nil, os.ErrNotExist
			},
		},
	}

	IDPtest.SP.IDPMetadata = IDPtest.IDP.Metadata()

	return &IDPtest
}

func NewIdpAuthnRequest(t *testing.T, idp *saml.IdentityProvider, acsURL string, issuer string, destination string, issueInstant string) (saml.IdpAuthnRequest, string) {
	uuid, err := uuid.NewV4()
	assert.NilError(t, err)
	id := "id-" + strings.Replace(uuid.String(), "-", "", -1)

	authnRequest := saml.IdpAuthnRequest{
		Now: TimeNow(),
		IDP: idp,
		RequestBuffer: []byte("" +
			"<AuthnRequest xmlns=\"urn:oasis:names:tc:SAML:2.0:protocol\" " +
			"  AssertionConsumerServiceURL=\"" + acsURL + "\" " +
			"  Destination=\"" + destination + "\" " +
			"  ID=\"" + id + "\" " +
			"  IssueInstant=\"" + issueInstant + "\" ProtocolBinding=\"\" " +
			"  Version=\"2.0\">" +
			"  <Issuer xmlns=\"urn:oasis:names:tc:SAML:2.0:assertion\" " +
			"    Format=\"urn:oasis:names:tc:SAML:2.0:nameid-format:entity\">" + issuer + "</Issuer>" +
			"  <NameIDPolicy xmlns=\"urn:oasis:names:tc:SAML:2.0:protocol\" " +
			"    AllowCreate=\"true\">urn:oasis:names:tc:SAML:2.0:nameid-format:transient</NameIDPolicy>" +
			"</AuthnRequest>"),
	}

	authnRequest.HTTPRequest, err = http.NewRequest("POST", acsURL, nil)
	assert.NilError(t, err)
	assert.NilError(t, authnRequest.Validate())

	return authnRequest, id
}

func NewTestIdpAuthnRequest(t *testing.T, idp *saml.IdentityProvider, acsURL string, issuer string) (saml.IdpAuthnRequest, string) {
	authnRequest, id := NewIdpAuthnRequest(t, idp, acsURL, issuer, "https://idp.example.com/saml/sso", "2014-01-01T01:57:09Z")
	return authnRequest, id
}

func MakeAssertion(t *testing.T, authnRequest *saml.IdpAuthnRequest, userSession *saml.Session) {
	err := saml.DefaultAssertionMaker{}.MakeAssertion(authnRequest, userSession)
	assert.NilError(t, err)
}

func prepareTestEnvironment(t *testing.T) (*MiddlewareTest, *samlhandler.Strategy, *IdentityProviderTest, saml.IdpAuthnRequest, string) {
	// Set timeNow for SAML Requests and Responses
	setSAMLTimeNow("Wed Jan 1 01:57:09.123456789 UTC 2014")

	// Create a SAML SP
	testMiddleware, strategy, ts := NewMiddlewareTest(t)

	// Create a SAML IdP
	testIDP := NewIdentifyProviderTest(t, testMiddleware.Middleware.ServiceProvider, ts.URL)

	// SP ACS URL
	acsURL := ts.URL + "/self-service/methods/saml/acs/samlProvider"

	// Create a SAML AuthnRequest as it would be taken into account by the IdP
	// so that it can send the SAML Response back to the SP via the SP ACS
	authnRequest, authnRequestID := NewTestIdpAuthnRequest(t, &testIDP.IDP, acsURL, testMiddleware.Middleware.ServiceProvider.EntityID)

	return testMiddleware, strategy, testIDP, authnRequest, authnRequestID
}

func startContinuity(resp *httptest.ResponseRecorder, r *http.Request, strategy *samlhandler.Strategy) {
	conf := strategy.D().Config()
	f, _ := login.NewFlow(conf, conf.SelfServiceFlowLoginRequestLifespan(r.Context()), strategy.D().GenerateCSRFToken(r), r, flow.TypeBrowser)
	strategy.D().LoginFlowPersister().CreateLoginFlow(r.Context(), f)
	state := x.NewUUID().String()

	strategy.D().RelayStateContinuityManager().Pause(r.Context(), resp, r, "ory_kratos_saml_auth_code_session",
		continuity.WithPayload(&authCodeContainer{
			State:  state,
			FlowID: f.ID.String(),
		}),
		continuity.WithLifespan(time.Minute*30))
}

func initRouterParams() httprouter.Params {
	ps := httprouter.Params{
		httprouter.Param{
			Key:   "provider",
			Value: "samlProvider",
		},
	}
	return ps
}

func prepareTestEnvironmentTwoServiceProvider(t *testing.T) (*MiddlewareTest, *MiddlewareTest, *samlhandler.Strategy, *IdentityProviderTest, saml.IdpAuthnRequest, string) {
	// Set timeNow for SAML Requests and Responses
	setSAMLTimeNow("Wed Jan 1 01:57:09.123456789 UTC 2014")

	// Create a SAML SP
	testMiddleware, strategy, ts := NewMiddlewareTest(t)

	// Create a SAML IdP
	testIDP := NewIdentifyProviderTest(t, testMiddleware.Middleware.ServiceProvider, ts.URL)

	// SP ACS URL
	acsURL := ts.URL + "/self-service/methods/saml/acs/samlProvider"

	// Create a SAML AuthnRequest as it would be taken into account by the IdP
	// so that it can send the SAML Response back to the SP via the SP ACS
	authnRequest, authnRequestID := NewTestIdpAuthnRequest(t, &testIDP.IDP, acsURL, testMiddleware.Middleware.ServiceProvider.EntityID)

	return testMiddleware, nil, strategy, testIDP, authnRequest, authnRequestID
}

func PrepareTestSAMLResponse(t *testing.T, testMiddleware *MiddlewareTest, authnRequest saml.IdpAuthnRequest, authnRequestID string) saml.IdpAuthnRequest {
	// User session
	userSession := &saml.Session{
		ID:        "f00df00df00d",
		UserEmail: "alice@example.com",
	}

	return PrepareTestSAMLResponseWithSession(t, testMiddleware, authnRequest, authnRequestID, userSession)
}

func PrepareTestSAMLResponseWithSession(t *testing.T, testMiddleware *MiddlewareTest, authnRequest saml.IdpAuthnRequest, authnRequestID string, userSession *saml.Session) saml.IdpAuthnRequest {
	// Make SAML Assertion
	MakeAssertion(t, &authnRequest, userSession)

	// Make SAML Response
	authnRequest.MakeResponse()

	return authnRequest
}

func PrepareTestSAMLResponseHTTPRequest(t *testing.T, testMiddleware *MiddlewareTest, authnRequest saml.IdpAuthnRequest, authnRequestID string, responseStr string) *http.Request {
	// Prepare SAMLResponse body attribute
	v1 := &url.Values{}
	v1.Set("SAMLResponse", base64.StdEncoding.EncodeToString([]byte(responseStr)))

	// Set SAML AuthnRequest HTTP Request body with the SAML Response
	req := authnRequest.HTTPRequest
	req, err := http.NewRequest(req.Method, req.URL.String(), bytes.NewReader([]byte(v1.Encode())))
	assert.NilError(t, err)

	// Make tracked request and get its index
	trackedRequestToken, trackedRequestIndex := testMiddleware.makeTrackedRequest(authnRequestID)

	// Set SAML AuthnRequest HTTP Request headers Content-Type and session cookie
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Cookie", ""+
		"saml_"+trackedRequestIndex+"="+trackedRequestToken)

	return req
}

func GetAndDecryptAssertionEl(t *testing.T, testMiddleware *MiddlewareTest, responseDoc *etree.Document) *etree.Element {
	// Get the Encrypted Assertion Data
	spKey := testMiddleware.Middleware.ServiceProvider.Key
	encryptedAssertionDataEl := responseDoc.Element.FindElement("//EncryptedAssertion/EncryptedData")

	// Decrypt the Encrypted Assertion
	plaintextAssertion, err := xmlenc.Decrypt(spKey, encryptedAssertionDataEl)
	require.NoError(t, err)
	stringAssertion := string(plaintextAssertion)
	newAssertion := etree.NewDocument()
	newAssertion.ReadFromString(stringAssertion)

	return newAssertion.Root()
}

// Replace the Encrypted Assertion by the modified Assertion
func ReplaceResponseAssertion(t *testing.T, responseEl *etree.Element, newAssertionEl *etree.Element) {
	encryptedAssertionEl := responseEl.FindElement("//EncryptedAssertion")
	encryptedAssertionEl.Parent().RemoveChild(encryptedAssertionEl)
	responseEl.AddChild(newAssertionEl)
}

// Remove the SAML Response signature
func RemoveResponseSignature(responseDoc *etree.Document) {
	responseSignatureEl := responseDoc.FindElement("//Signature")
	responseSignatureEl.Parent().RemoveChild(responseSignatureEl)
}

func RemoveAssertionSignature(responseDoc *etree.Document) {
	assertionSignatureEl := responseDoc.FindElement("//Assertion/Signature")
	assertionSignatureEl.Parent().RemoveChild(assertionSignatureEl)
}
