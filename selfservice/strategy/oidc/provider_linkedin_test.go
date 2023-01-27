// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

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

func TestProviderLinkedin_Claims(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", "https://api.linkedin.com/v2/me?projection=(id,localizedFirstName,localizedLastName,profilePicture(displayImage~digitalmediaAsset:playableStreams))",
		func(req *http.Request) (*http.Response, error) {
			resp, err := httpmock.NewJsonResponse(200, map[string]interface{}{
				"localizedLastName": "Doe",
				"profilePicture": map[string]interface{}{
					"displayImage": "urn:li:digitalmediaAsset:D5F04AQG1h-anJtsLeA",
					"displayImage~": map[string]interface{}{
						"elements": []map[string]interface{}{
							{
								"identifiers": []map[string]interface{}{
									{
										"identifier": "https://media-exp1.licdn.com/dms/image/D5F04AQG1h-anJtsLeA/profile-displayphoto-shrink_100_100/0/1661502148285?e=1674086400&v=beta&t=5zdsvAqGS1idrfNAai-zwp01K4hApGeJyRV8zBTjZck",
									},
								},
							},
							{
								"identifiers": []map[string]interface{}{
									{
										"identifier": "https://media-exp1.licdn.com/dms/image/D5F04AQG1h-anJtsLeA/profile-displayphoto-shrink_200_200/0/1661502148285?e=1674086400&v=beta&t=fbHprGci4vEpeQdzRTyt9YJjNXTgv-14h9ROr_DE5R4",
									},
								},
							},
							{
								"identifiers": []map[string]interface{}{
									{
										"identifier": "https://media-exp1.licdn.com/dms/image/D5F04AQG1h-anJtsLeA/profile-displayphoto-shrink_400_400/0/1661502148285?e=1674086400&v=beta&t=SMgXuiZ5DiiPf4iMrTPQWoCnFRh4Y-DfoWjopUZ0qCo",
									},
								},
							},
							{
								"identifiers": []map[string]interface{}{
									{
										"identifier": "https://media-exp1.licdn.com/dms/image/D5F04AQG1h-anJtsLeA/profile-displayphoto-shrink_800_800/0/1661502148285?e=1674086400&v=beta&t=QzC9b32DmLRIwU7P27LfJUqglXsJm056WkA_TTOHxHU",
									},
								},
							},
						},
					},
				},
				"localizedFirstName": "John",
				"id":                 "5foOWOiYXD",
			})
			return resp, err
		})

	httpmock.RegisterResponder("GET", "https://api.linkedin.com/v2/emailAddress?q=members&projection=(elements*(handle~))",
		func(req *http.Request) (*http.Response, error) {
			resp, err := httpmock.NewJsonResponse(200, map[string]interface{}{
				"elements": []map[string]interface{}{
					{
						"handle~": map[string]interface{}{
							"emailAddress": "john.doe@gmail.com",
						},
						"handle": "urn:li:emailAddress:-1810278278",
					},
				},
			})
			return resp, err
		},
	)

	httpmock.RegisterResponder("POST", "https://www.linkedin.com/oauth/v2/introspectToken",
		func(req *http.Request) (*http.Response, error) {
			resp, err := httpmock.NewJsonResponse(200, map[string]interface{}{
				"active":        true,
				"client_id":     "78e533g1dq0rs5",
				"authorized_at": 1668518671,
				"created_at":    1668518671,
				"status":        "active",
				"expires_at":    2673702676,
				"scope":         "r_emailaddress,r_liteprofile",
				"auth_type":     "3L",
			})
			return resp, err
		},
	)

	_, reg := internal.NewFastRegistryWithMocks(t)
	c := &oidc.Configuration{
		ID:           "linkedin",
		Provider:     "linkedin",
		ClientID:     "abcd",
		ClientSecret: "secret",
		Mapper:       "file://./stub/oidc.linkedin.jsonnet",
		Scope:        []string{"r_emailaddress", "r_liteprofile"},
	}
	linkedin := oidc.NewProviderLinkedIn(c, reg)

	const fakeLinkedinIDToken = "id_token_mock_"
	actual, err := linkedin.Claims(
		context.Background(),
		(&oauth2.Token{AccessToken: "foo", Expiry: time.Now().Add(time.Hour)}).WithExtra(map[string]interface{}{"id_token": fakeLinkedinIDToken}),
		url.Values{},
	)
	require.NoError(t, err)

	assert.Equal(t, &oidc.Claims{
		Issuer:    "https://login.linkedin.com/",
		Subject:   "5foOWOiYXD",
		GivenName: "John",
		LastName:  "Doe",
		Email:     "john.doe@gmail.com",
		Picture:   "https://media-exp1.licdn.com/dms/image/D5F04AQG1h-anJtsLeA/profile-displayphoto-shrink_400_400/0/1661502148285?e=1674086400&v=beta&t=SMgXuiZ5DiiPf4iMrTPQWoCnFRh4Y-DfoWjopUZ0qCo",
	}, actual)
}

func TestProviderLinkedin_No_Picture(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", "https://api.linkedin.com/v2/me?projection=(id,localizedFirstName,localizedLastName,profilePicture(displayImage~digitalmediaAsset:playableStreams))",
		func(req *http.Request) (*http.Response, error) {
			resp, err := httpmock.NewJsonResponse(200, map[string]interface{}{
				"localizedLastName":  "Doe",
				"localizedFirstName": "John",
				"id":                 "5foOWOiYXD",
			})
			return resp, err
		})

	httpmock.RegisterResponder("GET", "https://api.linkedin.com/v2/emailAddress?q=members&projection=(elements*(handle~))",
		func(req *http.Request) (*http.Response, error) {
			resp, err := httpmock.NewJsonResponse(200, map[string]interface{}{
				"elements": []map[string]interface{}{
					{
						"handle~": map[string]interface{}{
							"emailAddress": "john.doe@gmail.com",
						},
						"handle": "urn:li:emailAddress:-1810278278",
					},
				},
			})
			return resp, err
		},
	)

	httpmock.RegisterResponder("POST", "https://www.linkedin.com/oauth/v2/introspectToken",
		func(req *http.Request) (*http.Response, error) {
			resp, err := httpmock.NewJsonResponse(200, map[string]interface{}{
				"active":        true,
				"client_id":     "78e533g1dq0rs5",
				"authorized_at": 1668518671,
				"created_at":    1668518671,
				"status":        "active",
				"expires_at":    2673702676,
				"scope":         "r_emailaddress,r_liteprofile",
				"auth_type":     "3L",
			})
			return resp, err
		},
	)

	_, reg := internal.NewFastRegistryWithMocks(t)
	c := &oidc.Configuration{
		ID:           "linkedin",
		Provider:     "linkedin",
		ClientID:     "abcd",
		ClientSecret: "secret",
		Mapper:       "file://./stub/oidc.linkedin.jsonnet",
		Scope:        []string{"r_emailaddress", "r_liteprofile"},
	}
	linkedin := oidc.NewProviderLinkedIn(c, reg)

	const fakeLinkedinIDToken = "id_token_mock_"
	actual, err := linkedin.Claims(
		context.Background(),
		(&oauth2.Token{AccessToken: "foo", Expiry: time.Now().Add(time.Hour)}).WithExtra(map[string]interface{}{"id_token": fakeLinkedinIDToken}),
		url.Values{},
	)
	require.NoError(t, err)

	assert.Equal(t, &oidc.Claims{
		Issuer:    "https://login.linkedin.com/",
		Subject:   "5foOWOiYXD",
		GivenName: "John",
		LastName:  "Doe",
		Email:     "john.doe@gmail.com",
		Picture:   "",
	}, actual)
}
