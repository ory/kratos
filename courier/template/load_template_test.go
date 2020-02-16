package template

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/shurcooL/go/ioutil"
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
		actual := executeTemplate(t, "test_stub/email.body.gotmpl")
		assert.Contains(t, actual, "stub email")
	})

	t.Run("method=cache works", func(t *testing.T) {
		fp := filepath.Join(os.TempDir(), x.NewUUID().String()) + ".body.gotmpl"
		require.NoError(t, ioutil.WriteFile(fp, bytes.NewBufferString("cached stub body")))
		assert.Contains(t, executeTemplate(t, fp), "cached stub body")

		require.NoError(t, os.RemoveAll(fp))
		assert.Contains(t, executeTemplate(t, fp), "cached stub body")
	})
}
