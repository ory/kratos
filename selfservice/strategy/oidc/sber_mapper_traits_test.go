// Copyright © 2026 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oidc

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestApplySberClaimsToMapperTraitsOutput(t *testing.T) {
	t.Parallel()

	claims := &Claims{
		GivenName:  "Иван",
		FamilyName: "Петров",
		Email:      "User@Example.com",
		Birthdate:  "1990-05-01",
	}

	in := `{"identity":{"traits":{"given_name":"ИВАН","family_name":"ПЕТРОВ","email":"USER@EXAMPLE.COM","birthdate":"","extra":"x"}}}`

	out, err := applySberClaimsToMapperTraitsOutput("sber-ift", claims, in)
	require.NoError(t, err)
	require.Contains(t, out, `"given_name":"Иван"`)
	require.Contains(t, out, `"family_name":"Петров"`)
	require.Contains(t, out, `"email":"User@Example.com"`)
	require.Contains(t, out, `"birthdate":"1990-05-01"`)
	require.Contains(t, out, `"extra":"x"`)
}

func TestApplySberClaimsToMapperTraitsOutputNonSberNoop(t *testing.T) {
	t.Parallel()

	claims := &Claims{GivenName: "A"}
	in := `{"identity":{"traits":{"given_name":"B"}}}`

	out, err := applySberClaimsToMapperTraitsOutput("google", claims, in)
	require.NoError(t, err)
	require.Equal(t, in, out)
}

func TestApplySberClaimsToMapperTraitsOutputNestedTraits(t *testing.T) {
	t.Parallel()

	claims := &Claims{
		GivenName:  "Иван",
		FamilyName: "Петров",
		Email:      "User@Example.com",
	}

	in := `{"identity":{"traits":{"profile":{"given_name":"ИВАН","family_name":"ПЕТРОВ","email":"USER@EXAMPLE.COM"}}}}`

	out, err := applySberClaimsToMapperTraitsOutput("sber", claims, in)
	require.NoError(t, err)
	require.Contains(t, out, `"given_name":"Иван"`)
	require.Contains(t, out, `"family_name":"Петров"`)
	require.Contains(t, out, `"email":"User@Example.com"`)
}

func TestApplySberClaimsToIdentityTraitsBytesAfterMerge(t *testing.T) {
	t.Parallel()

	claims := &Claims{GivenName: "Иван", FamilyName: "Петров"}
	merged := []byte(`{"given_name":"ИВАН","family_name":"ПЕТРОВ","extra":"x"}`)

	out, err := applySberClaimsToIdentityTraitsBytes("sber-ift", claims, merged)
	require.NoError(t, err)
	require.Contains(t, string(out), `"given_name":"Иван"`)
	require.Contains(t, string(out), `"family_name":"Петров"`)
	require.Contains(t, string(out), `"extra":"x"`)
}

func TestApplySberClaimsFrontendTraitNames(t *testing.T) {
	t.Parallel()

	claims := &Claims{
		GivenName:   "Иван",
		FamilyName:  "Петров",
		Email:       "a@b.c",
		PhoneNumber: "+79001234567",
		Birthdate:   "1990-05-01",
		MiddleName:  "Сергеевич",
		Picture:     "https://cdn.example/avatar.jpg",
		City:        "Москва",
	}

	in := `{"identity":{"traits":{"first_name":"ИВАН","last_name":"ПЕТРОВ","middle_name":null,"email":"A@B.C","phone_number":"+79001234567","birth_date":"","avatar_url":null,"city":""}}}`

	out, err := applySberClaimsToMapperTraitsOutput("sber", claims, in)
	require.NoError(t, err)
	require.Contains(t, out, `"first_name":"Иван"`)
	require.Contains(t, out, `"last_name":"Петров"`)
	require.Contains(t, out, `"middle_name":"Сергеевич"`)
	require.Contains(t, out, `"email":"a@b.c"`)
	require.Contains(t, out, `"birth_date":"1990-05-01"`)
	require.Contains(t, out, `"avatar_url":"https://cdn.example/avatar.jpg"`)
	require.Contains(t, out, `"city":"Москва"`)
}
