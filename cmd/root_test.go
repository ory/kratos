// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"testing"

	"github.com/ory/x/cmdx"
)

func TestUsageStrings(t *testing.T) {
	cmdx.AssertUsageTemplates(t, NewRootCmd())
}
