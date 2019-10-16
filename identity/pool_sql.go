package identity

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"

	"github.com/ory/x/stringsx"

	"github.com/ory/x/sqlcon"
	"github.com/ory/x/sqlxx"

	"github.com/ory/herodot"

	"github.com/ory/hive/driver/configuration"
)

var _ Pool = new(PoolSQL)

type (
	PoolSQL struct {
		*abstractPool
		db *sqlx.DB
	}
)

func NewPoolSQL(c configuration.Provider, d ValidationProvider, db *sqlx.DB) *PoolSQL {
	return &PoolSQL{abstractPool: newAbstractPool(c, d), db: db}
}

// FindByCredentialsIdentifier returns an identity by querying for it's credential identifiers.
func (p *PoolSQL) FindByCredentialsIdentifier(ctx context.Context, ct CredentialsType, match string) (*Identity, *Credentials, error) {
	i, err := p.get(ctx, "WHERE ici.identifier = ? AND ic.method = ?", []interface{}{match, string(ct)})
	if err != nil {
		if errors.Cause(err).Error() == herodot.ErrNotFound.Error() {
			return nil, nil, herodot.ErrNotFound.WithTrace(err).WithReasonf(`No identity matching credentials identifier "%s" could be found.`, match)
		}
		return nil, nil, err
	}

	creds, ok := i.Credentials[ct]
	if !ok {
		return nil, nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("The SQL adapter failed to return the appropriate credentials_type \"%s\". This is a bug in the code.", ct))
	}

	return p.declassify(*i), &creds, nil
}

func (p *PoolSQL) Create(ctx context.Context, i *Identity) (*Identity, error) {
	insert := p.augment(*i)
	if err := p.Validate(insert); err != nil {
		return nil, err
	}

	tx, err := p.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}

	if err := p.insert(ctx, tx, insert); err != nil {
		if err := tx.Rollback(); err != nil {
			return nil, errors.WithStack(err)
		}
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		if err := tx.Rollback(); err != nil {
			return nil, errors.WithStack(err)
		}
		return nil, errors.WithStack(err)
	}

	return p.abstractPool.declassify(*insert), nil
}

func (p *PoolSQL) List(ctx context.Context, limit, offset int) ([]Identity, error) {
	var rows []struct {
		ID              string          `db:"id"`
		TraitsSchemaURL string          `db:"traits_schema_url"`
		Traits          json.RawMessage `db:"traits"`
	}

	query := "SELECT id, traits, traits_schema_url FROM identity ORDER BY pk LIMIT ? OFFSET ?"
	if err := p.db.SelectContext(ctx, &rows, p.db.Rebind(query), limit, offset); err != nil {
		return nil, sqlcon.HandleError(err)
	}

	ids := make([]Identity, len(rows))
	for k, row := range rows {
		ids[k] = Identity{
			ID:              row.ID,
			TraitsSchemaURL: row.TraitsSchemaURL,
			Traits:          row.Traits,
		}
	}

	return p.declassifyAll(ids), nil
}

func (p *PoolSQL) UpdateConfidential(ctx context.Context, i *Identity, ct map[CredentialsType]Credentials) (*Identity, error) {
	return p.update(ctx, i, ct, true)
}

func (p *PoolSQL) Update(ctx context.Context, i *Identity) (*Identity, error) {
	return p.update(ctx, i, nil, false)
}

func (p *PoolSQL) Get(ctx context.Context, id string) (*Identity, error) {
	i, err := p.GetClassified(ctx, id)
	if err != nil {
		return nil, err
	}

	return p.declassify(*i), nil
}

func (p *PoolSQL) GetClassified(ctx context.Context, id string) (*Identity, error) {
	i, err := p.get(ctx, "WHERE i.id = ?", []interface{}{id})
	if err != nil {
		if errors.Cause(err).Error() == herodot.ErrNotFound.Error() {
			return nil, herodot.ErrNotFound.WithTrace(err).WithReasonf(`Identity "%s" could not be found.`, id)
		}
		return nil, err
	}
	return i, nil
}

func (p *PoolSQL) Delete(ctx context.Context, id string) error {
	_, err := p.db.ExecContext(ctx, p.db.Rebind("DELETE FROM identity WHERE id = ?"), id)
	return sqlcon.HandleError(err)
}

func (p *PoolSQL) insert(ctx context.Context, tx *sqlx.Tx, i *Identity) error {
	columns, arguments := sqlxx.NamedInsertArguments(i)
	query := fmt.Sprintf(`INSERT INTO identity (%s) VALUES (%s)`, columns, arguments)
	if _, err := tx.NamedExecContext(context.Background(), p.db.Rebind(query), i); err != nil {
		return sqlcon.HandleError(err)
	}

	return p.insertCredentials(ctx, tx, i)
}

func (p *PoolSQL) insertCredentials(ctx context.Context, tx *sqlx.Tx, i *Identity) error {

	for method, cred := range i.Credentials {
		query := `INSERT INTO identity_credential (method, config, identity_pk) VALUES (
	?,
	?,
	(SELECT pk FROM identity WHERE id = ?))`
		if _, err := tx.ExecContext(ctx,
			p.db.Rebind(query),
			string(method),
			stringsx.Coalesce(string(cred.Config), "{}"),
			i.ID,
		); err != nil {
			return sqlcon.HandleError(err)
		}

		for _, identifier := range cred.Identifiers {
			query = `INSERT INTO identity_credential_identifier (identifier, identity_credential_pk) VALUES (
	?,
	(SELECT ic.pk FROM identity_credential as ic JOIN identity as i ON (i.pk = ic.identity_pk) WHERE ic.method = ? AND i.id = ?))`
			if _, err := tx.ExecContext(ctx,
				p.db.Rebind(query),
				identifier,
				string(method),
				i.ID,
			); err != nil {
				return sqlcon.HandleError(err)
			}
		}
	}
	return nil
}

func (p *PoolSQL) update(ctx context.Context, i *Identity, ct map[CredentialsType]Credentials, updateConfidential bool) (*Identity, error) {
	insert := p.augment(*i)
	insert.Credentials = ct

	if err := p.Validate(insert); err != nil {
		return nil, err
	}

	tx, err := p.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}

	if err := p.runUpdateTx(ctx, tx, i, updateConfidential); err != nil {
		if err := tx.Rollback(); err != nil {
			return nil, errors.WithStack(err)
		}
		return nil, errors.WithStack(err)
	}

	if err := tx.Commit(); err != nil {
		if err := tx.Rollback(); err != nil {
			return nil, errors.WithStack(err)
		}
		return nil, errors.WithStack(err)
	}

	return p.declassify(*i), nil
}

func (p *PoolSQL) runUpdateTx(ctx context.Context, tx *sqlx.Tx, i *Identity, updateConfidential bool) error {
	arguments := sqlxx.NamedUpdateArguments(i, "id")
	query := fmt.Sprintf(`UPDATE identity SET %s WHERE id=:id`, arguments)
	if _, err := tx.NamedExecContext(context.Background(), query, i); err != nil {
		return sqlcon.HandleError(err)
	}

	if !updateConfidential {
		return nil
	}

	if _, err := tx.ExecContext(ctx, p.db.Rebind(`DELETE FROM identity_credential as ic USING identity as i WHERE i.pk = ic.identity_pk AND i.id = ?`), i.ID); err != nil {
		return sqlcon.HandleError(err)
	}
	return p.insertCredentials(ctx, tx, i)
}

func (p *PoolSQL) get(ctx context.Context, where string, args []interface{}) (*Identity, error) {
	var rows []struct {
		ID              string          `db:"id"`
		TraitsSchemaURL string          `db:"traits_schema_url"`
		Traits          json.RawMessage `db:"traits"`
		Identifier      sql.NullString  `db:"identifier"`
		Method          sql.NullString  `db:"method"`
		Config          sql.NullString  `db:"config"`
	}

	query := fmt.Sprintf(`
SELECT
	i.id as id, i.traits_schema_url as traits_schema_url, i.traits as traits, ic.config as config, ic.method as method, ici.identifier as identifier
FROM identity as i
LEFT OUTER JOIN identity_credential as ic ON
	ic.identity_pk = i.pk
LEFT OUTER JOIN identity_credential_identifier as ici ON
	ici.identity_credential_pk = ic.pk
%s`, where)
	if err := sqlcon.HandleError(p.db.SelectContext(ctx, &rows, p.db.Rebind(query), args...)); err != nil {
		if errors.Cause(err) == sqlcon.ErrNoRows {
			return nil, errors.WithStack(herodot.ErrNotFound.WithReasonf(`Identity could not be found.`))
		}
		return nil, err
	}

	if len(rows) == 0 {
		return nil, errors.WithStack(herodot.ErrNotFound.WithReasonf(`Identity could not be found.`))
	}

	credentials := map[CredentialsType]Credentials{}
	for _, row := range rows {
		if !(row.Method.Valid && row.Identifier.Valid && row.Config.Valid) {
			continue
		}

		if c, ok := credentials[CredentialsType(row.Method.String)]; ok {
			c.Identifiers = append(c.Identifiers, row.Identifier.String)
			credentials[CredentialsType(row.Method.String)] = c
		} else {
			credentials[CredentialsType(row.Method.String)] = Credentials{
				ID:          CredentialsType(row.Method.String),
				Config:      json.RawMessage(row.Config.String),
				Identifiers: []string{row.Identifier.String},
			}
		}
	}

	return &Identity{
		ID:              rows[0].ID,
		TraitsSchemaURL: rows[0].TraitsSchemaURL,
		Traits:          rows[0].Traits,
		Credentials:     credentials,
	}, nil
}
