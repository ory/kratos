package saml_test

import (
	"io/ioutil"
	"testing"

	"github.com/ory/kratos/selfservice/strategy/saml"
	"github.com/stretchr/testify/require"

	"gotest.tools/assert"
)

func TestInitMiddleWareWithMetadata(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	saml.DestroyMiddlewareIfExists("samlProvider")

	middleWare, _, _, err := InitTestMiddlewareWithMetadata(t,
		"file://testdata/SP_IDPMetadata.xml")

	require.NoError(t, err)
	assert.Check(t, middleWare != nil)
	assert.Check(t, middleWare.ServiceProvider.IDPMetadata != nil)
	assert.Check(t, middleWare.ServiceProvider.MetadataURL.Path == "/self-service/methods/saml/metadata/:provider")
	assert.Check(t, middleWare.ServiceProvider.IDPMetadata.EntityID == "https://idp.testshib.org/idp/shibboleth")
}

func TestInitMiddleWareWithoutMetadata(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	saml.DestroyMiddlewareIfExists("samlProvider")

	middleWare, _, _, err := InitTestMiddlewareWithoutMetadata(t,
		"https://samltest.id/idp/profile/SAML2/Redirect/SSO",
		"https://samltest.id/saml/idp",
		"file://testdata/samlkratos.crt",
		"https://samltest.id/idp/profile/SAML2/Redirect/SSO")

	require.NoError(t, err)
	assert.Check(t, middleWare != nil)
	assert.Check(t, middleWare.ServiceProvider.IDPMetadata != nil)
	assert.Check(t, middleWare.ServiceProvider.MetadataURL.Path == "/self-service/methods/saml/metadata/:provider")
	assert.Check(t, middleWare.ServiceProvider.IDPMetadata.EntityID == "https://samltest.id/saml/idp")
}

func TestGetMiddleware(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	saml.DestroyMiddlewareIfExists("samlProvider")

	InitTestMiddlewareWithMetadata(t,
		"file://testdata/SP_IDPMetadata.xml")

	middleWare, err := saml.GetMiddleware("samlProvider")

	require.NoError(t, err)
	assert.Check(t, middleWare != nil)
	assert.Check(t, middleWare.ServiceProvider.IDPMetadata != nil)
	assert.Check(t, middleWare.ServiceProvider.MetadataURL.Path == "/self-service/methods/saml/metadata/:provider")
	assert.Check(t, middleWare.ServiceProvider.IDPMetadata.EntityID == "https://idp.testshib.org/idp/shibboleth")
}

func TestMustParseCertificate(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	saml.DestroyMiddlewareIfExists("samlProvider")

	certificate, err := ioutil.ReadFile("testdata/samlkratos.crt")
	require.NoError(t, err)

	cert, err := saml.MustParseCertificate(certificate)

	require.NoError(t, err)
	assert.Check(t, cert.Issuer.Country[0] == "AU")
	assert.Check(t, cert.Issuer.Organization[0] == "Internet Widgits Pty Ltd")
	assert.Check(t, cert.Issuer.Province[0] == "Some-State")
	assert.Check(t, cert.Subject.Country[0] == "AU")
	assert.Check(t, cert.Subject.Organization[0] == "Internet Widgits Pty Ltd")
	assert.Check(t, cert.Subject.Province[0] == "Some-State")
	assert.Check(t, cert.NotBefore.String() == "2022-02-21 11:08:20 +0000 UTC")
	assert.Check(t, cert.NotAfter.String() == "2023-02-21 11:08:20 +0000 UTC")
	assert.Check(t, cert.SerialNumber.String() == "485646075402096403898806020771481121115125312047")
}
