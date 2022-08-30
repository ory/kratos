package saml_test

import (
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"reflect"
	"testing"

	"github.com/ory/kratos/selfservice/strategy/saml"
	"github.com/stretchr/testify/require"
	"gotest.tools/assert"
	is "gotest.tools/assert/cmp"
)

type Metadata struct {
	XMLName         xml.Name `xml:"EntityDescriptor"`
	Text            string   `xml:",chardata"`
	Xmlns           string   `xml:"xmlns,attr"`
	ValidUntil      string   `xml:"validUntil,attr"`
	EntityID        string   `xml:"entityID,attr"`
	SPSSODescriptor struct {
		Text                       string `xml:",chardata"`
		Xmlns                      string `xml:"xmlns,attr"`
		ValidUntil                 string `xml:"validUntil,attr"`
		ProtocolSupportEnumeration string `xml:"protocolSupportEnumeration,attr"`
		AuthnRequestsSigned        string `xml:"AuthnRequestsSigned,attr"`
		WantAssertionsSigned       string `xml:"WantAssertionsSigned,attr"`
		KeyDescriptor              []struct {
			Text    string `xml:",chardata"`
			Use     string `xml:"use,attr"`
			KeyInfo struct {
				Text     string `xml:",chardata"`
				Xmlns    string `xml:"xmlns,attr"`
				X509Data struct {
					Text            string `xml:",chardata"`
					Xmlns           string `xml:"xmlns,attr"`
					X509Certificate struct {
						Text  string `xml:",chardata"`
						Xmlns string `xml:"xmlns,attr"`
					} `xml:"X509Certificate"`
				} `xml:"X509Data"`
			} `xml:"KeyInfo"`
			EncryptionMethod []struct {
				Text      string `xml:",chardata"`
				Algorithm string `xml:"Algorithm,attr"`
			} `xml:"EncryptionMethod"`
		} `xml:"KeyDescriptor"`
		SingleLogoutService struct {
			Text             string `xml:",chardata"`
			Binding          string `xml:"Binding,attr"`
			Location         string `xml:"Location,attr"`
			ResponseLocation string `xml:"ResponseLocation,attr"`
		} `xml:"SingleLogoutService"`
		AssertionConsumerService []struct {
			Text     string `xml:",chardata"`
			Binding  string `xml:"Binding,attr"`
			Location string `xml:"Location,attr"`
			Index    string `xml:"index,attr"`
		} `xml:"AssertionConsumerService"`
	} `xml:"SPSSODescriptor"`
}

func TestXmlMetadataExist(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	saml.DestroyMiddlewareIfExists("samlProvider")

	_, _, ts, err := InitTestMiddlewareWithMetadata(t, "file://testdata/SP_IDPMetadata.xml")
	assert.NilError(t, err)
	res, err := NewTestClient(t, nil).Get(ts.URL + "/self-service/methods/saml/metadata/samlProvider")
	assert.NilError(t, err)
	body, _ := ioutil.ReadAll(res.Body)
	fmt.Println(body)
	assert.Check(t, is.Equal(http.StatusOK, res.StatusCode))
	assert.Check(t, is.Equal("text/xml", res.Header.Get("Content-Type")))
}

func TestXmlMetadataValues(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	saml.DestroyMiddlewareIfExists("samlProvider")

	_, _, ts, _ := InitTestMiddlewareWithMetadata(t, "file://testdata/SP_IDPMetadata.xml")
	res, _ := NewTestClient(t, nil).Get(ts.URL + "/self-service/methods/saml/metadata/samlProvider")
	body, _ := io.ReadAll(res.Body)

	assert.Check(t, is.Equal(http.StatusOK, res.StatusCode))
	assert.Check(t, is.Equal("text/xml",
		res.Header.Get("Content-Type")))

	expectedMetadata, err := ioutil.ReadFile("./testdata/expected_metadata.xml")
	assert.NilError(t, err)

	// The string is parse to a struct
	var expectedStructMetadata Metadata
	err = xml.Unmarshal(expectedMetadata, &expectedStructMetadata)
	require.NoError(t, err)

	var obtainedStructureMetadata Metadata
	err = xml.Unmarshal(body, &obtainedStructureMetadata)
	require.NoError(t, err)

	// We delete data that is likely to change naturally
	expectedStructMetadata.SPSSODescriptor.AssertionConsumerService[0].Location = ""
	expectedStructMetadata.SPSSODescriptor.AssertionConsumerService[1].Location = ""
	obtainedStructureMetadata.SPSSODescriptor.AssertionConsumerService[0].Location = ""
	obtainedStructureMetadata.SPSSODescriptor.AssertionConsumerService[1].Location = ""
	expectedStructMetadata.ValidUntil = ""
	expectedStructMetadata.SPSSODescriptor.ValidUntil = ""
	obtainedStructureMetadata.ValidUntil = ""
	obtainedStructureMetadata.SPSSODescriptor.ValidUntil = ""
	expectedStructMetadata.EntityID = ""
	obtainedStructureMetadata.EntityID = ""

	assert.Check(t, reflect.DeepEqual(expectedStructMetadata, obtainedStructureMetadata))
}
