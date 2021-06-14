package template

import (
	"os"
	"path/filepath"
	"testing"

	lru "github.com/hashicorp/golang-lru"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/x"
)

func TestLoadTextTemplate(t *testing.T) {
	var executeTemplate = func(t *testing.T, dir, name string) string {
		tp, err := loadTextTemplate(dir, name, nil)
		require.NoError(t, err)
		return tp
	}

	t.Run("method=from bundled", func(t *testing.T) {
		actual := executeTemplate(t, "courier/builtin/templates", "test_stub/email.body.gotmpl")
		assert.Contains(t, actual, "stub email")
	})

	t.Run("method=fallback to bundled", func(t *testing.T) {
		cache, _ = lru.New(16) // prevent cache hit
		actual := executeTemplate(t, "some/inexistent/dir", "test_stub/email.body.gotmpl")
		assert.Contains(t, actual, "stub email")
	})

	t.Run("method=cache works", func(t *testing.T) {
		dir := os.TempDir()
		name := x.NewUUID().String() + ".body.gotmpl"
		fp := filepath.Join(dir, name)

		require.NoError(t, os.WriteFile(fp, []byte("cached stub body"), 0666))
		assert.Contains(t, executeTemplate(t, dir, name), "cached stub body")

		require.NoError(t, os.RemoveAll(fp))
		assert.Contains(t, executeTemplate(t, dir, name), "cached stub body")
	})
}
