// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oidc

import (
	"fmt"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDecodeQuery(t *testing.T) {
	query := url.Values{
		"user": []string{`{"name": {"firstName": "first", "lastName": "last"}, "email": "email@email.com"}`},
	}

	for k, tc := range []struct {
		claims     *Claims
		familyName string
		givenName  string
		lastName   string
	}{
		{claims: &Claims{}, familyName: "first", givenName: "first", lastName: "last"},
		{claims: &Claims{FamilyName: "fam"}, familyName: "fam", givenName: "first", lastName: "last"},
		{claims: &Claims{FamilyName: "fam", GivenName: "giv"}, familyName: "fam", givenName: "giv", lastName: "last"},
		{claims: &Claims{FamilyName: "fam", GivenName: "giv", LastName: "las"}, familyName: "fam", givenName: "giv", lastName: "las"},
	} {
		t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
			decodeQuery(query, tc.claims)
			assert.Equal(t, tc.familyName, tc.claims.FamilyName)
			assert.Equal(t, tc.givenName, tc.claims.GivenName)
			assert.Equal(t, tc.lastName, tc.claims.LastName)
			// Never extract email from the query, as the same info can be extracted and verified from the ID token.
			assert.Empty(t, tc.claims.Email)
		})
	}

}
