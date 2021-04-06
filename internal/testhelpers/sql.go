package testhelpers

import (
	"context"
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
	ctx := context.Background()
	for _, table := range []string{
		new(continuity.Container).TableName(ctx),
		new(courier.Message).TableName(ctx),

		new(login.Flow).TableName(ctx),
		new(registration.Flow).TableName(ctx),
		new(settings.Flow).TableName(ctx),

		new(link.RecoveryToken).TableName(ctx),
		new(link.VerificationToken).TableName(ctx),

		new(recovery.Flow).TableName(ctx),

		new(verification.Flow).TableName(ctx),
		new(verification.FlowMethods).TableName(ctx),

		new(errorx.ErrorContainer).TableName(ctx),

		new(session.Session).TableName(ctx),
		new(identity.CredentialIdentifierCollection).TableName(ctx),
		new(identity.CredentialsCollection).TableName(ctx),
		new(identity.VerifiableAddress).TableName(ctx),
		new(identity.RecoveryAddress).TableName(ctx),
		new(identity.Identity).TableName(ctx),
		new(identity.CredentialsTypeTable).TableName(ctx),
		"schema_migration",
	} {
		if err := c.RawQuery("DROP TABLE IF EXISTS " + table).Exec(); err != nil {
			t.Logf(`Unable to clean up table "%s": %s`, table, err)
		}
	}
	t.Logf("Successfully cleaned up database: %s", c.Dialect.Name())
}
