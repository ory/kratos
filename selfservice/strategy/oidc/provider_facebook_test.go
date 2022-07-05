package oidc_test

import (
	"context"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/jarcoal/httpmock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"

	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/selfservice/strategy/oidc"
)

const fakeIDToken = "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyLCJleHAiOjk5OTk5OTk5OTksImF1ZCI6ImFiY2QiLCJpc3MiOiJodHRwczovL3Jhdy5naXRodWJ1c2VyY29udGVudC5jb20vYWVuZWFzci9wcml2YXRlLW9pZGMvbWFzdGVyL3Rva2VuIn0.G9v8pJXJrEOgdJ5ecE6sIIcTH_p-RKkBaImfZY5DDVCl7h5GEis1n3GKKYbL_O3fj8Fu-WzI2mquI8S8BOVCQ6wN0XtrqJv22iX_nzeVHc4V_JWV1q7hg2gPpoFFcnF3KKtxZLvDOA8ujsDbAXmoBu0fEBdwCN56xLOOKQDzULyfijuAa8hrCwespZ9HaqcHzD3iHf_Utd4nHqlTM-6upWpKIMkplS_NGcxrfIRIWusZ0wob6ryy8jECD9QeZpdTGUozq-YM64lZfMOZzuLuqichH_PCMKFyB_tOZb6lDIiiSX4Irz7_YF-DP-LmfxgIW4934RqTCeFGGIP64h4xAA"

func TestProviderFacebook_Claims(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", "https://graph.facebook.com/me",
		func(req *http.Request) (*http.Response, error) {
			if _, ok := req.URL.Query()["appsecret_proof"]; !ok {
				resp, err := httpmock.NewJsonResponse(400, map[string]interface{}{
					"error": map[string]interface{}{
						"message":    "API calls from the server require an appsecret_proof argument",
						"type":       "GraphMethodException",
						"code":       100,
						"fbtrace_id": "Ay8LR3n5BsHm809VYpJ3eDM",
					},
				})
				return resp, err
			}
			resp, err := httpmock.NewJsonResponse(200, map[string]interface{}{
				"id":         "123456789012345",
				"name":       "John Doe",
				"first_name": "John",
				"last_name":  "Doe",
				"email":      "john.doe@example.com",
				"birthday":   "01/01/1990",
			})
			return resp, err
		},
	)

	httpmock.RegisterResponder("GET", "https://www.facebook.com/.well-known/openid-configuration",
		func(req *http.Request) (*http.Response, error) {
			resp, err := httpmock.NewJsonResponse(200, map[string]interface{}{
				"issuer": "https://www.facebook.com",
			})
			return resp, err
		},
	)

	_, reg := internal.NewFastRegistryWithMocks(t)
	c := &oidc.Configuration{
		ID:           "facebook",
		Provider:     "facebook",
		ClientID:     "abcd",
		ClientSecret: "secret",
		Mapper:       "file://./stub/oidc.facebook.jsonnet",
		Scope:        []string{"email"},
	}
	facebook := oidc.NewProviderFacebook(c, reg)

	actual, err := facebook.Claims(
		context.Background(),
		(&oauth2.Token{AccessToken: "foo", Expiry: time.Now().Add(time.Hour)}).WithExtra(map[string]interface{}{"id_token": fakeIDToken}),
		url.Values{},
	)
	require.NoError(t, err)

	assert.Equal(t, &oidc.Claims{
		Issuer:            "https://graph.facebook.com/me?fields=id,name,first_name,last_name,middle_name,email,picture,birthday,gender&appsecret_proof=773ba44693c7553d6ee20f61ea5d2757a9a4f4a44d2841ae4e95b52e4cd62db4",
		Subject:           "123456789012345",
		Name:              "John Doe",
		GivenName:         "John",
		FamilyName:        "Doe",
		Nickname:          "John Doe",
		PreferredUsername: "John Doe",
		Email:             "john.doe@example.com",
		EmailVerified:     true,
		Birthdate:         "01/01/1990",
	}, actual)
}
