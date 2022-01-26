package template_test

import (
	"encoding/base64"
	"github.com/julienschmidt/httprouter"
	"github.com/ory/kratos/driver/config"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/courier/template"
	"github.com/ory/kratos/internal"
)

func TestRecoverInvalid(t *testing.T) {

	t.Run("test=with courier templates directory", func(t *testing.T) {
		conf, _ := internal.NewFastRegistryWithMocks(t)
		tpl := template.NewRecoveryInvalid(conf, &template.RecoveryInvalidModel{})

		rendered, err := tpl.EmailBody()
		require.NoError(t, err)
		assert.NotEmpty(t, rendered)

		rendered, err = tpl.EmailSubject()
		require.NoError(t, err)
		assert.NotEmpty(t, rendered)
	})

	t.Run("test=with remote resources", func(t *testing.T) {
		setup := func(t *testing.T, plaintext string, html string, subject string) *config.Config {
			conf, _ := internal.NewFastRegistryWithMocks(t)
			require.NoError(t, conf.Set(config.ViperKeyCourierTemplatesRecoveryInvalid, &config.CourierEmailTemplate{
				TemplateRoot: "",
				Body: &config.CourierEmailBodyTemplate{
					PlainText: plaintext,
					HTML:      html,
				},
				Subject: subject,
			}))
			return conf
		}

		testRendered := func(t *testing.T, conf *config.Config) {
			tpl := template.NewRecoveryInvalid(conf, &template.RecoveryInvalidModel{})

			rendered, err := tpl.EmailBody()
			require.NoError(t, err)
			assert.NotEmpty(t, rendered)

			rendered, err = tpl.EmailSubject()
			require.NoError(t, err)
			assert.NotEmpty(t, rendered)
		}

		t.Run("case=http resource", func(t *testing.T) {
			router := httprouter.New()
			router.Handle("GET", "/email.body.plaintext.gotpml", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
				http.ServeFile(writer, request, "courier/builtin/templates/recovery/invalid/email.body.plaintext.gotmpl")
			})
			router.Handle("GET", "/plaintext", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
				http.ServeFile(writer, request, "courier/builtin/templates/recovery/invalid/email.body.gotmpl")
			})
			ts := httptest.NewServer(router)
			defer ts.Close()

			testRendered(t, setup(t,
				ts.URL+"/email.body.plaintext.gotmpl",
				ts.URL+"/email.body.gotmpl",
				ts.URL+"/email.subject.gotmpl"))
		})

		t.Run("case=base64 resource", func(t *testing.T) {
			baseUrl := "courier/builtin/templates/recovery/invalid/"

			toBase64 := func(filePath string) string {
				f, err := ioutil.ReadFile(filePath)
				require.NoError(t, err)
				return base64.StdEncoding.EncodeToString(f)
			}

			testRendered(t, setup(t,
				toBase64(baseUrl+"email.body.plaintext.gotmpl"),
				toBase64(baseUrl+"email.body.gotmpl"),
				toBase64(baseUrl+"email.subject.gotmpl")))
		})

		t.Run("case=file resource", func(t *testing.T) {
			baseUrl := "file://courier/builtin/templates/recovery/invalid/"

			testRendered(t, setup(t,
				baseUrl+"email.body.plaintext.gotmpl",
				baseUrl+"email.body.gotmpl",
				baseUrl+"email.subject.gotmpl"))
		})
	})
}
