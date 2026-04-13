// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package throttle_test

import (
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/selfservice/flow/login/throttle"
)

func TestLimiter_BasicLockout(t *testing.T) {
	cfg := throttle.Config{
		MaxAttempts:     3,
		ThrottleWindow:  1 * time.Minute,
		LockoutDuration: 5 * time.Minute,
	}
	l := throttle.NewLimiter(cfg)
	id := uuid.Must(uuid.NewV4())

	// Not locked out initially
	locked, _ := l.IsLockedOut(id)
	assert.False(t, locked)

	// Record failures below threshold
	assert.False(t, l.RecordFailure(id))
	assert.False(t, l.RecordFailure(id))
	assert.Equal(t, 2, l.RecentFailures(id))

	// Third failure triggers lockout
	assert.True(t, l.RecordFailure(id))
	locked, remaining := l.IsLockedOut(id)
	assert.True(t, locked)
	assert.True(t, remaining > 0)
}

func TestLimiter_SuccessResetsCounter(t *testing.T) {
	cfg := throttle.Config{
		MaxAttempts:     3,
		ThrottleWindow:  1 * time.Minute,
		LockoutDuration: 5 * time.Minute,
	}
	l := throttle.NewLimiter(cfg)
	id := uuid.Must(uuid.NewV4())

	l.RecordFailure(id)
	l.RecordFailure(id)
	assert.Equal(t, 2, l.RecentFailures(id))

	// Successful login clears the counter
	l.RecordSuccess(id)
	assert.Equal(t, 0, l.RecentFailures(id))

	locked, _ := l.IsLockedOut(id)
	assert.False(t, locked)
}

func TestLimiter_DifferentIdentities(t *testing.T) {
	cfg := throttle.Config{
		MaxAttempts:     2,
		ThrottleWindow:  1 * time.Minute,
		LockoutDuration: 5 * time.Minute,
	}
	l := throttle.NewLimiter(cfg)

	id1 := uuid.Must(uuid.NewV4())
	id2 := uuid.Must(uuid.NewV4())

	l.RecordFailure(id1)
	l.RecordFailure(id1)

	// id1 should be locked
	locked, _ := l.IsLockedOut(id1)
	assert.True(t, locked)

	// id2 should NOT be affected
	locked, _ = l.IsLockedOut(id2)
	assert.False(t, locked)
	assert.Equal(t, 0, l.RecentFailures(id2))
}

func TestLimiter_UnknownIdentity(t *testing.T) {
	l := throttle.NewLimiter(throttle.DefaultConfig())
	id := uuid.Must(uuid.NewV4())

	locked, remaining := l.IsLockedOut(id)
	assert.False(t, locked)
	assert.Equal(t, time.Duration(0), remaining)
	assert.Equal(t, 0, l.RecentFailures(id))
}

func TestLimiter_Cleanup(t *testing.T) {
	cfg := throttle.Config{
		MaxAttempts:     3,
		ThrottleWindow:  10 * time.Millisecond,
		LockoutDuration: 10 * time.Millisecond,
	}
	l := throttle.NewLimiter(cfg)
	id := uuid.Must(uuid.NewV4())

	l.RecordFailure(id)
	l.RecordFailure(id)
	l.RecordFailure(id)

	locked, _ := l.IsLockedOut(id)
	require.True(t, locked)

	// Wait for lockout and window to expire
	time.Sleep(20 * time.Millisecond)

	l.Cleanup()

	// Should be cleaned up
	locked, _ = l.IsLockedOut(id)
	assert.False(t, locked)
	assert.Equal(t, 0, l.RecentFailures(id))
}

func TestDefaultConfig(t *testing.T) {
	cfg := throttle.DefaultConfig()
	assert.Equal(t, 5, cfg.MaxAttempts)
	assert.Equal(t, 5*time.Minute, cfg.ThrottleWindow)
	assert.Equal(t, 15*time.Minute, cfg.LockoutDuration)
}
