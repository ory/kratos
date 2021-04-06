package courier_test

import (
	"fmt"
	"testing"

	"github.com/ory/kratos/courier"
	"github.com/ory/kratos/courier/template"
	"github.com/stretchr/testify/require"
)

func TestGetTemplateType(t *testing.T) {
	for expectedType, tmpl := range map[courier.TemplateType]courier.EmailTemplate{
		courier.TypeRecoveryInvalid:     &template.RecoveryInvalid{},
		courier.TypeRecoveryValid:       &template.RecoveryValid{},
		courier.TypeVerificationInvalid: &template.VerificationInvalid{},
		courier.TypeVerificationValid:   &template.VerificationValid{},
		courier.TypeTestStub:            &template.TestStub{},
	} {
		t.Run(fmt.Sprintf("case=%s", expectedType), func(t *testing.T) {
			actualType, err := courier.GetTemplateType(tmpl)
			require.NoError(t, err)
			require.Equal(t, expectedType, actualType)

		})

	}
}
