// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package password_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/hash"
	"github.com/ory/kratos/pkg"
)

// TestLoginConfigMutationIsolation guards against Race B: a config-mutating
// subtest must not share its *configx.Provider with requests that read config
// concurrently. The reader path (Bcrypt.Generate -> Config().HasherBcrypt ->
// IsInsecureDevMode -> Provider.Bool) and the writer path (Config.MustSet ->
// Provider.Set -> replaceKoanf, swapping p.Koanf) operate on the same
// *configx.Provider when a registry is shared, which the race detector flags.
//
// The fix is that a config-mutating subtest owns its own registry+config, so
// reads on one registry never race writes on another. This test verifies that
// the per-registry isolation pattern is race-free: a reader keeps hashing on
// one registry while a writer mutates config on its own, separate registry.
func TestLoginConfigMutationIsolation(t *testing.T) {
	t.Parallel()

	// Reader owns its registry, mirroring an in-flight request that hashes a
	// password and reads IsInsecureDevMode on every call.
	_, readerReg := pkg.NewFastRegistryWithMocks(t)
	hasher := hash.NewHasherBcrypt(readerReg)

	// Writer owns a separate registry+config, mirroring a config-mutating
	// subtest that calls MustSet (and its cleanup), as required by
	// .claude/rules/go.md: "subtests that need parallelization must own their
	// resources".
	writerConf, _ := pkg.NewFastRegistryWithMocks(t)

	var eg errgroup.Group
	eg.Go(func() error {
		for range 200 {
			if _, err := hasher.Generate(t.Context(), []byte("a-long-enough-password")); err != nil {
				return err
			}
		}
		return nil
	})
	eg.Go(func() error {
		for range 200 {
			writerConf.MustSet(t.Context(), config.ViperKeyUseLegacyRequireVerifiedLoginError, true)
			writerConf.MustSet(t.Context(), config.ViperKeyUseLegacyRequireVerifiedLoginError, false)
		}
		return nil
	})
	require.NoError(t, eg.Wait())
}
