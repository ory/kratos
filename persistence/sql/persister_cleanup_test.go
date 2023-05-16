// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package sql_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/ory/kratos/internal"
)

func TestPersister_Cleanup(t *testing.T) {
	_, reg := internal.NewFastRegistryWithMocks(t)
	p := reg.Persister()
	ctx := context.Background()

	t.Run("case=should not throw error on cleanup", func(t *testing.T) {
		assert.Nil(t, p.CleanupDatabase(ctx, 0, 0, reg.Config().DatabaseCleanupBatchSize(ctx)))
	})

	t.Run("case=should throw error on cleanup", func(t *testing.T) {
		p.GetConnection(ctx).Close()
		assert.Error(t, p.CleanupDatabase(ctx, 0, 0, reg.Config().DatabaseCleanupBatchSize(ctx)))
	})
}

func TestPersister_Continuity_Cleanup(t *testing.T) {
	_, reg := internal.NewFastRegistryWithMocks(t)
	p := reg.Persister()
	currentTime := time.Now()
	ctx := context.Background()

	t.Run("case=should not throw error on cleanup continuity sessions", func(t *testing.T) {
		assert.Nil(t, p.DeleteExpiredContinuitySessions(ctx, currentTime, reg.Config().DatabaseCleanupBatchSize(ctx)))
	})

	t.Run("case=should throw error on cleanup continuity sessions", func(t *testing.T) {
		p.GetConnection(ctx).Close()
		assert.Error(t, p.DeleteExpiredContinuitySessions(ctx, currentTime, reg.Config().DatabaseCleanupBatchSize(ctx)))
	})
}

func TestPersister_Login_Cleanup(t *testing.T) {
	_, reg := internal.NewFastRegistryWithMocks(t)
	p := reg.Persister()
	currentTime := time.Now()
	ctx := context.Background()

	t.Run("case=should not throw error on cleanup login flows", func(t *testing.T) {
		assert.Nil(t, p.DeleteExpiredLoginFlows(ctx, currentTime, reg.Config().DatabaseCleanupBatchSize(ctx)))
	})

	t.Run("case=should throw error on cleanup login flows", func(t *testing.T) {
		p.GetConnection(ctx).Close()
		assert.Error(t, p.DeleteExpiredLoginFlows(ctx, currentTime, reg.Config().DatabaseCleanupBatchSize(ctx)))
	})
}

func TestPersister_Recovery_Cleanup(t *testing.T) {
	_, reg := internal.NewFastRegistryWithMocks(t)
	p := reg.Persister()
	currentTime := time.Now()
	ctx := context.Background()

	t.Run("case=should not throw error on cleanup recovery flows", func(t *testing.T) {
		assert.Nil(t, p.DeleteExpiredRecoveryFlows(ctx, currentTime, reg.Config().DatabaseCleanupBatchSize(ctx)))
	})

	t.Run("case=should throw error on cleanup recovery flows", func(t *testing.T) {
		p.GetConnection(ctx).Close()
		assert.Error(t, p.DeleteExpiredRecoveryFlows(ctx, currentTime, reg.Config().DatabaseCleanupBatchSize(ctx)))
	})
}

func TestPersister_Registration_Cleanup(t *testing.T) {
	_, reg := internal.NewFastRegistryWithMocks(t)
	p := reg.Persister()
	currentTime := time.Now()
	ctx := context.Background()

	t.Run("case=should not throw error on cleanup registration flows", func(t *testing.T) {
		assert.Nil(t, p.DeleteExpiredRegistrationFlows(ctx, currentTime, reg.Config().DatabaseCleanupBatchSize(ctx)))
	})

	t.Run("case=should throw error on cleanup registration flows", func(t *testing.T) {
		p.GetConnection(ctx).Close()
		assert.Error(t, p.DeleteExpiredRegistrationFlows(ctx, currentTime, reg.Config().DatabaseCleanupBatchSize(ctx)))
	})
}

func TestPersister_Session_Cleanup(t *testing.T) {
	_, reg := internal.NewFastRegistryWithMocks(t)
	p := reg.Persister()
	currentTime := time.Now()
	ctx := context.Background()

	t.Run("case=should not throw error on cleanup sessions", func(t *testing.T) {
		assert.Nil(t, p.DeleteExpiredSessions(ctx, currentTime, reg.Config().DatabaseCleanupBatchSize(ctx)))
	})

	t.Run("case=should throw error on cleanup sessions", func(t *testing.T) {
		p.GetConnection(ctx).Close()
		assert.Error(t, p.DeleteExpiredSessions(ctx, currentTime, reg.Config().DatabaseCleanupBatchSize(ctx)))
	})
}

func TestPersister_Settings_Cleanup(t *testing.T) {
	_, reg := internal.NewFastRegistryWithMocks(t)
	p := reg.Persister()
	currentTime := time.Now()
	ctx := context.Background()

	t.Run("case=should not throw error on cleanup setting flows", func(t *testing.T) {
		assert.Nil(t, p.DeleteExpiredSettingsFlows(ctx, currentTime, reg.Config().DatabaseCleanupBatchSize(ctx)))
	})

	t.Run("case=should throw error on cleanup setting flows", func(t *testing.T) {
		p.GetConnection(ctx).Close()
		assert.Error(t, p.DeleteExpiredSettingsFlows(ctx, currentTime, reg.Config().DatabaseCleanupBatchSize(ctx)))
	})
}

func TestPersister_Verification_Cleanup(t *testing.T) {
	_, reg := internal.NewFastRegistryWithMocks(t)
	p := reg.Persister()
	currentTime := time.Now()
	ctx := context.Background()

	t.Run("case=should not throw error on cleanup verification flows", func(t *testing.T) {
		assert.Nil(t, p.DeleteExpiredVerificationFlows(ctx, currentTime, reg.Config().DatabaseCleanupBatchSize(ctx)))
	})

	t.Run("case=should throw error on cleanup verification flows", func(t *testing.T) {
		p.GetConnection(ctx).Close()
		assert.Error(t, p.DeleteExpiredVerificationFlows(ctx, currentTime, reg.Config().DatabaseCleanupBatchSize(ctx)))
	})
}

func TestPersister_SessionTokenExchange_Cleanup(t *testing.T) {
	_, reg := internal.NewFastRegistryWithMocks(t)
	p := reg.Persister()
	currentTime := time.Now()
	ctx := context.Background()

	t.Run("case=should not throw error on cleanup session token exchangers", func(t *testing.T) {
		assert.Nil(t, p.DeleteExpiredExchangers(ctx, currentTime, reg.Config().DatabaseCleanupBatchSize(ctx)))
	})

	t.Run("case=should throw error on cleanup session token exchangers if DB is closed", func(t *testing.T) {
		p.GetConnection(ctx).Close()
		assert.Error(t, p.DeleteExpiredExchangers(ctx, currentTime, reg.Config().DatabaseCleanupBatchSize(ctx)))
	})
}
