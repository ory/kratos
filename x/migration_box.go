package x

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/gobuffalo/pop/v5"
	"github.com/markbates/pkger"
	"github.com/pkg/errors"
)

// MigrationPkger is a wrapper around pkger.Dir and Migrator.
// This will allow you to run migrations from migrations packed
// inside of a compiled binary.
type MigrationPkger struct {
	pop.Migrator
	Dir pkger.Dir
	r   LoggingProvider
}

// NewPkgerMigration from a packr.Box and a Connection.
//
//	migrations, err := NewPkgerMigration(pkger.Dir("/migrations"))
//
func NewPkgerMigration(dir pkger.Dir, c *pop.Connection, r LoggingProvider) (MigrationPkger, error) {
	fm := MigrationPkger{
		Migrator: pop.NewMigrator(c),
		Dir:      dir,
		r:        r,
	}

	runner := func(f io.Reader) func(mf pop.Migration, tx *pop.Connection) error {
		return func(mf pop.Migration, tx *pop.Connection) error {
			content, err := pop.MigrationContent(mf, tx, f, true)
			if err != nil {
				return errors.Wrapf(err, "error processing %s", mf.Path)
			}
			if content == "" {
				return nil
			}
			err = tx.RawQuery(content).Exec()
			if err != nil {
				return errors.Wrapf(err, "error executing %s, sql: %s", mf.Path, content)
			}
			return nil
		}
	}

	err := fm.findMigrations(runner)
	if err != nil {
		return fm, err
	}

	return fm, nil
}

func (fm *MigrationPkger) findMigrations(runner func(f io.Reader) func(mf pop.Migration, tx *pop.Connection) error) error {
	return pkger.Walk(string(fm.Dir), func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		match, err := pop.ParseMigrationFilename(info.Name())
		if err != nil {
			if strings.HasPrefix(err.Error(), "unsupported dialect") {
				fm.r.Logger().Debugf("Ignoring migration file because dialect is not supported: %s", err.Error())
				return nil
			}
			return err
		}

		if match == nil {
			return nil
		}

		file, err := pkger.Open(p)
		if err != nil {
			return err
		}
		defer file.Close()

		content, err := ioutil.ReadAll(file)
		if err != nil {
			return err
		}

		mf := pop.Migration{
			Path:      p,
			Version:   match.Version,
			Name:      match.Name,
			DBType:    match.DBType,
			Direction: match.Direction,
			Type:      match.Type,
			Runner:    runner(bytes.NewReader(content)),
		}
		fm.Migrations[mf.Direction] = append(fm.Migrations[mf.Direction], mf)
		return nil
	})
}
