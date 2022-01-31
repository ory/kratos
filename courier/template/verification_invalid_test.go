package template_test

import (
	"context"
	"encoding/base64"
	"github.com/julienschmidt/httprouter"
	"github.com/ory/kratos/courier/template/testhelpers"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ory/kratos/courier/template"
	"github.com/ory/kratos/internal"
)

func TestVerifyInvalid(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	t.Run("test=with courier templates directory", func(t *testing.T) {
		_, reg := internal.NewFastRegistryWithMocks(t)
		tpl := template.NewVerificationInvalid(reg, &template.VerificationInvalidModel{})

		testhelpers.TestRendered(t, ctx, tpl)
	})

	t.Run("test=with remote resources", func(t *testing.T) {
		t.Run("case=http resource", func(t *testing.T) {
			router := httprouter.New()
			router.Handle("GET", "/email.body.plaintext.gotpml", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
				http.ServeFile(writer, request, "courier/builtin/templates/verification/invalid/email.body.plaintext.gotmpl")
			})
			router.Handle("GET", "/plaintext", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
				http.ServeFile(writer, request, "courier/builtin/templates/verification/invalid/email.body.gotmpl")
			})
			ts := httptest.NewServer(router)
			defer ts.Close()

			tpl := template.NewVerificationInvalid(testhelpers.SetupRemoteConfig(t, ctx,
				ts.URL+"/email.body.plaintext.gotmpl",
				ts.URL+"/email.body.gotmpl",
				ts.URL+"/email.subject.gotmpl"),
				&template.VerificationInvalidModel{})

			testhelpers.TestRendered(t, ctx, tpl)
		})

		t.Run("case=base64 resource", func(t *testing.T) {
			baseUrl := "courier/builtin/templates/verification/invalid/"

			toBase64 := func(filePath string) string {
				f, err := ioutil.ReadFile(filePath)
				require.NoError(t, err)
				return base64.StdEncoding.EncodeToString(f)
			}

			tpl := template.NewVerificationInvalid(testhelpers.SetupRemoteConfig(t, ctx,
				toBase64(baseUrl+"email.body.plaintext.gotmpl"),
				toBase64(baseUrl+"email.body.gotmpl"),
				toBase64(baseUrl+"email.subject.gotmpl")),
				&template.VerificationInvalidModel{})
			testhelpers.TestRendered(t, ctx, tpl)
		})

		t.Run("case=file resource", func(t *testing.T) {
			baseUrl := "file://courier/builtin/templates/verification/invalid/"

			tpl := template.NewVerificationInvalid(testhelpers.SetupRemoteConfig(t, ctx,
				baseUrl+"email.body.plaintext.gotmpl",
				baseUrl+"email.body.gotmpl",
				baseUrl+"email.subject.gotmpl"),
				&template.VerificationInvalidModel{},
			)

			testhelpers.TestRendered(t, ctx, tpl)
		})
	})
}
