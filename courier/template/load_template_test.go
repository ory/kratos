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
	var executeTextTemplate = func(t *testing.T, dir, name, pattern string, model map[string]interface{}) string {
		tp, err := LoadTextTemplate(dir, name, pattern, model)
		require.NoError(t, err)
		return tp
	}

	var executeHTMLTemplate = func(t *testing.T, dir, name, pattern string, model map[string]interface{}) string {
		tp, err := LoadHTMLTemplate(dir, name, pattern, model)
		require.NoError(t, err)
		return tp
	}

	t.Run("method=from bundled", func(t *testing.T) {
		actual := executeTextTemplate(t, "courier/builtin/templates/test_stub", "email.body.gotmpl", "", nil)
		assert.Contains(t, actual, "stub email")
	})

	t.Run("method=fallback to bundled", func(t *testing.T) {
		cache, _ = lru.New(16) // prevent cache hit
		actual := executeTextTemplate(t, "some/inexistent/dir", "test_stub/email.body.gotmpl", "", nil)
		assert.Contains(t, actual, "stub email")
	})

	t.Run("method=with Sprig functions", func(t *testing.T) {
		cache, _ = lru.New(16)                              // prevent cache hit
		m := map[string]interface{}{"input": "hello world"} // create a simple model
		actual := executeTextTemplate(t, "courier/builtin/templates/test_stub", "email.body.sprig.gotmpl", "", m)
		assert.Contains(t, actual, "HelloWorld,HELLOWORLD")
	})

	t.Run("method=html with nested templates", func(t *testing.T) {
		cache, _ = lru.New(16)                       // prevent cache hit
		m := map[string]interface{}{"lang": "en_US"} // create a simple model
		actual := executeHTMLTemplate(t, "courier/builtin/templates/test_stub", "email.body.html.gotmpl", "email.body.html*", m)
		assert.Contains(t, actual, "lang=en_US")
	})

	t.Run("method=cache works", func(t *testing.T) {
		dir := os.TempDir()
		name := x.NewUUID().String() + ".body.gotmpl"
		fp := filepath.Join(dir, name)

		require.NoError(t, os.WriteFile(fp, []byte("cached stub body"), 0666))
		assert.Contains(t, executeTextTemplate(t, dir, name, "", nil), "cached stub body")

		require.NoError(t, os.RemoveAll(fp))
		assert.Contains(t, executeTextTemplate(t, dir, name, "", nil), "cached stub body")
	})
}
