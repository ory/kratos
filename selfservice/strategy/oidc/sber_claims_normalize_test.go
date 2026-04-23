// Copyright © 2026 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oidc

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNormalizeSberBirthdateISO8601(t *testing.T) {
	t.Parallel()

	require.Equal(t, "1990-05-01T00:00:00Z", normalizeSberBirthdateISO8601("1990-05-01"))
	require.Equal(t, "1990-05-01T00:00:00Z", normalizeSberBirthdateISO8601("01.05.1990"))
	require.Equal(t, "1990-05-01T00:00:00Z", normalizeSberBirthdateISO8601("01/05/1990"))
	require.Equal(t, "1990-05-01T10:20:30Z", normalizeSberBirthdateISO8601("1990-05-01T10:20:30Z"))
	require.Equal(t, "", normalizeSberBirthdateISO8601("bad-date"))
}

func TestNormalizeRussianMobilePlus79(t *testing.T) {
	t.Parallel()

	require.Equal(t, "+79991234567", normalizeRussianMobilePlus79("+7 (999) 123 45 67"))
}

func TestNormalizeNameTitle(t *testing.T) {
	t.Parallel()

	require.Equal(t, "Иван", normalizeNameTitle("ИВАН"))
	require.Equal(t, "Петров", normalizeNameTitle("пЕТРОВ"))
	require.Equal(t, "Сергеевич", normalizeNameTitle(" СЕРГЕЕВИЧ "))
	require.Equal(t, "", normalizeNameTitle(""))
}

func TestNormalizeEmailLower(t *testing.T) {
	t.Parallel()

	require.Equal(t, "user@example.com", normalizeEmailLower("User@Example.COM"))
	require.Equal(t, "a@b.c", normalizeEmailLower("  A@B.C  "))
	require.Equal(t, "", normalizeEmailLower("   "))
}
