// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package popx

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"os"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach-go/v2/crdb"
	"github.com/ory/pop/v6"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/ory/x/cmdx"
	"github.com/ory/x/logrusx"
	"github.com/ory/x/otelx"
)

const (
	Pending          = "Pending"
	Applied          = "Applied"
	tracingComponent = "github.com/ory/x/popx"
)

// Up runs pending "up" migrations and applies them to the database.
func (mb *MigrationBox) Up(ctx context.Context) error {
	_, err := mb.UpTo(ctx, 0)
	return err
}

// UpTo runs up to step "up" migrations and applies them to the database.
// If step <= 0 all pending migrations are run.
func (mb *MigrationBox) UpTo(ctx context.Context, step int) (applied int, err error) {
	ctx, span := startSpan(ctx, MigrationUpOpName, trace.WithAttributes(attribute.Int("step", step)))
	defer otelx.End(span, &err)

	c := mb.c.WithContext(ctx)
	err = mb.exec(ctx, func() error {
		mtn := sanitizedMigrationTableName(c)
		mfs := mb.migrationsUp.sortAndFilter(c.Dialect.Name())
		for _, mi := range mfs {
			l := mb.l.WithField("version", mi.Version).WithField("migration_name", mi.Name).WithField("migration_file", mi.Path)

			appliedMigrations := make([]string, 0, 2)
			legacyVersion := mi.Version
			if len(legacyVersion) > 14 {
				legacyVersion = legacyVersion[:14]
			}
			err := c.RawQuery(fmt.Sprintf("SELECT version FROM %s WHERE version IN (?, ?)", mtn), mi.Version, legacyVersion).All(&appliedMigrations)
			if err != nil {
				return errors.Wrapf(err, "problem checking for migration version %s", mi.Version)
			}

			if slices.Contains(appliedMigrations, mi.Version) {
				l.Debug("Migration has already been applied, skipping.")
				continue
			}

			if slices.Contains(appliedMigrations, legacyVersion) {
				l.WithField("legacy_version", legacyVersion).WithField("migration_table", mtn).Debug("Migration has already been applied in a legacy migration run. Updating version in migration table.")
				if err := mb.isolatedTransaction(ctx, "init-migrate", func(conn *pop.Connection) error {
					// We do not want to remove the legacy migration version or subsequent migrations might be applied twice.
					//
					// Do not activate the following - it is just for reference.
					//
					// if _, err := tx.Store.Exec(fmt.Sprintf("DELETE FROM %s WHERE version = ?", mtn), legacyVersion); err != nil {
					//	return errors.Wrapf(err, "problem removing legacy version %s", mi.Version)
					// }

					// #nosec G201 - mtn is a system-wide const
					err := conn.RawQuery(fmt.Sprintf("INSERT INTO %s (version) VALUES (?)", mtn), mi.Version).Exec()
					return errors.Wrapf(err, "problem inserting migration version %s", mi.Version)
				}); err != nil {
					return err
				}
				continue
			}

			l.Info("Migration has not yet been applied, running migration.")

			if err := mi.Valid(); err != nil {
				return err
			}

			if mi.Runner != nil {
				err := mb.isolatedTransaction(ctx, "up", func(conn *pop.Connection) error {
					if err := mi.Runner(mi, conn, conn.TX); err != nil {
						return err
					}

					// #nosec G201 - mtn is a system-wide const
					if err := conn.RawQuery(fmt.Sprintf("INSERT INTO %s (version) VALUES (?)", mtn), mi.Version).Exec(); err != nil {
						return errors.Wrapf(err, "problem inserting migration version %s", mi.Version)
					}
					return nil
				})
				if err != nil {
					return err
				}
			} else {
				l.Warn("Migration has requested running outside a transaction. Proceed with caution.")
				if err := mi.RunnerNoTx(mi, c); err != nil {
					return err
				}

				// #nosec G201 - mtn is a system-wide const
				if err := c.RawQuery(fmt.Sprintf("INSERT INTO %s (version) VALUES (?)", mtn), mi.Version).Exec(); err != nil {
					return errors.Wrapf(err, "problem inserting migration version %s. YOUR DATABASE MAY BE IN AN INCONSISTENT STATE! MANUAL INTERVENTION REQUIRED!", mi.Version)
				}
			}

			l.Infof("> %s applied successfully", mi.Name)
			applied++
			if step > 0 && applied >= step {
				break
			}
		}
		if applied == 0 {
			mb.l.Infof("Migrations already up to date, nothing to apply")
		} else {
			mb.l.Infof("Successfully applied %d migrations.", applied)
		}
		return nil
	})
	return
}

// Down runs pending "down" migrations and rolls back the
// database by the specified number of steps.
func (mb *MigrationBox) Down(ctx context.Context, steps int) (err error) {
	ctx, span := startSpan(ctx, MigrationDownOpName, trace.WithAttributes(attribute.Int("steps", steps)))
	defer otelx.End(span, &err)

	if steps <= 0 {
		steps = math.MaxInt
	}

	c := mb.c.WithContext(ctx)
	return mb.exec(ctx, func() (err error) {
		mtn := sanitizedMigrationTableName(c)
		count, err := c.Count(mtn)
		if err != nil {
			return errors.Wrap(err, "migration down: unable count existing migration")
		}
		steps = min(steps, count)

		mfs := mb.migrationsDown.sortAndFilter(c.Dialect.Name())
		slices.Reverse(mfs)
		if len(mfs) > count {
			// skip all migrations that were not yet applied
			mfs = mfs[len(mfs)-count:]
		}

		reverted := 0
		defer func() {
			mb.l.Debugf("Successfully reverted %d migrations.", reverted)
			if err != nil {
				mb.l.WithError(err).Error("Problem reverting migrations.")
			}
		}()
		for i, mi := range mfs {
			if i >= steps {
				break
			}
			l := mb.l.WithField("version", mi.Version).WithField("migration_name", mi.Name).WithField("migration_file", mi.Path)
			exists, err := c.Where("version = ?", mi.Version).Exists(mtn)
			if err != nil {
				return errors.Wrapf(err, "problem checking for migration version %s", mi.Version)
			}

			if !exists && len(mi.Version) > 14 {
				legacyVersion := mi.Version[:14]
				legacyVersionExists, err := c.Where("version = ?", legacyVersion).Exists(mtn)
				if err != nil {
					return errors.Wrapf(err, "problem checking for legacy migration version %s", legacyVersion)
				}

				if !legacyVersionExists {
					return errors.Errorf("neither normal (%s) nor legacy migration (%s) exist", mi.Version, legacyVersion)
				}
			} else if !exists {
				return errors.Errorf("migration version %s does not exist", mi.Version)
			}

			if err := mi.Valid(); err != nil {
				return err
			}

			if mi.Runner != nil {
				err := mb.isolatedTransaction(ctx, "down", func(conn *pop.Connection) error {
					err := mi.Runner(mi, conn, conn.TX)
					if err != nil {
						return err
					}

					// #nosec G201 - mtn is a system-wide const
					if err := conn.RawQuery(fmt.Sprintf("DELETE FROM %s WHERE version = ?", mtn), mi.Version).Exec(); err != nil {
						return errors.Wrapf(err, "problem deleting migration version %s", mi.Version)
					}

					return nil
				})
				if err != nil {
					return err
				}
			} else {
				err := mi.RunnerNoTx(mi, c)
				if err != nil {
					return err
				}

				// #nosec G201 - mtn is a system-wide const
				if err := c.RawQuery(fmt.Sprintf("DELETE FROM %s WHERE version = ?", mtn), mi.Version).Exec(); err != nil {
					return errors.Wrapf(err, "problem deleting migration version %s. YOUR DATABASE MAY BE IN AN INCONSISTENT STATE! MANUAL INTERVENTION REQUIRED!", mi.Version)
				}
			}

			l.Infof("< %s applied successfully", mi.Name)
			reverted++
		}
		return nil
	})
}

func (mb *MigrationBox) createTransactionalMigrationTable(ctx context.Context, c *pop.Connection, l *logrusx.Logger) error {
	mtn := sanitizedMigrationTableName(c)

	if err := mb.execMigrationTransaction(ctx, []string{
		fmt.Sprintf(`CREATE TABLE %s (version VARCHAR (48) NOT NULL, version_self INT NOT NULL DEFAULT 0)`, mtn),
		fmt.Sprintf(`CREATE UNIQUE INDEX %s_version_idx ON %s (version)`, mtn, mtn),
		fmt.Sprintf(`CREATE INDEX %s_version_self_idx ON %s (version_self)`, mtn, mtn),
	}); err != nil {
		return err
	}

	l.WithField("migration_table", mtn).Debug("Transactional migration table created successfully.")

	return nil
}

func (mb *MigrationBox) migrateToTransactionalMigrationTable(ctx context.Context, c *pop.Connection, l *logrusx.Logger) error {
	// This means the new pop migrator has also not yet been applied, do that now.
	mtn := sanitizedMigrationTableName(c)

	withOn := fmt.Sprintf(" ON %s", mtn)
	if c.Dialect.Name() != "mysql" {
		withOn = ""
	}

	interimTable := fmt.Sprintf("%s_transactional", mtn)
	workload := [][]string{
		{
			fmt.Sprintf(`DROP INDEX %s_version_idx%s`, mtn, withOn),
			fmt.Sprintf(`CREATE TABLE %s (version VARCHAR (48) NOT NULL, version_self INT NOT NULL DEFAULT 0)`, interimTable),
			fmt.Sprintf(`CREATE UNIQUE INDEX %s_version_idx ON %s (version)`, mtn, interimTable),
			fmt.Sprintf(`CREATE INDEX %s_version_self_idx ON %s (version_self)`, mtn, interimTable),
			// #nosec G201 - mtn is a system-wide const
			fmt.Sprintf(`INSERT INTO %s (version) SELECT version FROM %s`, interimTable, mtn),
			fmt.Sprintf(`ALTER TABLE %s RENAME TO %s_pop_legacy`, mtn, mtn),
		},
		{
			fmt.Sprintf(`ALTER TABLE %s RENAME TO %s`, interimTable, mtn),
		},
	}

	if err := mb.execMigrationTransaction(ctx, workload...); err != nil {
		return err
	}

	l.WithField("migration_table", mtn).Debug("Successfully migrated legacy schema_migration to new transactional schema_migration table.")

	return nil
}

func (mb *MigrationBox) isolatedTransaction(ctx context.Context, direction string, fn func(c *pop.Connection) error) (err error) {
	ctx, span := startSpan(ctx, MigrationRunTransactionOpName, trace.WithAttributes(attribute.String("migration_direction", direction)))
	defer otelx.End(span, &err)

	if mb.perMigrationTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, mb.perMigrationTimeout)
		defer cancel()
	}

	conn, dberr := mb.c.NewTransactionContextOptions(ctx, &sql.TxOptions{
		Isolation: sql.LevelSerializable,
		ReadOnly:  false,
	})
	if dberr != nil {
		return dberr
	}

	err = fn(conn)
	if err != nil {
		dberr = conn.TX.Rollback()
	} else {
		dberr = conn.TX.Commit()
	}

	if dberr != nil {
		return errors.Wrapf(dberr, "error committing or rolling back transaction; original error: %v", err)
	}

	return err
}

func (mb *MigrationBox) execMigrationTransaction(ctx context.Context, transactions ...[]string) error {
	for _, statements := range transactions {
		if err := mb.isolatedTransaction(ctx, "init", func(conn *pop.Connection) error {
			for _, statement := range statements {
				if _, err := conn.TX.ExecContext(ctx, statement); err != nil {
					return errors.Wrapf(err, "unable to execute statement: %s", statement)
				}
			}
			return nil
		}); err != nil {
			return err
		}
	}

	return nil
}

// CreateSchemaMigrations sets up a table to track migrations. This is an idempotent
// operation.
func (mb *MigrationBox) CreateSchemaMigrations(ctx context.Context) error {
	ctx, span := startSpan(ctx, MigrationInitOpName)
	defer span.End()

	c := mb.c.WithContext(ctx)

	mtn := sanitizedMigrationTableName(c)
	mb.l.WithField("migration_table", mtn).Debug("Checking if legacy migration table exists.")
	_, err := c.Store.Exec(fmt.Sprintf("select version from %s", mtn))
	if err != nil {
		mb.l.WithError(err).WithField("migration_table", mtn).Debug("An error occurred while checking for the legacy migration table, maybe it does not exist yet? Trying to create.")
		// This means that the legacy pop migrator has not yet been applied
		return mb.createTransactionalMigrationTable(ctx, c, mb.l)
	}

	mb.l.WithField("migration_table", mtn).Debug("A migration table exists, checking if it is a transactional migration table.")
	_, err = c.Store.Exec(fmt.Sprintf("select version, version_self from %s", mtn))
	if err != nil {
		mb.l.WithError(err).WithField("migration_table", mtn).Debug("An error occurred while checking for the transactional migration table, maybe it does not exist yet? Trying to create.")
		return mb.migrateToTransactionalMigrationTable(ctx, c, mb.l)
	}

	mb.l.WithField("migration_table", mtn).Debug("Migration tables exist and are up to date.")
	return nil
}

type MigrationStatus struct {
	State       string `json:"state"`
	Version     string `json:"version"`
	Name        string `json:"name"`
	ContentUp   string `json:"content"`
	ContentDown string `json:"content_down"`
}

type MigrationStatuses []MigrationStatus

var _ cmdx.Table = (MigrationStatuses)(nil)

func (m MigrationStatuses) Header() []string {
	return []string{"Version", "Name", "Status"}
}

func (m MigrationStatuses) Table() [][]string {
	t := make([][]string, len(m))
	for i, s := range m {
		t[i] = []string{s.Version, s.Name, s.State}
	}
	return t
}

func (m MigrationStatuses) Interface() interface{} {
	return m
}

func (m MigrationStatuses) Len() int {
	return len(m)
}

func (m MigrationStatuses) IDs() []string {
	ids := make([]string, len(m))
	for i, s := range m {
		ids[i] = s.Version
	}
	return ids
}

func (m MigrationStatuses) HasPending() bool {
	for _, mm := range m {
		if mm.State == Pending {
			return true
		}
	}
	return false
}

func sanitizedMigrationTableName(con *pop.Connection) string {
	return regexp.MustCompile(`\W`).ReplaceAllString(con.MigrationTableName(), "")
}

func errIsTableNotFound(err error) bool {
	return strings.Contains(err.Error(), "no such table:") || // sqlite
		strings.Contains(err.Error(), "Error 1146") || // MySQL
		strings.Contains(err.Error(), "SQLSTATE 42P01") // PostgreSQL / CockroachDB
}

// Status prints out the status of applied/pending migrations.
func (mb *MigrationBox) Status(ctx context.Context) (MigrationStatuses, error) {
	ctx, span := startSpan(ctx, MigrationStatusOpName)
	defer span.End()

	con := mb.c.WithContext(ctx)

	migrationsUp := mb.migrationsUp.sortAndFilter(con.Dialect.Name())

	if len(migrationsUp) == 0 {
		return nil, errors.Errorf("unable to find any migrations for dialect: %s", con.Dialect.Name())
	}

	alreadyApplied := make([]string, 0, len(migrationsUp))
	err := con.RawQuery(fmt.Sprintf("SELECT version FROM %s", sanitizedMigrationTableName(con))).All(&alreadyApplied)
	if err != nil {
		if errIsTableNotFound(err) {
			// This means that no migrations have been applied and we need to apply all of them first!
			//
			// It also means that we can ignore this state and act as if no migrations have been applied yet.
		} else {
			// On any other error, we fail.
			return nil, errors.Wrapf(err, "problem with migration")
		}
	}

	statuses := make(MigrationStatuses, len(migrationsUp))
	for k, mf := range migrationsUp {
		downContent := "-- error: no down migration defined for this migration"
		if mDown := mb.migrationsDown.find(mf.Version, con.Dialect.Name()); mDown != nil {
			downContent = mDown.Content
		}
		statuses[k] = MigrationStatus{
			State:       Pending,
			Version:     mf.Version,
			Name:        mf.Name,
			ContentUp:   mf.Content,
			ContentDown: downContent,
		}

		if slices.ContainsFunc(alreadyApplied, func(applied string) bool {
			return applied == mf.Version || (len(mf.Version) > 14 && applied == mf.Version[:14])
		}) {
			statuses[k].State = Applied
			continue
		}
	}

	return statuses, nil
}

// DumpMigrationSchema will generate a file of the current database schema
func (mb *MigrationBox) DumpMigrationSchema(ctx context.Context) error {
	c := mb.c.WithContext(ctx)
	schema := "schema.sql"
	f, err := os.Create(schema) //#nosec:G304) //#nosec:G304
	if err != nil {
		return err
	}
	err = c.Dialect.DumpSchema(f)
	if err != nil {
		_ = os.RemoveAll(schema)
		return err
	}
	return nil
}

func (mb *MigrationBox) exec(ctx context.Context, fn func() error) error {
	now := time.Now()
	defer func() {
		if !mb.dumpMigrations {
			return
		}
		err := mb.DumpMigrationSchema(ctx)
		if err != nil {
			mb.l.WithError(err).Error("Migrator: unable to dump schema")
		}
	}()
	defer mb.printTimer(now)

	err := mb.CreateSchemaMigrations(ctx)
	if err != nil {
		return errors.Wrap(err, "migrator: problem creating schema migrations")
	}

	if mb.c.Dialect.Name() == "sqlite3" {
		if err := mb.c.RawQuery("PRAGMA foreign_keys=OFF").Exec(); err != nil {
			return err
		}
	}

	if mb.c.Dialect.Name() == "cockroach" {
		outer := fn
		fn = func() error {
			return crdb.Execute(outer)
		}
	}

	if err := fn(); err != nil {
		return err
	}

	if mb.c.Dialect.Name() == "sqlite3" {
		if err := mb.c.RawQuery("PRAGMA foreign_keys=ON").Exec(); err != nil {
			return err
		}
	}

	return nil
}

func (mb *MigrationBox) printTimer(timerStart time.Time) {
	diff := time.Since(timerStart).Seconds()
	if diff > 60 {
		mb.l.Debugf("%.4f minutes", diff/60)
	} else {
		mb.l.Debugf("%.4f seconds", diff)
	}
}
