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
	cmd.Execute()
	out, err := io.ReadAll(b)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(out), "expected to get the DSN as an argument") {
		t.Fatalf("expected \"%s\" got \"%s\"", "expected to get the DSN as an argument", string(out))
	}
	cmd.Execute()
}
