//go:build !commercial

// Copyright © 2026 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package x

// Edition reports which build edition this binary was compiled as.
func Edition() string { return "oss" }
