// Copyright © 2025 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package gomigrations

import (
	"embed"
	"fmt"
	"path/filepath"
	"runtime"
	"slices"
)

var (
	All = slices.Concat(
		backfillIdentityID,
	)

	//go:embed *.go
	Src embed.FS
)

func path() string {
	_, file, line, _ := runtime.Caller(1)
	return fmt.Sprintf("%s:%d", filepath.Base(file), line)
}
