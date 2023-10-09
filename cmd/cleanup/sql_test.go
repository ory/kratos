// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package cleanup

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

func Test_ExecuteCleanupFailedDSN(t *testing.T) {
	cmd := NewCleanupSQLCmd()
	b := bytes.NewBufferString("")
	cmd.SetOut(b)
	cmd.SetArgs([]string{"--read-from-env=false"})
	_ = cmd.Execute()
	out, err := io.ReadAll(b)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(out), "expected to get the DSN as an argument") {
		t.Fatalf("expected \"%s\" got \"%s\"", "expected to get the DSN as an argument", string(out))
	}
	_ = cmd.Execute()
}
