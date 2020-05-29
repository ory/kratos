package sql

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/ory/jsonschema/v3"

	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/otp"

	"github.com/gobuffalo/pop/v5"
	"github.com/gofrs/uuid"
	"github.com/pkg/errors"

	"github.com/ory/herodot"
	"github.com/ory/x/errorsx"
	"github.com/ory/x/sqlcon"

	"github.com/ory/kratos/identity"
)

var _ identity.Pool = new(Persister)
var _ identity.PrivilegedPool = new(Persister)

func (p *Persister) FindByCredentialsIdentifier(ctx context.Context, ct identity.CredentialsType, match string) (*identity.Identity, *identity.Credentials, error) {
	var cts []identity.CredentialsTypeTable
	if err := p.GetConnection(ctx).All(&cts); err != nil {
		return nil, nil, sqlcon.HandleError(err)
	}

	var find struct {
		IdentityID uuid.UUID `db:"identity_id"`
	}

	if err := p.GetConnection(ctx).RawQuery(`SELECT
    ic.identity_id
FROM identity_credentials ic
         INNER JOIN identity_credential_types ict on ic.identity_credential_type_id = ict.id
         INNER JOIN identity_credential_identifiers ici on ic.id = ici.identity_credential_id
WHERE ici.identifier = ?
  AND ict.name = ?`, match, ct).First(&find); err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, nil, herodot.ErrNotFound.WithTrace(err).WithReasonf(`No identity matching credentials identifier "%s" could be found.`, match)
		}

		return nil, nil, sqlcon.HandleError(err)
	}

	i, err := p.GetIdentityConfidential(ctx, find.IdentityID)
	if err != nil {
		return nil, nil, err
	}

	creds, ok := i.GetCredentials(ct)
	if !ok {
		return nil, nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("The SQL adapter failed to return the appropriate credentials_type \"%s\". This is a bug in the code.", ct))
	}

	return i.CopyWithoutCredentials(), creds, nil
}

func findOrCreateIdentityCredentialsType(_ context.Context, tx *pop.Connection, ct identity.CredentialsType) (*identity.CredentialsTypeTable, error) {
	var m identity.CredentialsTypeTable
	if err := tx.Where("name = ?", ct).First(&m); err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			m.Name = ct
			if err := sqlcon.HandleError(tx.Create(&m)); err != nil {
				return nil, err
			}
			return &m, nil
		}
		return nil, sqlcon.HandleError(err)
	}

	return &m, nil
}

func createIdentityCredentials(ctx context.Context, tx *pop.Connection, i *identity.Identity) error {
	for k, cred := range i.Credentials {
		cred.IdentityID = i.ID
		if len(cred.Config) == 0 {
			cred.Config = json.RawMessage("{}")
		}

		ct, err := findOrCreateIdentityCredentialsType(ctx, tx, cred.Type)
		if err != nil {
			return err
		}

		cred.CredentialTypeID = ct.ID
		if err := tx.Create(&cred); err != nil {
			return err
		}

		for _, ids := range cred.Identifiers {
			// Force case-insensitivity for email addresses
			if strings.Contains(ids, "@") && cred.Type == identity.CredentialsTypePassword {
				ids = strings.ToLower(ids)
			}

			if len(ids) == 0 {
				return errors.WithStack(herodot.ErrInternalServerError.WithReasonf("Unable to create identity credentials with missing or empty identifier."))
			}

			ci := &identity.CredentialIdentifier{
				Identifier:            ids,
				IdentityCredentialsID: cred.ID,
			}
			if err := tx.Create(ci); err != nil {
				return err
			}
		}

		i.Credentials[k] = cred
	}

	return nil
}

func createVerifiableAddresses(ctx context.Context, tx *pop.Connection, i *identity.Identity) error {
	for k := range i.VerifiableAddresses {
		i.VerifiableAddresses[k].IdentityID = i.ID
		if err := tx.Create(&i.VerifiableAddresses[k]); err != nil {
			return err
		}
	}
	return nil
}

func createRecoveryAddresses(ctx context.Context, tx *pop.Connection, i *identity.Identity) error {
	for k := range i.RecoveryAddresses {
		i.RecoveryAddresses[k].IdentityID = i.ID
		if err := tx.Create(&i.RecoveryAddresses[k]); err != nil {
			return err
		}
	}
	return nil
}

func (p *Persister) CreateIdentity(ctx context.Context, i *identity.Identity) error {
	if i.TraitsSchemaID == "" {
		i.TraitsSchemaID = configuration.DefaultIdentityTraitsSchemaID
	}

	if len(i.Traits) == 0 {
		i.Traits = identity.Traits("{}")
	}

	if err := p.injectTraitsSchemaURL(i); err != nil {
		return err
	}

	if err := p.validateIdentity(i); err != nil {
		return err
	}

	return sqlcon.HandleError(p.Transaction(ctx, func(tx *pop.Connection) error {
		ctx := WithTransaction(ctx, tx)

		if err := tx.Create(i); err != nil {
			return err
		}

		if err := createVerifiableAddresses(ctx, tx, i); err != nil {
			return err
		}

		if err := createRecoveryAddresses(ctx, tx, i); err != nil {
			return err
		}

		return createIdentityCredentials(ctx, tx, i)
	}))
}

func (p *Persister) ListIdentities(ctx context.Context, limit, offset int) ([]identity.Identity, error) {
	is := make([]identity.Identity, 0)

	/* #nosec G201 TableName is static */
	if err := sqlcon.HandleError(p.GetConnection(ctx).
		RawQuery(fmt.Sprintf("SELECT * FROM %s LIMIT ? OFFSET ?", new(identity.Identity).TableName()), limit, offset).
		Eager("VerifiableAddresses", "RecoveryAddresses").All(&is)); err != nil {
		return nil, err
	}

	for i := range is {
		if err := p.injectTraitsSchemaURL(&(is[i])); err != nil {
			return nil, err
		}
	}

	return is, nil
}

func (p *Persister) UpdateIdentity(ctx context.Context, i *identity.Identity) error {
	if err := p.validateIdentity(i); err != nil {
		return err
	}

	return sqlcon.HandleError(p.Transaction(ctx, func(tx *pop.Connection) error {
		ctx := WithTransaction(ctx, tx)
		if count, err := tx.Where("id = ?", i.ID).Count(i); err != nil {
			return err
		} else if count == 0 {
			return sql.ErrNoRows
		}

		for _, tn := range []string{
			new(identity.Credentials).TableName(),
			new(identity.VerifiableAddress).TableName(),
			new(identity.RecoveryAddress).TableName(),
		} {
			/* #nosec G201 TableName is static */
			if err := tx.RawQuery(fmt.Sprintf(
				`DELETE FROM %s WHERE identity_id = ?`, tn), i.ID).Exec(); err != nil {
				return err
			}
		}

		if err := tx.Update(i); err != nil {
			return err
		}

		if err := createVerifiableAddresses(ctx, tx, i); err != nil {
			return err
		}

		if err := createRecoveryAddresses(ctx, tx, i); err != nil {
			return err
		}

		return createIdentityCredentials(ctx, tx, i)
	}))
}

func (p *Persister) DeleteIdentity(ctx context.Context, id uuid.UUID) error {
	/* #nosec G201 TableName is static */
	count, err := p.GetConnection(ctx).RawQuery(fmt.Sprintf("DELETE FROM %s WHERE id = ?", new(identity.Identity).TableName()), id).ExecWithCount()
	if err != nil {
		return sqlcon.HandleError(err)
	}
	if count == 0 {
		return sqlcon.ErrNoRows
	}
	return nil
}

func (p *Persister) GetIdentity(ctx context.Context, id uuid.UUID) (*identity.Identity, error) {
	var i identity.Identity
	if err := p.GetConnection(ctx).Eager("VerifiableAddresses", "RecoveryAddresses").Find(&i, id); err != nil {
		return nil, sqlcon.HandleError(err)
	}
	i.Credentials = nil
	if err := p.injectTraitsSchemaURL(&i); err != nil {
		return nil, err
	}

	return &i, nil
}

func (p *Persister) GetIdentityConfidential(ctx context.Context, id uuid.UUID) (*identity.Identity, error) {
	var i identity.Identity
	if err := p.GetConnection(ctx).Eager().Find(&i, id); err != nil {
		return nil, sqlcon.HandleError(err)
	}

	var cts []identity.CredentialsTypeTable
	if err := p.GetConnection(ctx).All(&cts); err != nil {
		return nil, sqlcon.HandleError(err)
	}

	i.Credentials = map[identity.CredentialsType]identity.Credentials{}
	for _, creds := range i.CredentialsCollection {
		var cs identity.CredentialIdentifierCollection
		if err := p.GetConnection(ctx).Where("identity_credential_id = ?", creds.ID).All(&cs); err != nil {
			return nil, sqlcon.HandleError(err)
		}

		creds.CredentialIdentifierCollection = nil
		creds.Identifiers = make([]string, len(cs))
		for k := range cs {
			for _, ct := range cts {
				if ct.ID == creds.CredentialTypeID {
					creds.Type = ct.Name
				}
			}
			creds.Identifiers[k] = cs[k].Identifier
		}
		i.Credentials[creds.Type] = creds
	}
	i.CredentialsCollection = nil
	if err := p.injectTraitsSchemaURL(&i); err != nil {
		return nil, err
	}

	return &i, nil
}

func (p *Persister) FindVerifiableAddressByValue(ctx context.Context, via identity.VerifiableAddressType, value string) (*identity.VerifiableAddress, error) {
	var address identity.VerifiableAddress
	if err := p.GetConnection(ctx).Where("via = ? AND value = ?", via, value).First(&address); err != nil {
		return nil, sqlcon.HandleError(err)
	}

	return &address, nil
}

func (p *Persister) FindRecoveryAddressByValue(ctx context.Context, via identity.RecoveryAddressType, value string) (*identity.RecoveryAddress, error) {
	var address identity.RecoveryAddress
	if err := p.GetConnection(ctx).Where("via = ? AND value = ?", via, value).First(&address); err != nil {
		return nil, sqlcon.HandleError(err)
	}

	return &address, nil
}

func (p *Persister) VerifyAddress(ctx context.Context, code string) error {
	newCode, err := otp.New()
	if err != nil {
		return err
	}

	count, err := p.GetConnection(ctx).RawQuery(
		/* #nosec G201 TableName is static */
		fmt.Sprintf(
			"UPDATE %s SET status = ?, verified = true, verified_at = ?, code = ? WHERE code = ? AND expires_at > ?",
			new(identity.VerifiableAddress).TableName(),
		),
		identity.VerifiableAddressStatusCompleted,
		time.Now().UTC().Round(time.Second),
		newCode,
		code,
		time.Now().UTC(),
	).ExecWithCount()
	if err != nil {
		return sqlcon.HandleError(err)
	}

	if count == 0 {
		return sqlcon.HandleError(sqlcon.ErrNoRows)
	}

	return nil
}

func (p *Persister) UpdateVerifiableAddress(ctx context.Context, address *identity.VerifiableAddress) error {
	return sqlcon.HandleError(p.GetConnection(ctx).Update(address))
}

func (p *Persister) validateIdentity(i *identity.Identity) error {
	if err := p.r.IdentityValidator().ValidateWithRunner(i); err != nil {
		if _, ok := errorsx.Cause(err).(*jsonschema.ValidationError); ok {
			return errors.WithStack(herodot.ErrBadRequest.WithReasonf("%s", err))
		}
		return err
	}

	return nil
}

func (p *Persister) injectTraitsSchemaURL(i *identity.Identity) error {
	s, err := p.r.IdentityTraitsSchemas().GetByID(i.TraitsSchemaID)
	if err != nil {
		return errors.WithStack(herodot.ErrInternalServerError.WithReasonf(
			`The JSON Schema "%s" for this identity's traits could not be found.`, i.TraitsSchemaID))
	}
	i.TraitsSchemaURL = s.SchemaURL(p.cf.SelfPublicURL()).String()
	return nil
}
