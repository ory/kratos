// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package sql_test

import (
	"fmt"
	"net"
	"os/exec"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/pkg"
	"github.com/ory/kratos/pkg/testhelpers"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/recovery"
	"github.com/ory/kratos/selfservice/strategy/code"
	"github.com/ory/kratos/selfservice/strategy/link"
	"github.com/ory/kratos/x"
	"github.com/ory/pop/v6"
)

// TestOneTimeSecretSingleUseUnderReadCommitted is the regression test for
// HackerOne #3692149 and its duplicate #3695625.
//
// One-time secrets (codes and tokens) must be consumed atomically. On READ
// COMMITTED backends (self-hosted Postgres/MySQL) the consume previously did a
// SELECT (matching only unused rows) followed by an unguarded UPDATE. Two
// concurrent submissions of the same valid secret could both read the row as
// unused and both run the UPDATE, so both succeeded — a lost update that let an
// attacker mint duplicate sessions for the same identity.
//
// Postgres defaults to READ COMMITTED, so this test reproduces the race
// deterministically against a real Postgres instance using real production
// code: before the fix multiple concurrent consumers succeed, after the fix
// exactly one does. The test starts a throwaway local Postgres cluster and
// skips if the Postgres binaries are not installed.
func TestOneTimeSecretSingleUseUnderReadCommitted(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping Postgres-backed race test under -short")
	}

	dsn := startLocalPostgres(t)

	ctx := testhelpers.WithDefaultIdentitySchema(t.Context(), "file://./stub/identity.schema.json")
	_, reg := pkg.NewRegistryDefaultWithDSN(t, dsn)
	_, p := testhelpers.NewNetwork(t, ctx, reg.Persister())

	// assertSingleUse runs many concurrent consumers of the same one-time
	// secret and asserts that no more than one of them succeeds.
	assertSingleUse := func(t *testing.T, consume func() error) {
		const concurrency = 16
		var success int32
		var start sync.WaitGroup
		var done sync.WaitGroup
		start.Add(1)
		for range concurrency {
			done.Add(1)
			go func() {
				defer done.Done()
				start.Wait()
				if err := consume(); err == nil {
					atomic.AddInt32(&success, 1)
				}
			}()
		}
		start.Done()
		done.Wait()

		require.EqualValues(t, 1, atomic.LoadInt32(&success),
			"exactly one concurrent consumer may use the one-time secret")
	}

	newIdentityWithRecoveryAddress := func(t *testing.T, email string) *identity.Identity {
		var i identity.Identity
		i.Traits = identity.Traits(`{}`)
		i.RecoveryAddresses = append(i.RecoveryAddresses, identity.RecoveryAddress{
			Value: email, Via: identity.AddressTypeEmail, IdentityID: i.ID,
		})
		require.NoError(t, p.CreateIdentity(ctx, &i))
		return &i
	}

	newRecoveryFlow := func(t *testing.T) *recovery.Flow {
		f := &recovery.Flow{
			ID:        x.NewUUID(),
			Type:      flow.TypeBrowser,
			State:     flow.StateChooseMethod,
			ExpiresAt: time.Now().Add(time.Hour),
			IssuedAt:  time.Now(),
		}
		require.NoError(t, p.CreateRecoveryFlow(ctx, f))
		return f
	}

	t.Run("case=recovery code", func(t *testing.T) {
		f := newRecoveryFlow(t)
		i := newIdentityWithRecoveryAddress(t, "rc-race@ory.sh")
		dto := &code.CreateRecoveryCodeParams{
			RawCode:         "12345678",
			FlowID:          f.ID,
			RecoveryAddress: &i.RecoveryAddresses[0],
			ExpiresIn:       time.Hour,
			IdentityID:      i.ID,
		}
		_, err := p.CreateRecoveryCode(ctx, dto)
		require.NoError(t, err)

		assertSingleUse(t, func() error {
			_, err := p.UseRecoveryCode(ctx, f.ID, dto.RawCode)
			return err
		})
	})

	t.Run("case=recovery token", func(t *testing.T) {
		f := newRecoveryFlow(t)
		i := newIdentityWithRecoveryAddress(t, "rt-race@ory.sh")
		token := &link.RecoveryToken{
			Token:           x.NewUUID().String(),
			FlowID:          uuid.NullUUID{UUID: f.ID, Valid: true},
			RecoveryAddress: &i.RecoveryAddresses[0],
			ExpiresAt:       time.Now().Add(time.Hour),
			IssuedAt:        time.Now(),
			IdentityID:      i.ID,
			TokenType:       link.RecoveryTokenTypeAdmin,
		}
		require.NoError(t, p.CreateRecoveryToken(ctx, token))

		raw := token.Token
		assertSingleUse(t, func() error {
			_, err := p.UseRecoveryToken(ctx, f.ID, raw)
			return err
		})
	})
}

// startLocalPostgres boots a throwaway PostgreSQL cluster for the test and
// returns its DSN. It skips the test when the Postgres binaries are not
// available. Postgres defaults to READ COMMITTED isolation, which is required
// to reproduce the one-time-secret consume race.
func startLocalPostgres(t *testing.T) string {
	t.Helper()

	bin := lookupPostgresBin(t)
	dataDir := filepath.Join(t.TempDir(), "pgdata")
	port := freePort(t)

	run := func(name string, args ...string) {
		//#nosec G204 -- test helper runs trusted Postgres binaries from a fixed, test-controlled path.
		cmd := exec.Command(filepath.Join(bin, name), args...)
		out, err := cmd.CombinedOutput()
		require.NoErrorf(t, err, "%s failed: %s", name, string(out))
	}

	run("initdb", "-D", dataDir, "-U", "postgres", "--auth=trust", "-E", "UTF8")

	logFile := filepath.Join(t.TempDir(), "pg.log")
	run("pg_ctl", "-D", dataDir, "-l", logFile, "-w",
		"-o", fmt.Sprintf("-p %d -c listen_addresses=127.0.0.1 -c fsync=off", port),
		"start")
	t.Cleanup(func() {
		//#nosec G204 -- test helper runs trusted Postgres binaries from a fixed, test-controlled path.
		stop := exec.Command(filepath.Join(bin, "pg_ctl"), "-D", dataDir, "-m", "immediate", "stop")
		_ = stop.Run()
	})

	dsn := fmt.Sprintf("postgres://postgres@127.0.0.1:%d/postgres?sslmode=disable", port)

	// Wait for the server to accept connections.
	require.EventuallyWithT(t, func(t *assert.CollectT) {
		c, err := pop.NewConnection(&pop.ConnectionDetails{URL: dsn})
		if !assert.NoError(t, err) {
			return
		}
		assert.NoError(t, c.Open())
		assert.NoError(t, c.RawQuery("SELECT 1").Exec())
		_ = c.Close()
	}, 30*time.Second, 200*time.Millisecond)

	return dsn
}

func lookupPostgresBin(t *testing.T) string {
	t.Helper()
	// Prefer the Homebrew keg-only path, then fall back to PATH.
	for _, candidate := range []string{
		"/opt/homebrew/opt/postgresql@14/bin",
		"/usr/local/opt/postgresql@14/bin",
	} {
		if _, err := exec.LookPath(filepath.Join(candidate, "pg_ctl")); err == nil {
			return candidate
		}
	}
	p, err := exec.LookPath("pg_ctl")
	if err != nil {
		t.Skip("PostgreSQL binaries (pg_ctl/initdb) not found; skipping READ COMMITTED race test")
	}
	return filepath.Dir(p)
}

func freePort(t *testing.T) int {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer func() { _ = l.Close() }()
	return l.Addr().(*net.TCPAddr).Port
}
