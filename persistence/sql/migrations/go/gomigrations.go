// Copyright © 2025 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package gomigrations

import (
	"fmt"
	"path/filepath"
	"runtime"
	"slices"
)

var All = slices.Concat(
	backfillIdentityID,
)

func path() string {
	_, file, line, _ := runtime.Caller(1)
	return fmt.Sprintf("%s:%d", filepath.Base(file), line)
}
