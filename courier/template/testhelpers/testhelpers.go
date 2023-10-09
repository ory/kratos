// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package testhelpers

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"testing"

	"github.com/ory/kratos/courier/template/email"

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
	_, reg := internal.NewVeryFastRegistryWithoutDB(t)
	require.NoError(t, reg.Config().Set(ctx, config.ViperKeyCourierTemplatesRecoveryInvalidEmail, &config.CourierEmailTemplate{
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
},
) {
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
		f, err := os.ReadFile(filePath)
		require.NoError(t, err)
		return base64.StdEncoding.EncodeToString(f)
	}

	getTemplate := func(tmpl courier.TemplateType, d template.Dependencies) interface {
		EmailBody(context.Context) (string, error)
		EmailSubject(context.Context) (string, error)
	} {
		switch tmpl {
		case courier.TypeRecoveryInvalid:
			return email.NewRecoveryInvalid(d, &email.RecoveryInvalidModel{})
		case courier.TypeRecoveryValid:
			return email.NewRecoveryValid(d, &email.RecoveryValidModel{})
		case courier.TypeRecoveryCodeValid:
			return email.NewRecoveryCodeValid(d, &email.RecoveryCodeValidModel{})
		case courier.TypeRecoveryCodeInvalid:
			return email.NewRecoveryCodeInvalid(d, &email.RecoveryCodeInvalidModel{})
		case courier.TypeTestStub:
			return email.NewTestStub(d, &email.TestStubModel{})
		case courier.TypeVerificationInvalid:
			return email.NewVerificationInvalid(d, &email.VerificationInvalidModel{})
		case courier.TypeVerificationValid:
			return email.NewVerificationValid(d, &email.VerificationValidModel{})
		case courier.TypeVerificationCodeInvalid:
			return email.NewVerificationCodeInvalid(d, &email.VerificationCodeInvalidModel{})
		case courier.TypeVerificationCodeValid:
			return email.NewVerificationCodeValid(d, &email.VerificationCodeValidModel{})
		case courier.TypeLoginCodeValid:
			return email.NewLoginCodeValid(d, &email.LoginCodeValidModel{})
		case courier.TypeRegistrationCodeValid:
			return email.NewRegistrationCodeValid(d, &email.RegistrationCodeValidModel{})
		default:
			return nil
		}
	}

	t.Run("case=http resource", func(t *testing.T) {
		t.Parallel()
		router := httprouter.New()
		router.Handle("GET", "/:filename", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
			http.ServeFile(writer, request, path.Join(basePath, params.ByName("filename")))
		})
		ts := httptest.NewServer(router)
		defer ts.Close()

		tpl := getTemplate(tmplType, SetupRemoteConfig(t, ctx,
			ts.URL+"/email.body.plaintext.gotmpl",
			ts.URL+"/email.body.gotmpl",
			ts.URL+"/email.subject.gotmpl"))

		require.NotNil(t, tpl, "Expected to find template for %s in %s", tmplType, basePath)

		TestRendered(t, ctx, tpl)
	})

	t.Run("case=base64 resource", func(t *testing.T) {
		t.Parallel()
		tpl := getTemplate(tmplType, SetupRemoteConfig(t, ctx,
			"base64://"+toBase64(path.Join(basePath, "email.body.plaintext.gotmpl")),
			"base64://"+toBase64(path.Join(basePath, "email.body.gotmpl")),
			"base64://"+toBase64(path.Join(basePath, "email.subject.gotmpl"))))

		require.NotNil(t, tpl, "Expected to find template for %s in %s", tmplType, basePath)

		TestRendered(t, ctx, tpl)
	})

	t.Run("case=file resource", func(t *testing.T) {
		t.Parallel()
		tpl := getTemplate(tmplType, SetupRemoteConfig(t, ctx,
			"file://"+path.Join(basePath, "email.body.plaintext.gotmpl"),
			"file://"+path.Join(basePath, "email.body.gotmpl"),
			"file://"+path.Join(basePath, "email.subject.gotmpl")))

		require.NotNil(t, tpl, "Expected to find template for %s in %s", tmplType, basePath)
		TestRendered(t, ctx, tpl)
	})

	t.Run("case=partial subject override", func(t *testing.T) {
		t.Parallel()
		tpl := getTemplate(tmplType, SetupRemoteConfig(t, ctx,
			"",
			"",
			"base64://"+toBase64(path.Join(basePath, "email.subject.gotmpl"))))

		require.NotNil(t, tpl, "Expected to find template for %s in %s", tmplType, basePath)
		TestRendered(t, ctx, tpl)
	})

	t.Run("case=partial body override", func(t *testing.T) {
		t.Parallel()
		tpl := getTemplate(tmplType, SetupRemoteConfig(t, ctx,
			"base64://"+toBase64(path.Join(basePath, "email.body.plaintext.gotmpl")),
			"base64://"+toBase64(path.Join(basePath, "email.body.gotmpl")),
			""))

		require.NotNil(t, tpl, "Expected to find template for %s in %s", tmplType, basePath)
		TestRendered(t, ctx, tpl)
	})
}
