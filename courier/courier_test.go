// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package courier_test

import (
	"testing"

	"github.com/ory/kratos/x"
)

func TestMain(m *testing.M) {
	m.Run()
	x.CleanUpTestSMTP()
}
