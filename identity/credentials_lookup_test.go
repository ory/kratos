// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package identity_test

import (
	_ "embed"
	"testing"
	"time"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal/testhelpers"

	"github.com/ory/x/sqlxx"
)

func TestToNode(t *testing.T) {
	c := identity.CredentialsLookupConfig{RecoveryCodes: []identity.RecoveryCode{
		{Code: "foo", UsedAt: sqlxx.NullTime(time.Unix(1629199958, 0).UTC())},
		{Code: "bar"},
		{Code: "baz"},
		{Code: "oof", UsedAt: sqlxx.NullTime(time.Unix(1629199968, 0).UTC())},
	}}

	testhelpers.SnapshotTExcept(t, c.ToNode(), []string{})
}
