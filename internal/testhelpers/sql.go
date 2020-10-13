package testhelpers

import (
	"testing"

	"github.com/gobuffalo/pop/v5"

	"github.com/ory/kratos/selfservice/errorx"

	"github.com/ory/kratos/continuity"
	"github.com/ory/kratos/courier"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/selfservice/flow/recovery"
	"github.com/ory/kratos/selfservice/flow/registration"
	"github.com/ory/kratos/selfservice/flow/settings"
	"github.com/ory/kratos/selfservice/flow/verification"
	"github.com/ory/kratos/selfservice/strategy/link"
	"github.com/ory/kratos/session"
)

func CleanSQL(t *testing.T, c *pop.Connection) {
	for _, table := range []string{
		new(continuity.Container).TableName(),
		new(courier.Message).TableName(),

		new(login.FlowMethods).TableName(),
		new(login.Flow).TableName(),

		new(registration.FlowMethods).TableName(),
		new(registration.Flow).TableName(),

		new(settings.FlowMethods).TableName(),
		new(settings.Flow).TableName(),

		new(link.RecoveryToken).TableName(),
		new(link.VerificationToken).TableName(),

		new(recovery.FlowMethods).TableName(),
		new(recovery.Flow).TableName(),

		new(verification.Flow).TableName(),
		new(verification.FlowMethods).TableName(),

		new(errorx.ErrorContainer).TableName(),

		new(session.Session).TableName(),
		new(identity.CredentialIdentifierCollection).TableName(),
		new(identity.CredentialsCollection).TableName(),
		new(identity.VerifiableAddress).TableName(),
		new(identity.RecoveryAddress).TableName(),
		new(identity.Identity).TableName(),
		new(identity.CredentialsTypeTable).TableName(),
		"schema_migration",
	} {
		if err := c.RawQuery("DROP TABLE IF EXISTS " + table).Exec(); err != nil {
			t.Logf(`Unable to clean up table "%s": %s`, table, err)
		}
	}
	t.Logf("Successfully cleaned up database: %s", c.Dialect.Name())
}
