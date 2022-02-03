package testhelpers

import (
	"context"
	"encoding/base64"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"path"
	"testing"

	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/courier"
	"github.com/ory/kratos/courier/template"
	"github.com/ory/kratos/driver"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/internal"
)

func SetupRemoteConfig(t *testing.T, ctx context.Context, plaintext string, html string, subject string) *driver.RegistryDefault {
	_, reg := internal.NewFastRegistryWithMocks(t)
	require.NoError(t, reg.Config(ctx).Set(config.ViperKeyCourierTemplatesRecoveryInvalidEmail, &config.CourierEmailTemplate{
		TemplateRoot: "",
		Body: &config.CourierEmailBodyTemplate{
			PlainText: plaintext,
			HTML:      html,
		},
		Subject: subject,
	}))
	return reg
}

func TestRendered(t *testing.T, ctx context.Context, tpl interface {
	EmailBody(context.Context) (string, error)
	EmailSubject(context.Context) (string, error)
}) {
	rendered, err := tpl.EmailBody(ctx)
	require.NoError(t, err)
	assert.NotEmpty(t, rendered)

	rendered, err = tpl.EmailSubject(ctx)
	require.NoError(t, err)
	assert.NotEmpty(t, rendered)
}

func TestRemoteTemplates(t *testing.T, basePath string, tmplType courier.TemplateType) {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	toBase64 := func(filePath string) string {
		f, err := ioutil.ReadFile(filePath)
		require.NoError(t, err)
		return base64.StdEncoding.EncodeToString(f)
	}

	getTemplate := func(tmpl courier.TemplateType, d template.TemplateDependencies) interface {
		EmailBody(context.Context) (string, error)
		EmailSubject(context.Context) (string, error)
	} {
		switch tmpl {
		case courier.TypeRecoveryInvalid:
			return template.NewRecoveryInvalid(d, &template.RecoveryInvalidModel{})
		case courier.TypeRecoveryValid:
			return template.NewRecoveryValid(d, &template.RecoveryValidModel{})
		case courier.TypeTestStub:
			return template.NewTestStub(d, &template.TestStubModel{})
		case courier.TypeVerificationInvalid:
			return template.NewVerificationInvalid(d, &template.VerificationInvalidModel{})
		case courier.TypeVerificationValid:
			return template.NewVerificationValid(d, &template.VerificationValidModel{})
		default:
			return nil
		}
	}

	t.Run("case=http resource", func(t *testing.T) {
		router := httprouter.New()
		router.Handle("GET", "/email.body.plaintext.gotpml", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
			http.ServeFile(writer, request, path.Join(basePath, "email.body.plaintext.gotmpl"))
		})
		router.Handle("GET", "/plaintext", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
			http.ServeFile(writer, request, path.Join(basePath, "courier/builtin/templates/recovery/invalid/email.body.gotmpl"))
		})
		ts := httptest.NewServer(router)
		defer ts.Close()

		tpl := getTemplate(tmplType, SetupRemoteConfig(t, ctx,
			ts.URL+"/email.body.plaintext.gotmpl",
			ts.URL+"/email.body.gotmpl",
			ts.URL+"/email.subject.gotmpl"))

		TestRendered(t, ctx, tpl)
	})

	t.Run("case=base64 resource", func(t *testing.T) {
		tpl := getTemplate(tmplType, SetupRemoteConfig(t, ctx,
			toBase64(path.Join(basePath, "email.body.plaintext.gotmpl")),
			toBase64(path.Join(basePath, "email.body.gotmpl")),
			toBase64(path.Join(basePath, "email.subject.gotmpl"))))

		TestRendered(t, ctx, tpl)
	})

	t.Run("case=file resource", func(t *testing.T) {
		baseUrl := "file://" + basePath

		tpl := getTemplate(tmplType, SetupRemoteConfig(t, ctx,
			baseUrl+"email.body.plaintext.gotmpl",
			baseUrl+"email.body.gotmpl",
			baseUrl+"email.subject.gotmpl"))

		TestRendered(t, ctx, tpl)
	})

	t.Run("case=partial subject override", func(t *testing.T) {
		tpl := getTemplate(tmplType, SetupRemoteConfig(t, ctx,
			"",
			"",
			toBase64(path.Join(basePath, "email.subject.gotmpl"))))
		TestRendered(t, ctx, tpl)
	})

	t.Run("case=partial body override", func(t *testing.T) {
		tpl := getTemplate(tmplType, SetupRemoteConfig(t, ctx,
			toBase64(path.Join(basePath, "email.body.plaintext.gotmpl")),
			toBase64(path.Join(basePath, "email.body.gotmpl")),
			""))
		TestRendered(t, ctx, tpl)
	})
}
