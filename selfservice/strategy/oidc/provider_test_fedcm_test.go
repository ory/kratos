// Copyright © 2025 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oidc_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/selfservice/strategy/oidc"
)

func TestFedcmTestProvider(t *testing.T) {
	_, reg := internal.NewVeryFastRegistryWithoutDB(t)

	p := oidc.NewProviderTestFedcm(&oidc.Configuration{}, reg)

	rawToken := `eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiI1NWVlMjgxNC02ZTQ4LTRmZTktYWIzNS1mM2QxYzczM2I3ZTciLCJub25jZSI6ImVkOWM0ZDcyMDZkMDc1YTg4NjY0ZmE3YjMwY2Q5ZGE2NGU4ZTkwMjY5MGJhZmI2YjNmMmY2OWU5YzU1ZGUyNTcwOTFlYTk3ZTFiZTFiYjdiNDZmMjJjYzY0ZSIsImV4cCI6MTczNzU1ODM4MTk3MSwiaWF0IjoxNzM3NDcxOTgxOTcxLCJlbWFpbCI6InhweGN3dnU1YjRuemZvdGZAZXhhbXBsZS5jb20iLCJuYW1lIjoiVXNlciBOYW1lIiwicGljdHVyZSI6Imh0dHBzOi8vYXBpLmRpY2ViZWFyLmNvbS83LngvYm90dHRzL3BuZz9zZWVkPSUyNDJiJTI0MTAlMjR5WEs3eWozNEg4SkhCNm8zOG1sc2xlYzl1WkozZ2F2UGlDaFdaeFFIbnk3VkFKRlouS3RGZSJ9.GnSP_x8J_yS5wrTwtB6B-BydYYljrpVjQjS2vZ5D8Hg` // #nosec G101 -- test code

	claims, err := p.(oidc.IDTokenVerifier).Verify(context.Background(), rawToken)
	require.NoError(t, err)
	require.NotNil(t, claims)
}
