package driver

import (
	"fmt"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	migrate "github.com/rubenv/sql-migrate"

	"github.com/ory/x/logrusx"

	"github.com/ory/x/dbal"
	"github.com/ory/x/sqlcon"
	"github.com/ory/x/urlx"

	"github.com/ory/hive/identity"
	"github.com/ory/hive/selfservice"
	"github.com/ory/hive/selfservice/oidc"
	"github.com/ory/hive/selfservice/password"
	"github.com/ory/hive/session"
)

var _ Registry = new(RegistrySQL)

var Migrations = map[string]*dbal.PackrMigrationSource{}

var requestManagerFactories = map[identity.CredentialsType]func() selfservice.RequestMethodConfig{
	identity.CredentialsTypeOIDC: func() selfservice.RequestMethodConfig {
		return oidc.NewRequestMethodConfig()
	},
	identity.CredentialsTypePassword: func() selfservice.RequestMethodConfig {
		return password.NewRequestMethodConfig()
	},
}

func init() {
	l := logrusx.New()
	Migrations[dbal.DriverPostgreSQL] = dbal.NewMustPackerMigrationSource(l, AssetNames(), Asset, []string{"../contrib/sql/migrations/postgres/"}, true)

	dbal.RegisterDriver(NewRegistrySQL())
}

type RegistrySQL struct {
	*RegistryAbstract

	identityPool              identity.Pool
	sessionManager            session.Manager
	selfserviceRequestManager selfservice.RequestManager

	db *sqlx.DB
}

func NewRegistrySQL() *RegistrySQL {
	r := &RegistrySQL{RegistryAbstract: new(RegistryAbstract)}
	r.RegistryAbstract.with(r)
	return r
}

func (m *RegistrySQL) WithDB(db *sqlx.DB) Registry {
	m.db = db
	return m
}

func (m *RegistrySQL) CanHandle(dsn string) bool {
	s := dbal.Canonicalize(urlx.ParseOrFatal(m.l, dsn).Scheme)
	return s == dbal.DriverPostgreSQL
}

func (m *RegistrySQL) Ping() error {
	return m.DB().Ping()
}

func (m *RegistrySQL) IdentityPool() identity.Pool {
	if m.identityPool == nil {
		m.identityPool = identity.NewPoolMemory(m.c, m)
	}
	return m.identityPool
}

func (m *RegistrySQL) SessionManager() session.Manager {
	if m.sessionManager == nil {
		m.sessionManager = session.NewManagerSQL(m.c, m, m.DB())
	}
	return m.sessionManager
}

func (m *RegistrySQL) Init() error {
	if m.db != nil {
		return nil
	}

	var options []sqlcon.OptionModifier
	if m.Tracer().IsLoaded() {
		options = append(options, sqlcon.WithDistributedTracing(), sqlcon.WithOmitArgsFromTraceSpans())
	}

	connection, err := sqlcon.NewSQLConnection(m.c.DSN(), m.Logger())
	if err != nil {
		return err
	}

	m.db, err = connection.GetDatabaseRetry(time.Second*5, time.Minute*5)
	if err != nil {
		return err
	}

	return err
}

func (m *RegistrySQL) DB() *sqlx.DB {
	if m.db == nil {
		if err := m.Init(); err != nil {
			m.Logger().WithError(err).Fatalf("Unable to initialize database.")
		}
	}

	return m.db
}

func (m *RegistrySQL) RegistrationRequestManager() selfservice.RegistrationRequestManager {
	if m.selfserviceRequestManager == nil {
		m.selfserviceRequestManager = selfservice.NewRequestManagerSQL(
			m.DB(),
			requestManagerFactories,
		)
	}
	return m.selfserviceRequestManager
}

func (m *RegistrySQL) LoginRequestManager() selfservice.LoginRequestManager {
	if m.selfserviceRequestManager == nil {
		m.selfserviceRequestManager = selfservice.NewRequestManagerSQL(
			m.DB(),
			requestManagerFactories,
		)
	}
	return m.selfserviceRequestManager
}

func (m *RegistrySQL) CreateSchemas(dbName string) (int, error) {
	var total int

	m.Logger().Debugf("Applying %s SQL migrations...", dbName)

	migrate.SetTable("hive_migration")
	n, err := migrate.Exec(m.DB().DB, dbal.Canonicalize(m.DB().DriverName()), Migrations[dbName], migrate.Up)
	if err != nil {
		return 0, errors.Wrapf(err, "Could not migrate sql schema, applied %d Migrations", n)
	}

	m.Logger().Debugf("Applied %d %s SQL migrations", total, dbName)
	return n, nil
}

func SQLPurgeTestDatabase(t *testing.T, db *sqlx.DB) {
	for _, table := range []string{
		"identity_credentials_identifiers",
		"identity_credentials",
		"identity",
	} {
		_, err := db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s", table))
		if err != nil {
			t.Logf("Unable to clean up table %s: %s", table, err)
		}
	}
}
