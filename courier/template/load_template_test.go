package template_test

import (
	"context"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/julienschmidt/httprouter"

	"github.com/ory/kratos/courier/template"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/internal"
	"github.com/ory/x/fetcher"

	lru "github.com/hashicorp/golang-lru"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/x"
)

func TestLoadTextTemplate(t *testing.T) {
	var executeTextTemplate = func(t *testing.T, dir, name, pattern string, model map[string]interface{}) string {
		ctx := context.Background()
		_, reg := internal.NewFastRegistryWithMocks(t)
		tp, err := template.LoadText(ctx, reg, os.DirFS(dir), name, pattern, model, "")
		require.NoError(t, err)
		return tp
	}

	var executeHTMLTemplate = func(t *testing.T, dir, name, pattern string, model map[string]interface{}) string {
		ctx := context.Background()
		_, reg := internal.NewFastRegistryWithMocks(t)
		tp, err := template.LoadHTML(ctx, reg, os.DirFS(dir), name, pattern, model, "")
		require.NoError(t, err)
		return tp
	}

	t.Run("method=from bundled", func(t *testing.T) {
		actual := executeTextTemplate(t, "courier/builtin/templates/test_stub", "email.body.gotmpl", "", nil)
		assert.Contains(t, actual, "stub email")
	})

	t.Run("method=fallback to bundled", func(t *testing.T) {
		template.Cache, _ = lru.New(16) // prevent Cache hit
		actual := executeTextTemplate(t, "some/inexistent/dir", "test_stub/email.body.gotmpl", "", nil)
		assert.Contains(t, actual, "stub email")
	})

	t.Run("method=with Sprig functions", func(t *testing.T) {
		template.Cache, _ = lru.New(16)                     // prevent Cache hit
		m := map[string]interface{}{"input": "hello world"} // create a simple model
		actual := executeTextTemplate(t, "courier/builtin/templates/test_stub", "email.body.sprig.gotmpl", "", m)
		assert.Contains(t, actual, "HelloWorld,HELLOWORLD")
	})

	t.Run("method=sprig should not support non-hermetic", func(t *testing.T) {
		template.Cache, _ = lru.New(16)
		ctx := context.Background()
		_, reg := internal.NewFastRegistryWithMocks(t)

		nonhermetic := []string{"date", "date_in_zone", "date_modify", "now", "htmlDate", "htmlDateInZone", "dateInZone", "dateModify", "env", "expandenv", "getHostByName", "uuidv4", "randNumeric", "randAscii", "randAlpha", "randAlphaNum"}

		for _, tc := range nonhermetic {
			t.Run("case=should not support function: "+tc, func(t *testing.T) {
				_, err := template.LoadText(ctx, reg, x.NewStubFS(tc, []byte(fmt.Sprintf("{{ %s }}", tc))), tc, "", map[string]interface{}{}, "")
				require.Error(t, err)
				require.Contains(t, err.Error(), fmt.Sprintf("function \"%s\" not defined", tc))
			})
		}
	})

	t.Run("method=html with nested templates", func(t *testing.T) {
		template.Cache, _ = lru.New(16)              // prevent Cache hit
		m := map[string]interface{}{"lang": "en_US"} // create a simple model
		actual := executeHTMLTemplate(t, "courier/builtin/templates/test_stub", "email.body.html.gotmpl", "email.body.html*", m)
		assert.Contains(t, actual, "lang=en_US")
	})

	t.Run("method=Cache works", func(t *testing.T) {
		dir := os.TempDir()
		name := x.NewUUID().String() + ".body.gotmpl"
		fp := filepath.Join(dir, name)

		require.NoError(t, os.WriteFile(fp, []byte("cached stub body"), 0666))
		assert.Contains(t, executeTextTemplate(t, dir, name, "", nil), "cached stub body")

		require.NoError(t, os.RemoveAll(fp))
		assert.Contains(t, executeTextTemplate(t, dir, name, "", nil), "cached stub body")
	})

	t.Run("method=remote resource", func(t *testing.T) {
		_, reg := internal.NewFastRegistryWithMocks(t)

		ctx, cancel := context.WithCancel(context.Background())
		t.Cleanup(cancel)

		t.Run("case=base64 encoded data", func(t *testing.T) {
			t.Run("html template", func(t *testing.T) {
				m := map[string]interface{}{"lang": "en_US"}
				f, err := ioutil.ReadFile("courier/builtin/templates/test_stub/email.body.html.en_US.gotmpl")
				require.NoError(t, err)
				b64 := base64.StdEncoding.EncodeToString(f)
				tp, err := template.LoadHTML(ctx, reg, nil, "", "", m, "base64://"+b64)
				require.NoError(t, err)
				assert.Contains(t, tp, "lang=en_US")
			})

			t.Run("case=plaintext", func(t *testing.T) {
				m := map[string]interface{}{"Body": "something"}
				f, err := ioutil.ReadFile("courier/builtin/templates/test_stub/email.body.plaintext.gotmpl")
				require.NoError(t, err)

				b64 := base64.StdEncoding.EncodeToString(f)

				tp, err := template.LoadText(ctx, reg, nil, "", "", m, "base64://"+b64)
				require.NoError(t, err)
				assert.Contains(t, tp, "stub email body something")
			})

		})

		t.Run("case=file resource", func(t *testing.T) {
			t.Run("case=html template", func(t *testing.T) {
				m := map[string]interface{}{"lang": "en_US"}
				tp, err := template.LoadHTML(ctx, reg, nil, "", "", m, "file://courier/builtin/templates/test_stub/email.body.html.en_US.gotmpl")
				require.NoError(t, err)
				assert.Contains(t, tp, "lang=en_US")
			})

			t.Run("case=plaintext", func(t *testing.T) {
				m := map[string]interface{}{"Body": "something"}
				tp, err := template.LoadText(ctx, reg, nil, "", "", m, "file://courier/builtin/templates/test_stub/email.body.plaintext.gotmpl")
				require.NoError(t, err)
				assert.Contains(t, tp, "stub email body something")
			})
		})

		t.Run("case=http resource", func(t *testing.T) {
			router := httprouter.New()
			router.Handle("GET", "/html", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
				http.ServeFile(writer, request, "courier/builtin/templates/test_stub/email.body.html.en_US.gotmpl")
			})
			router.Handle("GET", "/plaintext", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
				http.ServeFile(writer, request, "courier/builtin/templates/test_stub/email.body.plaintext.gotmpl")
			})
			ts := httptest.NewServer(router)
			defer ts.Close()

			t.Run("case=html template", func(t *testing.T) {
				m := map[string]interface{}{"lang": "en_US"}
				tp, err := template.LoadHTML(ctx, reg, nil, "", "", m, ts.URL+"/html")
				require.NoError(t, err)
				assert.Contains(t, tp, "lang=en_US")
			})

			t.Run("case=plaintext", func(t *testing.T) {
				m := map[string]interface{}{"Body": "something"}
				tp, err := template.LoadText(ctx, reg, nil, "", "", m, ts.URL+"/plaintext")
				require.NoError(t, err)
				assert.Contains(t, tp, "stub email body something")
			})

		})

		t.Run("case=unsupported resource", func(t *testing.T) {
			tp, err := template.LoadHTML(ctx, reg, nil, "", "", map[string]interface{}{}, "grpc://unsupported-url")

			require.ErrorIs(t, err, fetcher.ErrUnknownScheme)
			require.Empty(t, tp)

			tp, err = template.LoadText(ctx, reg, nil, "", "", map[string]interface{}{}, "grpc://unsupported-url")
			require.ErrorIs(t, err, fetcher.ErrUnknownScheme)
			require.Empty(t, tp)
		})

		t.Run("case=disallowed resources", func(t *testing.T) {
			require.NoError(t, reg.Config(ctx).Source().Set(config.ViperKeyClientHTTPNoPrivateIPRanges, true))
			reg.HTTPClient(ctx).RetryMax = 1
			reg.HTTPClient(ctx).RetryWaitMax = time.Millisecond

			_, err := template.LoadHTML(ctx, reg, nil, "", "", map[string]interface{}{}, "http://localhost:8080/1234")

			require.Error(t, err)
			assert.Contains(t, err.Error(), "is in the")

			_, err = template.LoadText(ctx, reg, nil, "", "", map[string]interface{}{}, "http://localhost:8080/1234")
			require.Error(t, err)
			assert.Contains(t, err.Error(), "is in the")

		})

		t.Run("method=cache works", func(t *testing.T) {
			tp1, err := template.LoadText(ctx, reg, nil, "", "", map[string]interface{}{}, "base64://e3sgJGwgOj0gY2F0ICJsYW5nPSIgLmxhbmcgfX0Ke3sgbm9zcGFjZSAkbCB9fQ==")
			assert.NoError(t, err)

			tp2, err := template.LoadText(ctx, reg, nil, "", "", map[string]interface{}{}, "base64://c3R1YiBlbWFpbCBib2R5IHt7IC5Cb2R5IH19")
			assert.NoError(t, err)

			require.NotEqualf(t, tp1, tp2, "Expected remote template 1 and remote template 2 to not be equal")
		})

	})
}
