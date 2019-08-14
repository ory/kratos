package identity

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/ory/x/sqlcon"
	"github.com/ory/x/sqlxx"
	"github.com/pkg/errors"

	"github.com/ory/herodot"

	"github.com/ory/hive/driver/configuration"
	"github.com/ory/hive/schema"
)

var _ Pool = new(PoolSQL)

type (
	PoolSQL struct {
		*abstractPool
		db *sqlx.DB
	}
)

func (i *identitySQL) toIdentity() *Identity {
	return &Identity{
		ID:              i.ID,
		TraitsSchemaURL: i.TraitsSchemaURL,
		Traits:          i.Traits,
	}
}

func NewPoolSQL(c configuration.Provider, d ValidationProvider, db *sqlx.DB) *PoolSQL {
	return &PoolSQL{abstractPool: newAbstractPool(c, d), db: db,}
}

// FindByCredentialsIdentifier returns an identity by querying for it's credential identifiers.
func (p *PoolSQL) FindByCredentialsIdentifier(ctx context.Context, ct CredentialsType, match string) (*Identity, *Credentials, error) {
	i, err := p.get(ctx, "WHERE ici.identifier = ? AND ici.method = ?", []interface{}{match, string(ct)})
	if err != nil {
		return nil, nil, err
	}

	creds, ok := i.Credentials[ct]
	if !ok {
		return nil, nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("The SQL adapter failed to return the appropriate credentials_type \"%s\". This is a bug in the code.", ct))

	}

	return p.declassify(i), &creds, nil
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

	query := "SELECT id, traits, traits_schema_url FROM identity LIMIT ? OFFSET ? ORDER BY pk"
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

func (p *PoolSQL) Update(_ context.Context, i *Identity) (*Identity, error) {
	var rows []struct {
		ID              string          `db:"id"`
		TraitsSchemaURL string          `db:"i.traits_schema_url"`
		Traits          json.RawMessage `db:"i.traits"`
		Identifier      string          `db:"ici.identifier"`
		Method          string          `db:"ic.method"`
		Config          string          `db:"ic.config"`
	}

	query := "SELECT id, traits, traits_schema_url FROM identity "

	insert := p.augment(*i)
	if err := p.Validate(insert); err != nil {
		return nil, err
	}

	if p.hasConflictingCredentials(insert) {
		return nil, errors.WithStack(schema.NewDuplicateCredentialsError())
	}

	p.RLock()
	for k, ii := range p.is {
		if ii.ID == insert.ID {
			p.RUnlock()

			p.Lock()
			p.is[k] = *insert
			p.Unlock()

			return p.declassify(*insert), nil
		}
	}
	p.RUnlock()
	return nil, errors.WithStack(herodot.ErrNotFound.WithReasonf("Identity with identifier %s does not exist.", i.ID))
}

func (p *PoolSQL) Get(ctx context.Context, id string) (*Identity, error) {
	i, err := p.GetClassified(ctx, id)
	if err != nil {
		return nil, err
	}

	return p.declassify(*i), nil
}

func (p *PoolSQL) GetClassified(ctx context.Context, id string) (*Identity, error) {
	return p.get(ctx, "WHERE i.ID = ?", []interface{}{id})
}

func (p *PoolSQL) Delete(ctx context.Context, id string) error {
	_, err := p.db.ExecContext(ctx, p.db.Rebind("DELETE FROM identity WHERE id = ?"), id)
	return sqlcon.HandleError(err)
}

func (p *PoolSQL) insert(ctx context.Context, tx *sqlx.Tx, i *Identity) error {
	columns, arguments := sqlxx.NamedInsertArguments(i)
	query := fmt.Sprintf(`INSERT INTO identity (%s) VALUES (%s)`, columns, arguments)
	if _, err := tx.ExecContext(context.Background(), p.db.Rebind(query), i); err != nil {
		return sqlcon.HandleError(err)
	}

	for method, cred := range i.Credentials {
		query = `INSERT INTO identity_credentials (method, options, identity_pk) VALUES (
	?,
	?,
	(SELECT pk FROM identity WHERE id = ?))`
		if _, err := tx.ExecContext(ctx,
			p.db.Rebind(query),
			string(method),
			string(cred.Options),
			i.ID,
		); err != nil {
			return sqlcon.HandleError(err)
		}

		for _, identifier := range cred.Identifiers {
			query = `INSERT INTO identity_credentials_identifiers (identifier, identity_credentials_pk) VALUES (
	?,
	(SELECT ic.pk FROM identity_credentials as ic JOIN identity as i ON (i.pk = ic.identity_pk) WHERE ic.method = ? AND i.id = ?))`
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

func (p *PoolSQL) update(ctx context.Context, tx *sqlx.Tx, i *Identity) error {
	arguments := sqlxx.NamedUpdateArguments(i, "id")
	query := fmt.Sprintf(`UPDATE identity SET (%s) WHERE id = :id`, arguments)

	if _, err := tx.ExecContext(context.Background(), p.db.Rebind(query), i); err != nil {
		return sqlcon.HandleError(err)
	}

	for method, cred := range i.Credentials {
		var ic = &struct {
			Method  string `db:"method"`
			Options string `db:"options"`
		}{Method:string(method), Options: string(cred.Options)}

		query = `UPDATE identity_credentials SET %s	WHERE SELECT pk FROM identity WHERE id = ?))`

		arguments = sqlxx.NamedUpdateArguments(ic)

		query = `INSERT INTO identity_credentials (method, options, identity_pk) VALUES (
	?,
	?,
	(SELECT pk FROM identity WHERE id = ?))`
		if _, err := tx.ExecContext(ctx,
			p.db.Rebind(query),
			string(method),
			string(cred.Options),
			i.ID,
		); err != nil {
			return sqlcon.HandleError(err)
		}

		for _, identifier := range cred.Identifiers {
			query = `INSERT INTO identity_credentials_identifiers (identifier, identity_credentials_pk) VALUES (
	?,
	(SELECT ic.pk FROM identity_credentials as ic JOIN identity as i ON (i.pk = ic.identity_pk) WHERE ic.method = ? AND i.id = ?))`
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

func (p *PoolSQL) get(ctx context.Context, where string, args []interface{}) (*Identity, error) {
	var rows []struct {
		ID              string          `db:"i.id"`
		TraitsSchemaURL string          `db:"i.traits_schema_url"`
		Traits          json.RawMessage `db:"i.traits"`
		Identifier      string          `db:"ici.identifier"`
		Method          string          `db:"ic.method"`
		Config          string          `db:"ic.config"`
	}

	query := fmt.Sprintf("SELECT i.id, i.traits_schema_url, i.traits, ic.config, ic.method, ici.identifier FROM identity as i JOIN identity_credentials_identifiers as ici ON (ici.identity_credentials_pk = ic.pk), identity_credentials as ic ON (ic.identity_pk = i.pk) %s", where)
	if err := p.db.SelectContext(ctx, &rows, p.db.Rebind(query), args...); err != nil {
		return nil, sqlcon.HandleError(err)
	}

	if len(rows) == 0 {
		return nil, errors.WithStack(herodot.ErrNotFound.WithReasonf("No identity matching the credentials identifiers"))
	}

	credentials := map[CredentialsType]Credentials{}
	for _, row := range rows {
		if c, ok := credentials[CredentialsType(row.Method)]; ok {
			c.Identifiers = append(c.Identifiers, row.Identifier)
			credentials[CredentialsType(row.Method)] = c
		} else {
			credentials[CredentialsType(row.Method)] = Credentials{
				ID:          CredentialsType(row.Method),
				Options:     json.RawMessage(row.Config),
				Identifiers: []string{row.Identifier},
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
