// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package throttle

import (
	"sync"
	"time"

	"github.com/gofrs/uuid"
)

// Config holds the configuration for login throttling.
type Config struct {
	// MaxAttempts is the maximum number of failed login attempts
	// before the identity gets locked out temporarily.
	MaxAttempts int

	// ThrottleWindow is the time window within which failed attempts are counted.
	ThrottleWindow time.Duration

	// LockoutDuration is how long an identity is locked out after
	// exceeding MaxAttempts within the ThrottleWindow.
	LockoutDuration time.Duration
}

// DefaultConfig returns the default throttle configuration.
func DefaultConfig() Config {
	return Config{
		MaxAttempts:     5,
		ThrottleWindow:  5 * time.Minute,
		LockoutDuration: 15 * time.Minute,
	}
}

type attempt struct {
	timestamp time.Time
}

type identityState struct {
	attempts  []attempt
	lockedAt  time.Time
	lockUntil time.Time
}

// Limiter tracks failed login attempts per identity and enforces throttling.
type Limiter struct {
	mu     sync.RWMutex
	states map[uuid.UUID]*identityState
	config Config
}

// NewLimiter creates a new login throttle limiter.
func NewLimiter(cfg Config) *Limiter {
	l := &Limiter{
		states: make(map[uuid.UUID]*identityState),
		config: cfg,
	}
	return l
}

// IsLockedOut checks whether the given identity is currently locked out.
// Returns true and the remaining lockout duration if locked out.
func (l *Limiter) IsLockedOut(identityID uuid.UUID) (bool, time.Duration) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	state, ok := l.states[identityID]
	if !ok {
		return false, 0
	}

	if !state.lockUntil.IsZero() && time.Now().Before(state.lockUntil) {
		return true, time.Until(state.lockUntil)
	}

	return false, 0
}

// RecordFailure records a failed login attempt for the given identity.
// Returns true if the identity is now locked out as a result.
func (l *Limiter) RecordFailure(identityID uuid.UUID) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	state, ok := l.states[identityID]
	if !ok {
		state = &identityState{}
		l.states[identityID] = state
	}

	// If currently locked, don't reset anything
	if !state.lockUntil.IsZero() && time.Now().Before(state.lockUntil) {
		return true
	}

	// Prune old attempts outside the throttle window
	now := time.Now()
	cutoff := now.Add(-l.config.ThrottleWindow)
	pruned := make([]attempt, 0, len(state.attempts))
	for _, a := range state.attempts {
		if a.timestamp.After(cutoff) {
			pruned = append(pruned, a)
		}
	}

	// Add the new failure
	pruned = append(pruned, attempt{timestamp: now})
	state.attempts = pruned

	// Check if we should lock out
	if len(state.attempts) >= l.config.MaxAttempts {
		state.lockedAt = now
		state.lockUntil = now.Add(l.config.LockoutDuration)
		return true
	}

	return false
}

// RecordSuccess clears the failed attempt history for the given identity.
func (l *Limiter) RecordSuccess(identityID uuid.UUID) {
	l.mu.Lock()
	defer l.mu.Unlock()

	delete(l.states, identityID)
}

// RecentFailures returns the number of recent failed attempts for the identity
// within the throttle window.
func (l *Limiter) RecentFailures(identityID uuid.UUID) int {
	l.mu.RLock()
	defer l.mu.RUnlock()

	state, ok := l.states[identityID]
	if !ok {
		return 0
	}

	cutoff := time.Now().Add(-l.config.ThrottleWindow)
	count := 0
	for _, a := range state.attempts {
		if a.timestamp.After(cutoff) {
			count++
		}
	}
	return count
}

// Cleanup removes expired entries. Should be called periodically
// to prevent unbounded memory growth.
func (l *Limiter) Cleanup() {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	for id, state := range l.states {
		// Remove entries where lockout has expired and no recent attempts
		if !state.lockUntil.IsZero() && now.After(state.lockUntil) {
			delete(l.states, id)
			continue
		}

		// Remove entries with no recent attempts in the window
		cutoff := now.Add(-l.config.ThrottleWindow)
		hasRecent := false
		for _, a := range state.attempts {
			if a.timestamp.After(cutoff) {
				hasRecent = true
				break
			}
		}
		if !hasRecent && state.lockUntil.IsZero() {
			delete(l.states, id)
		}
	}
}
