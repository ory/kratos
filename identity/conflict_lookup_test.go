// Copyright © 2026 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package identity

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/pop/v6"
	"github.com/ory/x/popx"
	"github.com/ory/x/sqlcon"
)

func TestFindFirstConflictReturnsLowestIndexMatch(t *testing.T) {
	match, err := findFirstConflict(t.Context(), []conflictLookup{
		func(ctx context.Context) (*conflictMatch, error) {
			return nil, sqlcon.ErrNoRows()
		},
		func(ctx context.Context) (*conflictMatch, error) {
			return &conflictMatch{address: "lookup-01", addressType: AddressTypeEmail}, nil
		},
		func(ctx context.Context) (*conflictMatch, error) {
			return &conflictMatch{address: "lookup-02", addressType: CredentialsTypePassword.String()}, nil
		},
	})

	require.NoError(t, err)
	require.NotNil(t, match)
	assert.Equal(t, "lookup-01", match.address)
}

func TestFindFirstConflictReturnsNoRowsWhenNothingMatches(t *testing.T) {
	var calls atomic.Int32
	var lookups []conflictLookup
	for range 20 {
		lookups = append(lookups, func(ctx context.Context) (*conflictMatch, error) {
			calls.Add(1)
			return nil, sqlcon.ErrNoRows()
		})
	}

	_, err := findFirstConflict(t.Context(), lookups)

	assert.ErrorIs(t, err, sqlcon.ErrNoRows())
	assert.EqualValues(t, 20, calls.Load())
}

func TestFindFirstConflictBoundsConcurrency(t *testing.T) {
	const lookupCount = 20
	started := make(chan struct{}, lookupCount)
	release := make(chan struct{})
	var active atomic.Int32

	var lookups []conflictLookup
	for range lookupCount {
		lookups = append(lookups, func(ctx context.Context) (*conflictMatch, error) {
			if n := active.Add(1); n > maxParallelConflictLookups {
				t.Errorf("%d lookups ran concurrently although the limit is %d", n, maxParallelConflictLookups)
			}
			defer active.Add(-1)
			started <- struct{}{}
			select {
			case <-release:
			case <-ctx.Done():
				return nil, ctx.Err()
			}
			return nil, sqlcon.ErrNoRows()
		})
	}

	done := make(chan error, 1)
	go func() {
		_, err := findFirstConflict(t.Context(), lookups)
		done <- err
	}()

	for range maxParallelConflictLookups {
		select {
		case <-started:
		case <-time.After(time.Second):
			t.Fatal("timed out waiting for the first lookups to start")
		}
	}

	select {
	case <-started:
		t.Fatal("a lookup started although the concurrency limit was reached")
	case <-time.After(20 * time.Millisecond):
	}

	close(release)
	require.ErrorIs(t, <-done, sqlcon.ErrNoRows())
}

func TestFindFirstConflictInTransactionShortCircuits(t *testing.T) {
	ctx := popx.WithTransaction(t.Context(), &pop.Connection{})
	require.True(t, popx.InTransaction(ctx))

	var calls []string
	record := func(name string, match bool) conflictLookup {
		return func(ctx context.Context) (*conflictMatch, error) {
			calls = append(calls, name)
			if match {
				return &conflictMatch{address: name}, nil
			}
			return nil, sqlcon.ErrNoRows()
		}
	}

	match, err := findFirstConflict(ctx, []conflictLookup{
		record("lookup-00", false),
		record("lookup-01", true),
		record("lookup-02", true),
	})

	require.NoError(t, err)
	require.NotNil(t, match)
	assert.Equal(t, "lookup-01", match.address)
	assert.Equal(t, []string{"lookup-00", "lookup-01"}, calls)
}
