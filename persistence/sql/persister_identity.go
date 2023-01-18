// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package sql

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"go.opentelemetry.io/otel/attribute"

	"github.com/ory/x/otelx"

	"github.com/ory/jsonschema/v3"
	"github.com/ory/x/sqlxx"

	"github.com/ory/kratos/otp"
	"github.com/ory/kratos/x"

	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
	"github.com/pkg/errors"

	"github.com/ory/herodot"
	"github.com/ory/x/errorsx"
	"github.com/ory/x/sqlcon"

	"github.com/ory/kratos/identity"
)

var _ identity.Pool = new(Persister)
var _ identity.PrivilegedPool = new(Persister)

func (p *Persister) ListVerifiableAddresses(ctx context.Context, page, itemsPerPage int) (a []identity.VerifiableAddress, err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.ListVerifiableAddresses")
	defer span.End()

	if err := p.GetConnection(ctx).Where("nid = ?", p.NetworkID(ctx)).Order("id DESC").Paginate(page, x.MaxItemsPerPage(itemsPerPage)).All(&a); err != nil {
		return nil, sqlcon.HandleError(err)
	}

	return a, err
}

func (p *Persister) ListRecoveryAddresses(ctx context.Context, page, itemsPerPage int) (a []identity.RecoveryAddress, err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.ListRecoveryAddresses")
	defer span.End()

	if err := p.GetConnection(ctx).Where("nid = ?", p.NetworkID(ctx)).Order("id DESC").Paginate(page, x.MaxItemsPerPage(itemsPerPage)).All(&a); err != nil {
		return nil, sqlcon.HandleError(err)
	}

	return a, err
}

func stringToLowerTrim(match string) string {
	return strings.ToLower(strings.TrimSpace(match))
}

func (p *Persister) normalizeIdentifier(ct identity.CredentialsType, match string) string {
	switch ct {
	case identity.CredentialsTypeLookup:
		// lookup credentials are case-sensitive
		return match
	case identity.CredentialsTypeTOTP:
		// totp credentials are case-sensitive
		return match
	case identity.CredentialsTypeOIDC:
		// OIDC credentials are case-sensitive
		return match
	case identity.CredentialsTypePassword:
		fallthrough
	case identity.CredentialsTypeWebAuthn:
		return stringToLowerTrim(match)
	}
	return match
}

func (p *Persister) FindByCredentialsIdentifier(ctx context.Context, ct identity.CredentialsType, match string) (*identity.Identity, *identity.Credentials, error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.FindByCredentialsIdentifier")
	defer span.End()

	nid := p.NetworkID(ctx)

	var find struct {
		IdentityID uuid.UUID `db:"identity_id"`
	}

	// Force case-insensitivity and trimming for identifiers
	match = p.normalizeIdentifier(ct, match)

	// #nosec G201
	if err := p.GetConnection(ctx).RawQuery(fmt.Sprintf(`SELECT
    ic.identity_id
FROM %s ic
         INNER JOIN %s ict on ic.identity_credential_type_id = ict.id
         INNER JOIN %s ici on ic.id = ici.identity_credential_id AND ici.identity_credential_type_id = ict.id
WHERE ici.identifier = ?
  AND ic.nid = ?
  AND ici.nid = ?
  AND ict.name = ?`,
		"identity_credentials",
		"identity_credential_types",
		"identity_credential_identifiers",
	),
		match,
		nid,
		nid,
		ct,
	).First(&find); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil, sqlcon.HandleError(err) // herodot.ErrNotFound.WithTrace(err).WithReasonf(`No identity matching credentials identifier "%s" could be found.`, match)
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

func (p *Persister) findIdentityCredentialsType(ctx context.Context, ct identity.CredentialsType) (*identity.CredentialsTypeTable, error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.findIdentityCredentialsType")
	defer span.End()

	var m identity.CredentialsTypeTable
	if err := p.GetConnection(ctx).Where("name = ?", ct).First(&m); err != nil {
		return nil, sqlcon.HandleError(err)
	}
	return &m, nil
}

func (p *Persister) createIdentityCredentials(ctx context.Context, i *identity.Identity) error {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.createIdentityCredentials")
	defer span.End()

	c := p.GetConnection(ctx)

	nid := p.NetworkID(ctx)
	for k := range i.Credentials {
		cred := i.Credentials[k]

		if len(cred.Config) == 0 {
			cred.Config = sqlxx.JSONRawMessage("{}")
		}

		ct, err := p.findIdentityCredentialsType(ctx, cred.Type)
		if err != nil {
			return err
		}

		cred.IdentityID = i.ID
		cred.NID = nid
		cred.IdentityCredentialTypeID = ct.ID
		if err := c.Create(&cred); err != nil {
			return sqlcon.HandleError(err)
		}

		for _, ids := range cred.Identifiers {
			// Force case-insensitivity and trimming for identifiers
			ids = p.normalizeIdentifier(cred.Type, ids)

			if len(ids) == 0 {
				return errors.WithStack(herodot.ErrInternalServerError.WithReasonf("Unable to create identity credentials with missing or empty identifier."))
			}

			if err := c.Create(&identity.CredentialIdentifier{
				Identifier:                ids,
				IdentityCredentialsID:     cred.ID,
				IdentityCredentialsTypeID: ct.ID,
				NID:                       p.NetworkID(ctx),
			}); err != nil {
				return sqlcon.HandleError(err)
			}
		}

		i.Credentials[k] = cred
	}

	return nil
}

func (p *Persister) createVerifiableAddresses(ctx context.Context, i *identity.Identity) error {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.createVerifiableAddresses")
	defer span.End()

	for k := range i.VerifiableAddresses {
		if err := p.GetConnection(ctx).Create(&i.VerifiableAddresses[k]); err != nil {
			return err
		}
	}
	return nil
}

func updateAssociation[T interface {
	Hash() string
}](ctx context.Context, p *Persister, i *identity.Identity, inID []T) error {
	var inDB []T
	if err := p.GetConnection(ctx).
		Where("identity_id = ? AND nid = ?", i.ID, p.NetworkID(ctx)).
		Order("id ASC").
		All(&inDB); err != nil {

		return sqlcon.HandleError(err)
	}

	newAssocs := make(map[string]*T)
	oldAssocs := make(map[string]*T)
	for i, a := range inID {
		newAssocs[a.Hash()] = &inID[i]
	}
	for i, a := range inDB {
		oldAssocs[a.Hash()] = &inDB[i]
	}

	// Subtle: we delete the old associations from the DB first, because else
	// they could cause UNIQUE constraints to fail on insert.
	for h, a := range oldAssocs {
		if _, found := newAssocs[h]; found {
			newAssocs[h] = nil // Ignore associations that are already in the db.
		} else {
			if err := p.GetConnection(ctx).Destroy(a); err != nil {
				return sqlcon.HandleError(err)
			}
		}
	}

	for _, a := range newAssocs {
		if a != nil {
			if err := p.GetConnection(ctx).Create(a); err != nil {
				return sqlcon.HandleError(err)
			}
		}
	}

	return nil
}

func (p *Persister) normalizeAllAddressess(ctx context.Context, id *identity.Identity) {
	p.normalizeRecoveryAddresses(ctx, id)
	p.normalizeVerifiableAddresses(ctx, id)
}

func (p *Persister) normalizeVerifiableAddresses(ctx context.Context, id *identity.Identity) {
	for k := range id.VerifiableAddresses {
		id.VerifiableAddresses[k].IdentityID = id.ID
		id.VerifiableAddresses[k].NID = p.NetworkID(ctx)
		id.VerifiableAddresses[k].Value = stringToLowerTrim(id.VerifiableAddresses[k].Value)
	}
}

func (p *Persister) normalizeRecoveryAddresses(ctx context.Context, id *identity.Identity) {
	for k := range id.RecoveryAddresses {
		id.RecoveryAddresses[k].IdentityID = id.ID
		id.RecoveryAddresses[k].NID = p.NetworkID(ctx)
		id.RecoveryAddresses[k].Value = stringToLowerTrim(id.RecoveryAddresses[k].Value)
	}
}

func (p *Persister) createRecoveryAddresses(ctx context.Context, i *identity.Identity) error {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.createRecoveryAddresses")
	defer span.End()

	for k := range i.RecoveryAddresses {
		if err := p.GetConnection(ctx).Create(&i.RecoveryAddresses[k]); err != nil {
			return err
		}
	}
	return nil
}

func (p *Persister) CountIdentities(ctx context.Context) (int64, error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.CountIdentities")
	defer span.End()

	count, err := p.c.WithContext(ctx).Where("nid = ?", p.NetworkID(ctx)).Count(new(identity.Identity))
	if err != nil {
		return 0, sqlcon.HandleError(err)
	}
	return int64(count), nil
}

func (p *Persister) CreateIdentity(ctx context.Context, i *identity.Identity) error {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.CreateIdentity")
	defer span.End()

	i.NID = p.NetworkID(ctx)

	if i.SchemaID == "" {
		i.SchemaID = p.r.Config().DefaultIdentityTraitsSchemaID(ctx)
	}

	stateChangedAt := sqlxx.NullTime(time.Now())
	i.StateChangedAt = &stateChangedAt
	if i.State == "" {
		i.State = identity.StateActive
	}

	if len(i.Traits) == 0 {
		i.Traits = identity.Traits("{}")
	}

	if err := p.injectTraitsSchemaURL(ctx, i); err != nil {
		return err
	}

	if err := p.validateIdentity(ctx, i); err != nil {
		return err
	}

	return p.Transaction(ctx, func(ctx context.Context, tx *pop.Connection) error {
		if err := tx.Create(i); err != nil {
			return sqlcon.HandleError(err)
		}

		p.normalizeAllAddressess(ctx, i)

		if err := p.createVerifiableAddresses(ctx, i); err != nil {
			return sqlcon.HandleError(err)
		}

		if err := p.createRecoveryAddresses(ctx, i); err != nil {
			return sqlcon.HandleError(err)
		}

		return p.createIdentityCredentials(ctx, i)
	})
}

func (p *Persister) HydrateIdentityAssociations(ctx context.Context, i *identity.Identity, expand identity.Expandables) (err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.HydrateIdentityAssociations")
	defer otelx.End(span, &err)

	con := p.GetConnection(ctx)
	if err := con.Load(i, expand.ToEager()...); err != nil {
		return err
	}

	if err := i.AfterEagerFind(con); err != nil {
		return err
	}

	return p.injectTraitsSchemaURL(ctx, i)
}

func (p *Persister) ListIdentities(ctx context.Context, expand identity.Expandables, page, perPage int) (res []identity.Identity, err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.ListIdentities")
	defer otelx.End(span, &err)
	span.SetAttributes(
		attribute.Int("page", page),
		attribute.Int("per_page", perPage),
		attribute.StringSlice("expand", expand.ToEager()),
		attribute.String("network.id", p.NetworkID(ctx).String()),
	)

	is := make([]identity.Identity, 0)

	con := p.GetConnection(ctx)
	query := con.
		Where("nid = ?", p.NetworkID(ctx)).
		Paginate(page, perPage).
		Order("id DESC")

	if len(expand) > 0 {
		query = query.EagerPreload(expand.ToEager()...)
	}

	/* #nosec G201 TableName is static */
	if err := sqlcon.HandleError(query.All(&is)); err != nil {
		return nil, err
	}

	schemaCache := map[string]string{}
	for k := range is {
		i := &is[k]

		if u, ok := schemaCache[i.SchemaID]; ok {
			i.SchemaURL = u
		} else {
			if err := p.injectTraitsSchemaURL(ctx, i); err != nil {
				return nil, err
			}
			schemaCache[i.SchemaID] = i.SchemaURL
		}

		is[k] = *i
	}

	return is, nil
}

func (p *Persister) UpdateIdentity(ctx context.Context, i *identity.Identity) error {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.UpdateIdentity")
	defer span.End()

	if err := p.validateIdentity(ctx, i); err != nil {
		return err
	}

	i.NID = p.NetworkID(ctx)
	return sqlcon.HandleError(p.Transaction(ctx, func(ctx context.Context, tx *pop.Connection) error {
		if count, err := tx.Where("id = ? AND nid = ?", i.ID, p.NetworkID(ctx)).Count(i); err != nil {
			return err
		} else if count == 0 {
			return sql.ErrNoRows
		}

		p.normalizeAllAddressess(ctx, i)
		if err := updateAssociation(ctx, p, i, i.RecoveryAddresses); err != nil {
			return err
		}
		if err := updateAssociation(ctx, p, i, i.VerifiableAddresses); err != nil {
			return err
		}

		/* #nosec G201 TableName is static */
		if err := tx.RawQuery(
			fmt.Sprintf(
				`DELETE FROM %s WHERE identity_id = ? AND nid = ?`,
				new(identity.Credentials).TableName(ctx)),
			i.ID, p.NetworkID(ctx)).Exec(); err != nil {

			return sqlcon.HandleError(err)
		}

		if err := p.update(WithTransaction(ctx, tx), i); err != nil {
			return err
		}

		return p.createIdentityCredentials(ctx, i)
	}))
}

func (p *Persister) DeleteIdentity(ctx context.Context, id uuid.UUID) error {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.DeleteIdentity")
	defer span.End()

	return p.delete(ctx, new(identity.Identity), id)
}

func (p *Persister) GetIdentity(ctx context.Context, id uuid.UUID, expand identity.Expandables) (res *identity.Identity, err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.GetIdentity")
	defer otelx.End(span, &err)

	span.SetAttributes(
		attribute.String("identity.id", id.String()),
		attribute.StringSlice("expand", expand.ToEager()),
		attribute.String("network.id", p.NetworkID(ctx).String()),
	)

	con := p.GetConnection(ctx)
	query := con.Where("id = ? AND nid = ?", id, p.NetworkID(ctx))
	if len(expand) > 0 {
		query = query.EagerPreload(expand.ToEager()...)
	}

	var i identity.Identity
	if err := query.First(&i); err != nil {
		return nil, sqlcon.HandleError(err)
	}

	if err := p.injectTraitsSchemaURL(ctx, &i); err != nil {
		return nil, err
	}

	return &i, nil
}

func (p *Persister) GetIdentityConfidential(ctx context.Context, id uuid.UUID) (res *identity.Identity, err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.GetIdentityConfidential")
	defer otelx.End(span, &err)

	return p.GetIdentity(ctx, id, identity.ExpandEverything)
}

func (p *Persister) FindVerifiableAddressByValue(ctx context.Context, via identity.VerifiableAddressType, value string) (*identity.VerifiableAddress, error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.FindVerifiableAddressByValue")
	defer span.End()

	var address identity.VerifiableAddress
	if err := p.GetConnection(ctx).Where("nid = ? AND via = ? AND value = ?", p.NetworkID(ctx), via, stringToLowerTrim(value)).First(&address); err != nil {
		return nil, sqlcon.HandleError(err)
	}

	return &address, nil
}

func (p *Persister) FindRecoveryAddressByValue(ctx context.Context, via identity.RecoveryAddressType, value string) (*identity.RecoveryAddress, error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.FindRecoveryAddressByValue")
	defer span.End()

	var address identity.RecoveryAddress
	if err := p.GetConnection(ctx).Where("nid = ? AND via = ? AND value = ?", p.NetworkID(ctx), via, stringToLowerTrim(value)).First(&address); err != nil {
		return nil, sqlcon.HandleError(err)
	}

	return &address, nil
}

func (p *Persister) VerifyAddress(ctx context.Context, code string) error {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.VerifyAddress")
	defer span.End()
	newCode, err := otp.New()
	if err != nil {
		return err
	}

	count, err := p.GetConnection(ctx).RawQuery(
		/* #nosec G201 TableName is static */
		fmt.Sprintf(
			"UPDATE %s SET status = ?, verified = true, verified_at = ?, code = ? WHERE nid = ? AND code = ? AND expires_at > ?",
			new(identity.VerifiableAddress).TableName(ctx),
		),
		identity.VerifiableAddressStatusCompleted,
		time.Now().UTC().Round(time.Second),
		newCode,
		p.NetworkID(ctx),
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
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.UpdateVerifiableAddress")
	defer span.End()

	address.NID = p.NetworkID(ctx)
	address.Value = stringToLowerTrim(address.Value)
	return p.update(ctx, address)
}

func (p *Persister) validateIdentity(ctx context.Context, i *identity.Identity) error {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.validateIdentity")
	defer span.End()

	if err := p.r.IdentityValidator().ValidateWithRunner(ctx, i); err != nil {
		if _, ok := errorsx.Cause(err).(*jsonschema.ValidationError); ok {
			return errors.WithStack(herodot.ErrBadRequest.WithReasonf("%s", err))
		}
		return err
	}

	return nil
}

func (p *Persister) injectTraitsSchemaURL(ctx context.Context, i *identity.Identity) error {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.injectTraitsSchemaURL")
	defer span.End()

	ss, err := p.r.IdentityTraitsSchemas(ctx)
	if err != nil {
		return err
	}
	s, err := ss.GetByID(i.SchemaID)
	if err != nil {
		return errors.WithStack(herodot.ErrInternalServerError.WithReasonf(
			`The JSON Schema "%s" for this identity's traits could not be found.`, i.SchemaID))
	}
	i.SchemaURL = s.SchemaURL(p.r.Config().SelfPublicURL(ctx)).String()
	return nil
}
