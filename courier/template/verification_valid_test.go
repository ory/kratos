package template_test

import (
	"encoding/base64"
	"github.com/julienschmidt/httprouter"
	"github.com/ory/kratos/courier/template/testhelpers"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/courier/template"
	"github.com/ory/kratos/internal"
)

func TestVerifyValid(t *testing.T) {

	t.Run("test=with courier templates directory", func(t *testing.T) {
		conf, _ := internal.NewFastRegistryWithMocks(t)
		tpl := template.NewVerificationValid(conf, &template.VerificationValidModel{})

		testhelpers.TestRendered(t, tpl)
	})

	t.Run("test=with remote resources", func(t *testing.T) {
		t.Run("case=http resource", func(t *testing.T) {
			router := httprouter.New()
			router.Handle("GET", "/email.body.plaintext.gotpml", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
				http.ServeFile(writer, request, "courier/builtin/templates/verification/valid/email.body.plaintext.gotmpl")
			})
			router.Handle("GET", "/plaintext", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
				http.ServeFile(writer, request, "courier/builtin/templates/verification/valid/email.body.gotmpl")
			})
			ts := httptest.NewServer(router)
			defer ts.Close()

			tpl := template.NewVerificationValid(testhelpers.SetupRemoteConfig(t,
				ts.URL+"/email.body.plaintext.gotmpl",
				ts.URL+"/email.body.gotmpl",
				ts.URL+"/email.subject.gotmpl"),
				&template.VerificationValidModel{})

			testhelpers.TestRendered(t, tpl)
		})

		t.Run("case=base64 resource", func(t *testing.T) {
			baseUrl := "courier/builtin/templates/verification/valid/"

			toBase64 := func(filePath string) string {
				f, err := ioutil.ReadFile(filePath)
				require.NoError(t, err)
				return base64.StdEncoding.EncodeToString(f)
			}

			tpl := template.NewVerificationValid(testhelpers.SetupRemoteConfig(t,
				toBase64(baseUrl+"email.body.plaintext.gotmpl"),
				toBase64(baseUrl+"email.body.gotmpl"),
				toBase64(baseUrl+"email.subject.gotmpl")),
				&template.VerificationValidModel{})
			testhelpers.TestRendered(t, tpl)
		})

		t.Run("case=file resource", func(t *testing.T) {
			baseUrl := "file://courier/builtin/templates/verification/valid/"

			tpl := template.NewVerificationValid(testhelpers.SetupRemoteConfig(t,
				baseUrl+"email.body.plaintext.gotmpl",
				baseUrl+"email.body.gotmpl",
				baseUrl+"email.subject.gotmpl"),
				&template.VerificationValidModel{},
			)

			testhelpers.TestRendered(t, tpl)
		})
	})
}
