package saml_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"testing"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/selfservice/strategy/saml"
	"github.com/ory/x/sqlxx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	gotest "gotest.tools/assert"
	is "gotest.tools/assert/cmp"
)

func TestGetAndDecryptAssertion(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	saml.DestroyMiddlewareIfExists("samlProvider")

	middleware, _, _, _ := InitTestMiddlewareWithMetadata(t,
		"file://testdata/SP_IDPMetadata.xml")

	assertion, err := GetAndDecryptAssertion(t, "./testdata/SP_SamlResponse.xml", middleware.ServiceProvider.Key)

	require.NoError(t, err)
	gotest.Check(t, assertion != nil)
}

func TestGetAttributesFromAssertion(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	saml.DestroyMiddlewareIfExists("samlProvider")

	middleware, strategy, _, _ := InitTestMiddlewareWithMetadata(t,
		"file://testdata/SP_IDPMetadata.xml")

	assertion, _ := GetAndDecryptAssertion(t, "./testdata/SP_SamlResponse.xml", middleware.ServiceProvider.Key)

	mapAttributes, err := strategy.GetAttributesFromAssertion(assertion)

	require.NoError(t, err)
	gotest.Check(t, mapAttributes["urn:oid:0.9.2342.19200300.100.1.1"][0] == "myself")
	gotest.Check(t, mapAttributes["urn:oid:1.3.6.1.4.1.5923.1.1.1.1"][0] == "Member")
	gotest.Check(t, mapAttributes["urn:oid:1.3.6.1.4.1.5923.1.1.1.1"][1] == "Staff")
	gotest.Check(t, mapAttributes["urn:oid:1.3.6.1.4.1.5923.1.1.1.6"][0] == "myself@testshib.org")
	gotest.Check(t, mapAttributes["urn:oid:2.5.4.4"][0] == "And I")
	gotest.Check(t, mapAttributes["urn:oid:1.3.6.1.4.1.5923.1.1.1.9"][0] == "Member@testshib.org")
	gotest.Check(t, mapAttributes["urn:oid:1.3.6.1.4.1.5923.1.1.1.9"][1] == "Staff@testshib.org")
	gotest.Check(t, mapAttributes["urn:oid:2.5.4.42"][0] == "Me Myself")
	gotest.Check(t, mapAttributes["urn:oid:1.3.6.1.4.1.5923.1.1.1.7"][0] == "urn:mace:dir:entitlement:common-lib-terms")
	gotest.Check(t, mapAttributes["urn:oid:2.5.4.3"][0] == "Me Myself And I")
	gotest.Check(t, mapAttributes["urn:oid:2.5.4.20"][0] == "555-5555")

	t.Log(mapAttributes)
}

func TestCreateAuthRequest(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	saml.DestroyMiddlewareIfExists("samlProvider")

	middleware, _, _, _ := InitTestMiddlewareWithMetadata(t,
		"file://testdata/SP_IDPMetadata.xml")

	authReq, err := middleware.ServiceProvider.MakeAuthenticationRequest("https://samltest.id/idp/profile/SAML2/Redirect/SSO", "saml.HTTPPostBinding", "saml.HTTPPostBinding")
	require.NoError(t, err)

	matchACS, err := regexp.MatchString(`http://127.0.0.1:\d{5}/self-service/methods/saml/acs`, authReq.AssertionConsumerServiceURL)
	require.NoError(t, err)
	gotest.Check(t, matchACS)

	matchMetadata, err := regexp.MatchString(`http://127.0.0.1:\d{5}/self-service/methods/saml/metadata`, authReq.Issuer.Value)
	require.NoError(t, err)
	gotest.Check(t, matchMetadata)

	gotest.Check(t, is.Equal(authReq.Destination, "https://samltest.id/idp/profile/SAML2/Redirect/SSO"))
}

func TestProvider(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	saml.DestroyMiddlewareIfExists("samlProvider")

	_, strategy, _, _ := InitTestMiddlewareWithMetadata(t,
		"file://testdata/SP_IDPMetadata.xml")

	provider, err := strategy.Provider(context.Background(), "samlProvider")
	require.NoError(t, err)
	gotest.Check(t, provider != nil)
	gotest.Check(t, provider.Config().ID == "samlProvider")
	gotest.Check(t, provider.Config().Label == "samlProviderLabel")
}

func TestConfig(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	saml.DestroyMiddlewareIfExists("samlProvider")

	_, strategy, _, _ := InitTestMiddlewareWithMetadata(t,
		"file://testdata/SP_IDPMetadata.xml")

	config, err := strategy.Config(context.Background())
	require.NoError(t, err)
	gotest.Check(t, config != nil)
	gotest.Check(t, len(config.SAMLProviders) == 1)
	gotest.Check(t, config.SAMLProviders[0].ID == "samlProvider")
	gotest.Check(t, config.SAMLProviders[0].Label == "samlProviderLabel")
}

func TestID(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	saml.DestroyMiddlewareIfExists("samlProvider")

	_, strategy, _, _ := InitTestMiddlewareWithMetadata(t,
		"file://testdata/SP_IDPMetadata.xml")

	id := strategy.ID()
	gotest.Check(t, id == "saml")
}

func TestCountActiveCredentials(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	saml.DestroyMiddlewareIfExists("samlProvider")

	_, strategy, _, _ := InitTestMiddlewareWithMetadata(t,
		"file://testdata/SP_IDPMetadata.xml")

	mapCredentials := make(map[identity.CredentialsType]identity.Credentials)

	var b bytes.Buffer
	err := json.NewEncoder(&b).Encode(identity.CredentialsSAML{
		Providers: []identity.CredentialsSAMLProvider{
			{
				Subject:  "testUserID",
				Provider: "saml",
			}},
	})
	require.NoError(t, err)

	mapCredentials[identity.CredentialsTypeSAML] = identity.Credentials{
		Type:        identity.CredentialsTypeSAML,
		Identifiers: []string{"saml:testUserID"},
		Config:      b.Bytes(),
	}

	count, err := strategy.CountActiveCredentials(mapCredentials)
	require.NoError(t, err)
	gotest.Check(t, count == 1)
}

func TestGetRegistrationIdentity(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	saml.DestroyMiddlewareIfExists("samlProvider")

	middleware, strategy, _, _ := InitTestMiddlewareWithMetadata(t,
		"file://testdata/SP_IDPMetadata.xml")

	provider, _ := strategy.Provider(context.Background(), "samlProvider")
	assertion, _ := GetAndDecryptAssertion(t, "./testdata/SP_SamlResponse.xml", middleware.ServiceProvider.Key)
	attributes, _ := strategy.GetAttributesFromAssertion(assertion)
	claims, _ := provider.Claims(context.Background(), strategy.D().Config(), attributes, "samlProvider")

	i, err := strategy.GetRegistrationIdentity(nil, context.Background(), provider, claims, false)
	require.NoError(t, err)
	gotest.Check(t, i != nil)
}

func TestCountActiveFirstFactorCredentials(t *testing.T) {
	_, reg := internal.NewFastRegistryWithMocks(t)
	strategy := saml.NewStrategy(reg)

	toJson := func(c identity.CredentialsSAML) []byte {
		out, err := json.Marshal(&c)
		require.NoError(t, err)
		return out
	}

	for k, tc := range []struct {
		in       identity.CredentialsCollection
		expected int
	}{
		{
			in: identity.CredentialsCollection{{
				Type:   strategy.ID(),
				Config: sqlxx.JSONRawMessage{},
			}},
		},
		{
			in: identity.CredentialsCollection{{
				Type: strategy.ID(),
				Config: toJson(identity.CredentialsSAML{Providers: []identity.CredentialsSAMLProvider{
					{Subject: "foo", Provider: "bar"},
				}}),
			}},
		},
		{
			in: identity.CredentialsCollection{{
				Type:        strategy.ID(),
				Identifiers: []string{""},
				Config: toJson(identity.CredentialsSAML{Providers: []identity.CredentialsSAMLProvider{
					{Subject: "foo", Provider: "bar"},
				}}),
			}},
		},
		{
			in: identity.CredentialsCollection{{
				Type:        strategy.ID(),
				Identifiers: []string{"bar:"},
				Config: toJson(identity.CredentialsSAML{Providers: []identity.CredentialsSAMLProvider{
					{Subject: "foo", Provider: "bar"},
				}}),
			}},
		},
		{
			in: identity.CredentialsCollection{{
				Type:        strategy.ID(),
				Identifiers: []string{":foo"},
				Config: toJson(identity.CredentialsSAML{Providers: []identity.CredentialsSAMLProvider{
					{Subject: "foo", Provider: "bar"},
				}}),
			}},
		},
		{
			in: identity.CredentialsCollection{{
				Type:        strategy.ID(),
				Identifiers: []string{"not-bar:foo"},
				Config: toJson(identity.CredentialsSAML{Providers: []identity.CredentialsSAMLProvider{
					{Subject: "foo", Provider: "bar"},
				}}),
			}},
		},
		{
			in: identity.CredentialsCollection{{
				Type:        strategy.ID(),
				Identifiers: []string{"bar:not-foo"},
				Config: toJson(identity.CredentialsSAML{Providers: []identity.CredentialsSAMLProvider{
					{Subject: "foo", Provider: "bar"},
				}}),
			}},
		},
		{
			in: identity.CredentialsCollection{{
				Type:        strategy.ID(),
				Identifiers: []string{"bar:foo"},
				Config: toJson(identity.CredentialsSAML{Providers: []identity.CredentialsSAMLProvider{
					{Subject: "foo", Provider: "bar"},
				}}),
			}},
			expected: 1,
		},
	} {
		t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
			in := make(map[identity.CredentialsType]identity.Credentials)
			for _, v := range tc.in {
				in[v.Type] = v
			}
			actual, err := strategy.CountActiveFirstFactorCredentials(in)
			require.NoError(t, err)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestModifyIdentityTraits(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	saml.DestroyMiddlewareIfExists("samlProvider")

}
