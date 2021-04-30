package template

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/x"
)

func TestLoadTextTemplate(t *testing.T) {
	var executeTemplate = func(t *testing.T, path string) string {
		tp, err := loadTextTemplate(path, nil)
		require.NoError(t, err)
		return tp
	}

	t.Run("method=from bundled", func(t *testing.T) {
		actual := executeTemplate(t, "courier/builtin/templates/test_stub/email.body.gotmpl")
		assert.Contains(t, actual, "stub email")
	})

	t.Run("method=cache works", func(t *testing.T) {
		fp := filepath.Join(os.TempDir(), x.NewUUID().String()) + ".body.gotmpl"

		require.NoError(t, os.WriteFile(fp, []byte("cached stub body"), 0666))
		assert.Contains(t, executeTemplate(t, fp), "cached stub body")

		require.NoError(t, os.RemoveAll(fp))
		assert.Contains(t, executeTemplate(t, fp), "cached stub body")
	})
}
