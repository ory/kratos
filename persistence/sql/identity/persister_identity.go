// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package identity

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/ory/x/contextx"
	"github.com/ory/x/pointerx"
	"github.com/ory/x/popx"

	"golang.org/x/sync/errgroup"

	"go.opentelemetry.io/otel/attribute"

	"github.com/ory/x/otelx"

	"github.com/ory/jsonschema/v3"
	"github.com/ory/x/sqlxx"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/otp"
	"github.com/ory/kratos/persistence/sql/update"
	"github.com/ory/kratos/schema"
	"github.com/ory/kratos/x"

	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
	"github.com/pkg/errors"

	"github.com/ory/herodot"
	"github.com/ory/x/errorsx"
	"github.com/ory/x/sqlcon"
)

var _ identity.Pool = new(IdentityPersister)
var _ identity.PrivilegedPool = new(IdentityPersister)

type dependencies interface {
	schema.IdentityTraitsProvider
	identity.ValidationProvider
	x.LoggingProvider
	config.Provider
	contextx.Provider
	x.TracingProvider
}

type IdentityPersister struct {
	r   dependencies
	c   *pop.Connection
	nid uuid.UUID
}

func NewPersister(r dependencies, c *pop.Connection) *IdentityPersister {
	return &IdentityPersister{
		c: c,
		r: r,
	}
}

func (p *IdentityPersister) NetworkID(ctx context.Context) uuid.UUID {
	return p.r.Contextualizer().Network(ctx, p.nid)
}

func (p IdentityPersister) WithNetworkID(nid uuid.UUID) identity.PrivilegedPool {
	p.nid = nid
	return &p
}

func WithTransaction(ctx context.Context, tx *pop.Connection) context.Context {
	return popx.WithTransaction(ctx, tx)
}

func (p *IdentityPersister) Transaction(ctx context.Context, callback func(ctx context.Context, connection *pop.Connection) error) error {
	return popx.Transaction(ctx, p.c.WithContext(ctx), callback)
}

func (p *IdentityPersister) GetConnection(ctx context.Context) *pop.Connection {
	return popx.GetConnection(ctx, p.c.WithContext(ctx))
}

func (p *IdentityPersister) ListVerifiableAddresses(ctx context.Context, page, itemsPerPage int) (a []identity.VerifiableAddress, err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.ListVerifiableAddresses")
	defer otelx.End(span, &err)

	if err := p.GetConnection(ctx).Where("nid = ?", p.NetworkID(ctx)).Order("id DESC").Paginate(page, x.MaxItemsPerPage(itemsPerPage)).All(&a); err != nil {
		return nil, sqlcon.HandleError(err)
	}

	return a, nil
}

func (p *IdentityPersister) ListRecoveryAddresses(ctx context.Context, page, itemsPerPage int) (a []identity.RecoveryAddress, err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.ListRecoveryAddresses")
	defer otelx.End(span, &err)

	if err := p.GetConnection(ctx).Where("nid = ?", p.NetworkID(ctx)).Order("id DESC").Paginate(page, x.MaxItemsPerPage(itemsPerPage)).All(&a); err != nil {
		return nil, sqlcon.HandleError(err)
	}

	return a, nil
}

func stringToLowerTrim(match string) string {
	return strings.ToLower(strings.TrimSpace(match))
}

func NormalizeIdentifier(ct identity.CredentialsType, match string) string {
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

func (p *IdentityPersister) FindByCredentialsIdentifier(ctx context.Context, ct identity.CredentialsType, match string) (_ *identity.Identity, _ *identity.Credentials, err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.FindByCredentialsIdentifier")
	defer otelx.End(span, &err)

	nid := p.NetworkID(ctx)

	var find struct {
		IdentityID uuid.UUID `db:"identity_id"`
	}

	// Force case-insensitivity and trimming for identifiers
	match = NormalizeIdentifier(ct, match)

	if err := p.GetConnection(ctx).RawQuery(`
		SELECT
			ic.identity_id
		FROM identity_credentials ic
				INNER JOIN identity_credential_types ict
					ON ic.identity_credential_type_id = ict.id
				INNER JOIN identity_credential_identifiers ici
					ON ic.id = ici.identity_credential_id AND ici.identity_credential_type_id = ict.id
		WHERE ici.identifier = ?
		AND ic.nid = ?
		AND ici.nid = ?
		AND ict.name = ?
		LIMIT 1`, // pop doesn't understand how to add a limit clause to this query
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

func (p *IdentityPersister) findIdentityCredentialsType(ctx context.Context, ct identity.CredentialsType) (_ *identity.CredentialsTypeTable, err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.findIdentityCredentialsType")
	defer otelx.End(span, &err)

	var m identity.CredentialsTypeTable
	if err := p.GetConnection(ctx).Where("name = ?", ct).First(&m); err != nil {
		return nil, sqlcon.HandleError(err)
	}
	return &m, nil
}

func (p *IdentityPersister) createIdentityCredentials(ctx context.Context, i *identity.Identity) (err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.createIdentityCredentials")
	defer otelx.End(span, &err)

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
			ids = NormalizeIdentifier(cred.Type, ids)

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

func (p *IdentityPersister) createVerifiableAddresses(ctx context.Context, i *identity.Identity) (err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.createVerifiableAddresses")
	defer otelx.End(span, &err)

	for k := range i.VerifiableAddresses {
		if err := p.GetConnection(ctx).Create(&i.VerifiableAddresses[k]); err != nil {
			return err
		}
	}
	return nil
}

func updateAssociation[T interface {
	Hash() string
}](ctx context.Context, p *IdentityPersister, i *identity.Identity, inID []T) (err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.updateAssociation")
	defer otelx.End(span, &err)

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

func (p *IdentityPersister) normalizeAllAddressess(ctx context.Context, id *identity.Identity) {
	p.normalizeRecoveryAddresses(ctx, id)
	p.normalizeVerifiableAddresses(ctx, id)
}

func (p *IdentityPersister) normalizeVerifiableAddresses(ctx context.Context, id *identity.Identity) {
	for k := range id.VerifiableAddresses {
		v := id.VerifiableAddresses[k]

		v.IdentityID = id.ID
		v.NID = p.NetworkID(ctx)
		v.Value = stringToLowerTrim(v.Value)
		v.Via = x.Coalesce(v.Via, identity.AddressTypeEmail)
		if len(v.Status) == 0 {
			if v.Verified {
				v.Status = identity.VerifiableAddressStatusCompleted
			} else {
				v.Status = identity.VerifiableAddressStatusPending
			}
		}

		// If verified is true but no timestamp is set, we default to time.Now
		if v.Verified && (v.VerifiedAt == nil || time.Time(*v.VerifiedAt).IsZero()) {
			v.VerifiedAt = pointerx.Ptr(sqlxx.NullTime(time.Now()))
		}

		id.VerifiableAddresses[k] = v
	}
}

func (p *IdentityPersister) normalizeRecoveryAddresses(ctx context.Context, id *identity.Identity) {
	for k := range id.RecoveryAddresses {
		id.RecoveryAddresses[k].IdentityID = id.ID
		id.RecoveryAddresses[k].NID = p.NetworkID(ctx)
		id.RecoveryAddresses[k].Value = stringToLowerTrim(id.RecoveryAddresses[k].Value)
		id.RecoveryAddresses[k].Via = x.Coalesce(id.RecoveryAddresses[k].Via, identity.AddressTypeEmail)
	}
}

func (p *IdentityPersister) createRecoveryAddresses(ctx context.Context, i *identity.Identity) (err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.createRecoveryAddresses")
	defer otelx.End(span, &err)

	for k := range i.RecoveryAddresses {
		if err := p.GetConnection(ctx).Create(&i.RecoveryAddresses[k]); err != nil {
			return err
		}
	}
	return nil
}

func (p *IdentityPersister) CountIdentities(ctx context.Context) (n int64, err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.CountIdentities")
	defer otelx.End(span, &err)

	count, err := p.c.WithContext(ctx).Where("nid = ?", p.NetworkID(ctx)).Count(new(identity.Identity))
	if err != nil {
		return 0, sqlcon.HandleError(err)
	}
	return int64(count), nil
}

func (p *IdentityPersister) CreateIdentity(ctx context.Context, i *identity.Identity) (err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.CreateIdentity")
	defer otelx.End(span, &err)

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

	if err := p.InjectTraitsSchemaURL(ctx, i); err != nil {
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

func (p *IdentityPersister) HydrateIdentityAssociations(ctx context.Context, i *identity.Identity, expand identity.Expandables) (err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.HydrateIdentityAssociations")
	defer otelx.End(span, &err)

	var (
		con                 = p.GetConnection(ctx)
		nid                 = p.NetworkID(ctx)
		credentials         map[identity.CredentialsType]identity.Credentials
		verifiableAddresses []identity.VerifiableAddress
		recoveryAddresses   []identity.RecoveryAddress
	)

	eg, ctx := errgroup.WithContext(ctx)
	if expand.Has(identity.ExpandFieldRecoveryAddresses) {
		eg.Go(func() error {
			// We use WithContext to get a copy of the connection struct, which solves the race detector
			// from complaining incorrectly.
			//
			// https://github.com/gobuffalo/pop/issues/723
			if err := con.WithContext(ctx).
				Where("identity_id = ? AND nid = ?", i.ID, nid).
				Order("id ASC").
				All(&recoveryAddresses); err != nil {
				return sqlcon.HandleError(err)
			}
			return nil
		})
	}

	if expand.Has(identity.ExpandFieldVerifiableAddresses) {
		eg.Go(func() error {
			// We use WithContext to get a copy of the connection struct, which solves the race detector
			// from complaining incorrectly.
			//
			// https://github.com/gobuffalo/pop/issues/723
			if err := con.WithContext(ctx).
				Order("id ASC").
				Where("identity_id = ? AND nid = ?", i.ID, nid).All(&verifiableAddresses); err != nil {
				return sqlcon.HandleError(err)
			}
			return nil
		})
	}

	if expand.Has(identity.ExpandFieldCredentials) {
		eg.Go(func() (err error) {
			// We use WithContext to get a copy of the connection struct, which solves the race detector
			// from complaining incorrectly.
			//
			// https://github.com/gobuffalo/pop/issues/723
			con := con.WithContext(ctx)
			creds, err := QueryForCredentials(con, Where{
				"(identity_credentials.identity_id = ? AND identity_credentials.nid = ?)",
				[]interface{}{i.ID, nid},
			})
			if err != nil {
				return err
			}
			credentials = creds[i.ID]
			return
		})
	}

	if err := eg.Wait(); err != nil {
		return err
	}

	i.VerifiableAddresses = verifiableAddresses
	i.RecoveryAddresses = recoveryAddresses
	i.Credentials = credentials

	if err := i.Validate(); err != nil {
		return err
	}

	if err := identity.UpgradeCredentials(i); err != nil {
		return err
	}

	return p.InjectTraitsSchemaURL(ctx, i)
}

type queryCredentials struct {
	ID         uuid.UUID                `db:"cred_id"`
	IdentityID uuid.UUID                `db:"identity_id"`
	NID        uuid.UUID                `db:"nid"`
	Type       identity.CredentialsType `db:"cred_type"`
	TypeID     uuid.UUID                `db:"cred_type_id"`
	Identifier string                   `db:"cred_identifier"`
	Config     sqlxx.JSONRawMessage     `db:"cred_config"`
	Version    int                      `db:"cred_version"`
	CreatedAt  time.Time                `db:"created_at"`
	UpdatedAt  time.Time                `db:"updated_at"`
}

func (queryCredentials) TableName() string {
	return "identity_credentials"
}

type Where struct {
	Condition string
	Args      []interface{}
}

// QueryForCredentials queries for identity credentials with custom WHERE
// clauses, returning the results resolved by the owning identity's UUID.
func QueryForCredentials(con *pop.Connection, where ...Where) (map[uuid.UUID](map[identity.CredentialsType]identity.Credentials), error) {
	q := con.Select(
		"identity_credentials.id cred_id",
		"identity_credentials.identity_id identity_id",
		"identity_credentials.nid nid",
		"ict.name cred_type",
		"ict.id cred_type_id",
		"COALESCE(ici.identifier, '') cred_identifier",
		"identity_credentials.config cred_config",
		"identity_credentials.version cred_version",
		"identity_credentials.created_at created_at",
		"identity_credentials.updated_at updated_at",
	).InnerJoin(
		"identity_credential_types ict",
		"(identity_credentials.identity_credential_type_id = ict.id)",
	).LeftJoin(
		"identity_credential_identifiers ici",
		"(ici.identity_credential_id = identity_credentials.id AND ici.nid = identity_credentials.nid)",
	)
	for _, w := range where {
		q = q.Where("("+w.Condition+")", w.Args...)
	}
	var results []queryCredentials
	if err := q.All(&results); err != nil {
		return nil, sqlcon.HandleError(err)
	}
	credentialsPerIdentity := map[uuid.UUID](map[identity.CredentialsType]identity.Credentials){}
	for _, res := range results {
		credentials, ok := credentialsPerIdentity[res.IdentityID]
		if !ok {
			credentialsPerIdentity[res.IdentityID] = make(map[identity.CredentialsType]identity.Credentials)
			credentials = credentialsPerIdentity[res.IdentityID]
		}
		identifiers := credentials[res.Type].Identifiers
		if res.Identifier != "" {
			identifiers = append(identifiers, res.Identifier)
		}
		if identifiers == nil {
			identifiers = make([]string, 0)
		}
		c := identity.Credentials{
			ID:                       res.ID,
			IdentityID:               res.IdentityID,
			NID:                      res.NID,
			Type:                     res.Type,
			IdentityCredentialTypeID: res.TypeID,
			Identifiers:              identifiers,
			Config:                   res.Config,
			Version:                  res.Version,
			CreatedAt:                res.CreatedAt,
			UpdatedAt:                res.UpdatedAt,
		}
		credentials[res.Type] = c
	}
	// We need deterministic ordering for testing, but sorting in the
	// database can be expensive under certain circumstances.
	for _, creds := range credentialsPerIdentity {
		for k := range creds {
			sort.Strings(creds[k].Identifiers)
		}
	}
	return credentialsPerIdentity, nil
}

func (p *IdentityPersister) ListIdentities(ctx context.Context, params identity.ListIdentityParameters) (res []identity.Identity, err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.ListIdentities")
	defer otelx.End(span, &err)

	span.SetAttributes(
		attribute.Int("page", params.Page),
		attribute.Int("per_page", params.PerPage),
		attribute.StringSlice("expand", params.Expand.ToEager()),
		attribute.Bool("use:credential_identifier_filter", params.CredentialsIdentifier != ""),
		attribute.String("network.id", p.NetworkID(ctx).String()),
	)

	is := make([]identity.Identity, 0)

	con := p.GetConnection(ctx)
	nid := p.NetworkID(ctx)
	query := con.Where("identities.nid = ?", nid).Order("identities.id DESC")

	// Credentials are not expanded through `EagerPreload` but manually after
	// fetching the identities, hence we filter out the relevant expand options.
	var expandExceptCredentials sqlxx.Expandables
	for _, e := range params.Expand {
		if e != identity.ExpandFieldCredentials {
			expandExceptCredentials = append(expandExceptCredentials, e)
		}
	}
	if len(expandExceptCredentials) > 0 {
		query = query.EagerPreload(expandExceptCredentials.ToEager()...)
	}

	if match := params.CredentialsIdentifier; len(match) > 0 {
		// When filtering by credentials identifier, we most likely are looking for a username or email. It is therefore
		// important to normalize the identifier before querying the database.
		match = NormalizeIdentifier(identity.CredentialsTypePassword, match)
		query = query.
			InnerJoin("identity_credentials ic", "ic.identity_id = identities.id").
			InnerJoin("identity_credential_types ict", "ict.id = ic.identity_credential_type_id").
			InnerJoin("identity_credential_identifiers ici", "ici.identity_credential_id = ic.id").
			Where("(ic.nid = ? AND ici.nid = ? AND ici.identifier = ?)", nid, nid, match).
			Where("ict.name IN (?)", identity.CredentialsTypeWebAuthn, identity.CredentialsTypePassword).
			Limit(1)
	} else {
		query = query.Paginate(params.Page, params.PerPage)
	}

	if err := sqlcon.HandleError(query.All(&is)); err != nil {
		return nil, err
	}

	if params.Expand.Has(identity.ExpandFieldCredentials) {
		var ids []interface{}
		for k := range is {
			ids = append(ids, is[k].ID)
		}
		creds, err := QueryForCredentials(con,
			Where{"(identity_credentials.nid = ?)", []interface{}{nid}},
			Where{"(identity_credentials.identity_id IN (?) )", ids},
		)
		if err != nil {
			return nil, err
		}
		for k := range is {
			is[k].Credentials = creds[is[k].ID]
		}
	}

	schemaCache := map[string]string{}
	for k := range is {
		i := &is[k]

		if u, ok := schemaCache[i.SchemaID]; ok {
			i.SchemaURL = u
		} else {
			if err := p.InjectTraitsSchemaURL(ctx, i); err != nil {
				return nil, err
			}
			schemaCache[i.SchemaID] = i.SchemaURL
		}

		if err := i.Validate(); err != nil {
			return nil, err
		}

		if err := identity.UpgradeCredentials(i); err != nil {
			return nil, err
		}

		is[k] = *i
	}

	return is, nil
}

func (p *IdentityPersister) UpdateIdentity(ctx context.Context, i *identity.Identity) (err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.UpdateIdentity")
	defer otelx.End(span, &err)

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

		//#nosec G201 -- TableName is static
		if err := tx.RawQuery(
			fmt.Sprintf(
				`DELETE FROM %s WHERE identity_id = ? AND nid = ?`,
				new(identity.Credentials).TableName(ctx)),
			i.ID, p.NetworkID(ctx)).Exec(); err != nil {

			return sqlcon.HandleError(err)
		}

		if err := update.Generic(WithTransaction(ctx, tx), tx, p.r.Tracer(ctx).Tracer(), i); err != nil {
			return err
		}

		return p.createIdentityCredentials(ctx, i)
	}))
}

func (p *IdentityPersister) DeleteIdentity(ctx context.Context, id uuid.UUID) (err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.DeleteIdentity")
	defer otelx.End(span, &err)

	nid := p.NetworkID(ctx)
	count, err := p.GetConnection(ctx).RawQuery(fmt.Sprintf("DELETE FROM %s WHERE id = ? AND nid = ?", new(identity.Identity).TableName(ctx)),
		id,
		nid,
	).ExecWithCount()
	if err != nil {
		return sqlcon.HandleError(err)
	}
	if count == 0 {
		return errors.WithStack(sqlcon.ErrNoRows)
	}
	return nil
}

func (p *IdentityPersister) GetIdentity(ctx context.Context, id uuid.UUID, expand identity.Expandables) (_ *identity.Identity, err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.GetIdentity")
	defer otelx.End(span, &err)

	span.SetAttributes(
		attribute.String("identity.id", id.String()),
		attribute.StringSlice("expand", expand.ToEager()),
		attribute.String("network.id", p.NetworkID(ctx).String()),
	)

	var i identity.Identity
	if err := p.GetConnection(ctx).Where("id = ? AND nid = ?", id, p.NetworkID(ctx)).First(&i); err != nil {
		return nil, sqlcon.HandleError(err)
	}

	if err := p.HydrateIdentityAssociations(ctx, &i, expand); err != nil {
		return nil, err
	}

	return &i, nil
}

func (p *IdentityPersister) GetIdentityConfidential(ctx context.Context, id uuid.UUID) (res *identity.Identity, err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.GetIdentityConfidential")
	defer otelx.End(span, &err)

	return p.GetIdentity(ctx, id, identity.ExpandEverything)
}

func (p *IdentityPersister) FindVerifiableAddressByValue(ctx context.Context, via identity.VerifiableAddressType, value string) (_ *identity.VerifiableAddress, err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.FindVerifiableAddressByValue")
	otelx.End(span, &err)

	var address identity.VerifiableAddress
	if err := p.GetConnection(ctx).Where("nid = ? AND via = ? AND value = ?", p.NetworkID(ctx), via, stringToLowerTrim(value)).First(&address); err != nil {
		return nil, sqlcon.HandleError(err)
	}

	return &address, nil
}

func (p *IdentityPersister) FindRecoveryAddressByValue(ctx context.Context, via identity.RecoveryAddressType, value string) (_ *identity.RecoveryAddress, err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.FindRecoveryAddressByValue")
	defer otelx.End(span, &err)

	var address identity.RecoveryAddress
	if err := p.GetConnection(ctx).Where("nid = ? AND via = ? AND value = ?", p.NetworkID(ctx), via, stringToLowerTrim(value)).First(&address); err != nil {
		return nil, sqlcon.HandleError(err)
	}

	return &address, nil
}

func (p *IdentityPersister) VerifyAddress(ctx context.Context, code string) (err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.VerifyAddress")
	defer otelx.End(span, &err)

	newCode, err := otp.New()
	if err != nil {
		return err
	}

	count, err := p.GetConnection(ctx).RawQuery(
		//#nosec G201 -- TableName is static
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

func (p *IdentityPersister) UpdateVerifiableAddress(ctx context.Context, address *identity.VerifiableAddress) (err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.UpdateVerifiableAddress")
	defer otelx.End(span, &err)

	address.NID = p.NetworkID(ctx)
	address.Value = stringToLowerTrim(address.Value)
	return update.Generic(ctx, p.GetConnection(ctx), p.r.Tracer(ctx).Tracer(), address)
}

func (p *IdentityPersister) validateIdentity(ctx context.Context, i *identity.Identity) (err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.validateIdentity")
	defer otelx.End(span, &err)

	if err := p.r.IdentityValidator().ValidateWithRunner(ctx, i); err != nil {
		if _, ok := errorsx.Cause(err).(*jsonschema.ValidationError); ok {
			return errors.WithStack(herodot.ErrBadRequest.WithReasonf("%s", err))
		}
		return err
	}

	return nil
}

func (p *IdentityPersister) InjectTraitsSchemaURL(ctx context.Context, i *identity.Identity) (err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.InjectTraitsSchemaURL")
	defer otelx.End(span, &err)

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
