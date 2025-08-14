// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package identity

import (
	"context"
	"database/sql"
	"encoding/base64"
	"fmt"
	"maps"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/sync/errgroup"

	"github.com/ory/herodot"
	"github.com/ory/jsonschema/v3"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/otp"
	"github.com/ory/kratos/persistence/sql/batch"
	"github.com/ory/kratos/persistence/sql/update"
	"github.com/ory/kratos/schema"
	"github.com/ory/kratos/x"
	"github.com/ory/kratos/x/events"
	"github.com/ory/pop/v6"
	"github.com/ory/x/contextx"
	"github.com/ory/x/crdbx"
	"github.com/ory/x/errorsx"
	"github.com/ory/x/otelx"
	"github.com/ory/x/pagination/keysetpagination"
	"github.com/ory/x/pointerx"
	"github.com/ory/x/popx"
	"github.com/ory/x/sqlcon"
	"github.com/ory/x/sqlxx"
)

var (
	_ identity.Pool           = new(IdentityPersister)
	_ identity.PrivilegedPool = new(IdentityPersister)
)

type dependencies interface {
	schema.IdentitySchemaProvider
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
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.ListVerifiableAddresses",
		trace.WithAttributes(
			attribute.Int("per_page", itemsPerPage),
			attribute.Stringer("network.id", p.NetworkID(ctx))))
	defer otelx.End(span, &err)

	if err := p.GetConnection(ctx).Where("nid = ?", p.NetworkID(ctx)).Order("id DESC").Paginate(page, x.MaxItemsPerPage(itemsPerPage)).All(&a); err != nil {
		return nil, sqlcon.HandleError(err)
	}

	return a, nil
}

func (p *IdentityPersister) ListRecoveryAddresses(ctx context.Context, page, itemsPerPage int) (a []identity.RecoveryAddress, err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.ListRecoveryAddresses",
		trace.WithAttributes(
			attribute.Int("per_page", itemsPerPage),
			attribute.Stringer("network.id", p.NetworkID(ctx))))
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
	case identity.CredentialsTypeOIDC, identity.CredentialsTypeSAML:
		// OIDC credentials are case-sensitive
		return match
	case identity.CredentialsTypePassword, identity.CredentialsTypeCodeAuth, identity.CredentialsTypeWebAuthn:
		return stringToLowerTrim(match)
	default:
		return match
	}
}

func (p *IdentityPersister) FindIdentityByCredentialIdentifier(ctx context.Context, identifier string, caseSensitive bool) (_ *identity.Identity, err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.FindIdentityByCredentialIdentifier",
		trace.WithAttributes(
			attribute.Stringer("network.id", p.NetworkID(ctx))))
	defer otelx.End(span, &err)

	var find struct {
		IdentityID uuid.UUID `db:"identity_id"`
	}

	if !caseSensitive {
		identifier = NormalizeIdentifier(identity.CredentialsTypePassword, identifier)
	}

	nid := p.NetworkID(ctx)
	if err := p.GetConnection(ctx).RawQuery(`
SELECT ic.identity_id
FROM identity_credentials ic
INNER JOIN identity_credential_identifiers ici
	ON ic.id = ici.identity_credential_id
WHERE ici.identifier = ?
AND ic.nid = ?
AND ici.nid = ?
LIMIT 1`,
		identifier,
		nid,
		nid,
	).First(&find); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sqlcon.HandleError(err)
		}

		return nil, sqlcon.HandleError(err)
	}
	span.SetAttributes(attribute.Stringer("identity.id", find.IdentityID))

	i, err := p.GetIdentity(ctx, find.IdentityID, identity.ExpandDefault)
	if err != nil {
		return nil, err
	}

	// we don't need the credentials. we just need the identity.
	return i.CopyWithoutCredentials(), nil
}

func (p *IdentityPersister) FindByCredentialsIdentifier(ctx context.Context, ct identity.CredentialsType, match string) (_ *identity.Identity, _ *identity.Credentials, err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.FindByCredentialsIdentifier",
		trace.WithAttributes(
			attribute.Stringer("network.id", p.NetworkID(ctx))))
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

	span.SetAttributes(attribute.String("identity.id", find.IdentityID.String()))

	i, err := p.GetIdentityConfidential(ctx, find.IdentityID)
	if err != nil {
		return nil, nil, err
	}

	creds, ok := i.GetCredentials(ct)
	if !ok {
		return nil, nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("The SQL adapter failed to return the appropriate credentials_type \"%s\". This is a bug in the code.", ct))
	}

	return i, creds, nil
}

func (p *IdentityPersister) FindIdentityByWebauthnUserHandle(ctx context.Context, userHandle []byte) (_ *identity.Identity, err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.FindIdentityByWebauthnUserHandle")
	defer otelx.End(span, &err)

	var id identity.Identity

	var jsonPath string
	con := p.GetConnection(ctx)
	switch con.Dialect.Name() {
	case "sqlite", "mysql":
		jsonPath = "$.user_handle"
	default:
		jsonPath = "user_handle"
	}

	columns := popx.DBColumns[identity.Identity](&popx.AliasQuoter{Alias: "identities", Quoter: con.Dialect})

	if err := con.RawQuery(fmt.Sprintf(`
SELECT %s
FROM identities
INNER JOIN identity_credentials
    ON  identities.id = identity_credentials.identity_id
    AND identities.nid = identity_credentials.nid
    AND identity_credentials.identity_credential_type_id = (
        SELECT id
        FROM identity_credential_types
        WHERE name = ?
     )
WHERE identity_credentials.config ->> '%s' = ? AND identity_credentials.config ->> '%s' IS NOT NULL
  AND identities.nid = ?
LIMIT 1`, columns,
		jsonPath, jsonPath),
		identity.CredentialsTypeWebAuthn,
		base64.StdEncoding.EncodeToString(userHandle),
		p.NetworkID(ctx),
	).First(&id); err != nil {
		return nil, sqlcon.HandleError(err)
	}

	return &id, nil
}

func (p *IdentityPersister) createIdentityCredentials(ctx context.Context, conn *pop.Connection, identities ...*identity.Identity) (err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.createIdentityCredentials",
		trace.WithAttributes(
			attribute.Int("num_identities", len(identities)),
			attribute.Stringer("network.id", p.NetworkID(ctx))))
	defer otelx.End(span, &err)

	var (
		nid         = p.NetworkID(ctx)
		traceConn   = &batch.TracerConnection{Tracer: p.r.Tracer(ctx), Connection: conn}
		credentials []*identity.Credentials
		identifiers []*identity.CredentialIdentifier
	)

	var opts []batch.CreateOpts
	if len(identities) > 1 {
		opts = append(opts, batch.WithPartialInserts)
	}

	for _, ident := range identities {
		for k := range ident.Credentials {
			cred := ident.Credentials[k]

			if len(cred.Config) == 0 {
				cred.Config = sqlxx.JSONRawMessage("{}")
			}

			ct, err := FindIdentityCredentialsTypeByName(conn, cred.Type)
			if err != nil {
				return err
			}

			cred.ID, err = uuid.NewV4()
			if err != nil {
				return err
			}
			cred.IdentityID = ident.ID
			cred.NID = nid
			cred.IdentityCredentialTypeID = ct
			credentials = append(credentials, &cred)

			ident.Credentials[k] = cred
		}
	}
	if err = batch.Create(ctx, traceConn, credentials, opts...); err != nil {
		return err
	}

	for _, cred := range credentials {
		for _, identifier := range cred.Identifiers {
			// Force case-insensitivity and trimming for identifiers
			identifier = NormalizeIdentifier(cred.Type, identifier)

			if identifier == "" {
				return errors.WithStack(herodot.ErrMisconfiguration.WithReasonf(
					"Unable to create identity credentials with missing or empty identifier."))
			}

			ct, err := FindIdentityCredentialsTypeByName(conn, cred.Type)
			if err != nil {
				return err
			}

			identifiers = append(identifiers, &identity.CredentialIdentifier{
				Identifier:                identifier,
				IdentityID:                pointerx.Ptr(cred.IdentityID),
				IdentityCredentialsID:     cred.ID,
				IdentityCredentialsTypeID: ct,
				NID:                       p.NetworkID(ctx),
			})
		}
	}

	if err = batch.Create(ctx, traceConn, identifiers, opts...); err != nil {
		return err
	}

	return nil
}

func (p *IdentityPersister) createVerifiableAddresses(ctx context.Context, conn *pop.Connection, identities ...*identity.Identity) (err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.createVerifiableAddresses",
		trace.WithAttributes(
			attribute.Int("num_identities", len(identities)),
			attribute.Stringer("network.id", p.NetworkID(ctx))))
	defer otelx.End(span, &err)

	work := make([]*identity.VerifiableAddress, 0, len(identities))
	for _, id := range identities {
		for i := range id.VerifiableAddresses {
			work = append(work, &id.VerifiableAddresses[i])
		}
	}
	var opts []batch.CreateOpts
	if len(identities) > 1 {
		opts = append(opts, batch.WithPartialInserts)
	}

	return batch.Create(ctx, &batch.TracerConnection{Tracer: p.r.Tracer(ctx), Connection: conn}, work, opts...)
}

type differ interface {
	Signature() string
	GetID() uuid.UUID
}

func updateAssociationWith[T differ](ctx context.Context, p *IdentityPersister, fromDatabase, updateTo []T,
) (result []T, err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.updateAssociationWith",
		trace.WithAttributes(
			attribute.Stringer("network.id", p.NetworkID(ctx))))
	defer otelx.End(span, &err)

	toKeep, toCreate, toRemoveIDs := diffAssociations(fromDatabase, updateTo)

	// Subtle: we delete the old associations from the DB first, because else
	// they could cause UNIQUE constraints to fail on insert.
	// Foreign key cascade will take care of deleting dependent records.
	if len(toRemoveIDs) > 0 {
		if err := p.GetConnection(ctx).Where("id IN (?)", toRemoveIDs).Where("nid = ?", p.NetworkID(ctx)).Delete(new(T)); err != nil {
			return nil, sqlcon.HandleError(err)
		}
	}

	if len(toCreate) > 0 {
		if err := batch.Create(ctx,
			&batch.TracerConnection{
				Tracer:     p.r.Tracer(ctx),
				Connection: p.GetConnection(ctx),
			},
			toCreate,
		); err != nil {
			return nil, err
		}
	}

	result = make([]T, 0, len(toKeep)+len(toCreate))
	for _, v := range toKeep {
		result = append(result, *v)
	}
	for _, v := range toCreate {
		result = append(result, *v)
	}

	return result, nil
}

func updateAssociation[T differ](ctx context.Context, p *IdentityPersister, i *identity.Identity, inID []T,
) (result []T, err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.updateAssociation",
		trace.WithAttributes(
			attribute.Stringer("identity.id", i.ID),
			attribute.Stringer("network.id", p.NetworkID(ctx))))
	defer otelx.End(span, &err)

	var inDB []T
	if err := p.GetConnection(ctx).
		Where("identity_id = ? AND nid = ?", i.ID, p.NetworkID(ctx)).
		All(&inDB); err != nil {
		return nil, sqlcon.HandleError(err)
	}

	return updateAssociationWith(ctx, p, inDB, inID)
}

func updateCredentialsAssociation(ctx context.Context, p *IdentityPersister, conn *pop.Connection, identityID uuid.UUID, fromDatabase []identity.Credentials, updateTo []identity.Credentials) (result map[identity.CredentialsType]identity.Credentials, err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.updateCredentialsAssociation",
		trace.WithAttributes(
			attribute.Stringer("identity.id", identityID),
			attribute.Stringer("network.id", p.NetworkID(ctx))))
	defer otelx.End(span, &err)

	nid := p.NetworkID(ctx)

	// Normalize new credentials by ensuring IdentityID, NID, and identifiers are set before hashing
	for i := range updateTo {
		updateTo[i].IdentityID = identityID
		updateTo[i].NID = nid
		// Normalize identifiers to match what's stored in the database (make a copy to avoid modifying original)
		normalizedIdentifiers := make([]string, len(updateTo[i].Identifiers))
		for j, identifier := range updateTo[i].Identifiers {
			normalizedIdentifiers[j] = NormalizeIdentifier(updateTo[i].Type, identifier)
		}
		updateTo[i].Identifiers = normalizedIdentifiers
	}

	credsToKeep, newCreds, credsToDeleteIDs := diffAssociations(fromDatabase, updateTo)

	if len(credsToDeleteIDs) > 0 {
		// Delete the credential and its identifiers.
		if err := conn.RawQuery(
			`DELETE FROM identity_credentials WHERE nid = ? AND id IN (?)`,
			nid,
			credsToDeleteIDs,
		).Exec(); err != nil {
			return nil, sqlcon.HandleError(err)
		}
	}

	// Create new credentials that aren't already in the database
	credsToCreate := make(map[identity.CredentialsType]identity.Credentials, len(newCreds))
	for _, c := range newCreds {
		credsToCreate[c.Type] = *c
	}

	if len(credsToCreate) > 0 {
		if err := p.createIdentityCredentials(ctx, conn, &identity.Identity{
			ID:          identityID,
			Credentials: credsToCreate,
		}); err != nil {
			return nil, err
		}
	}

	result = make(map[identity.CredentialsType]identity.Credentials, len(credsToKeep)+len(credsToCreate))
	for _, c := range credsToKeep {
		result[c.Type] = *c
	}
	maps.Copy(result, credsToCreate)

	return result, nil
}

func diffAssociations[T differ](fromDatabase, updateTo []T) (unchanged, toCreate []*T, toRemoveIDs []uuid.UUID) {
	newAssocs := make(map[string]*T, len(updateTo))
	oldAssocs := make(map[string]*T, len(fromDatabase))
	for i, a := range updateTo {
		newAssocs[a.Signature()] = &updateTo[i]
	}
	for i, a := range fromDatabase {
		oldAssocs[a.Signature()] = &fromDatabase[i]
	}

	toRemoveIDs = make([]uuid.UUID, 0, len(fromDatabase))
	toCreate = make([]*T, 0, len(updateTo))
	unchanged = make([]*T, 0, len(updateTo))

	for h, a := range oldAssocs {
		if _, found := newAssocs[h]; found {
			delete(newAssocs, h)
			unchanged = append(unchanged, a)
		} else {
			toRemoveIDs = append(toRemoveIDs, (*a).GetID())
		}
	}

	for _, a := range newAssocs {
		toCreate = append(toCreate, a)
	}

	return
}

func (p *IdentityPersister) normalizeAllAddressess(ctx context.Context, identities ...*identity.Identity) {
	for _, id := range identities {
		p.normalizeRecoveryAddresses(ctx, id)
		p.normalizeVerifiableAddresses(ctx, id)
	}
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
		if !v.Verified {
			v.VerifiedAt = nil
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

func (p *IdentityPersister) createRecoveryAddresses(ctx context.Context, conn *pop.Connection, identities ...*identity.Identity) (err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.createRecoveryAddresses",
		trace.WithAttributes(
			attribute.Int("num_identities", len(identities)),
			attribute.Stringer("network.id", p.NetworkID(ctx))))
	defer otelx.End(span, &err)

	// https://go.dev/play/p/b1kU5Bme2Fr
	work := make([]*identity.RecoveryAddress, 0, len(identities))
	for _, id := range identities {
		for i := range id.RecoveryAddresses {
			work = append(work, &id.RecoveryAddresses[i])
		}
	}

	var opts []batch.CreateOpts
	if len(identities) > 1 {
		opts = append(opts, batch.WithPartialInserts)
	}

	return batch.Create(ctx, &batch.TracerConnection{Tracer: p.r.Tracer(ctx), Connection: conn}, work, opts...)
}

func (p *IdentityPersister) CountIdentities(ctx context.Context) (n int64, err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.CountIdentities",
		trace.WithAttributes(
			attribute.Stringer("network.id", p.NetworkID(ctx))))
	defer otelx.End(span, &err)

	count, err := p.c.WithContext(ctx).Where("nid = ?", p.NetworkID(ctx)).Count(new(identity.Identity))
	if err != nil {
		return 0, sqlcon.HandleError(err)
	}
	span.SetAttributes(attribute.Int("num_identities", count))
	return int64(count), nil
}

func (p *IdentityPersister) CreateIdentity(ctx context.Context, ident *identity.Identity) (err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.CreateIdentity",
		trace.WithAttributes(
			attribute.Stringer("network.id", p.NetworkID(ctx))))
	defer otelx.End(span, &err)

	return p.CreateIdentities(ctx, ident)
}

func (p *IdentityPersister) CreateIdentities(ctx context.Context, identities ...*identity.Identity) (err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.CreateIdentities",
		trace.WithAttributes(
			attribute.Int("identities.count", len(identities)),
			attribute.Stringer("network.id", p.NetworkID(ctx))))
	defer otelx.End(span, &err)

	for _, ident := range identities {
		ident.NID = p.NetworkID(ctx)

		if ident.SchemaID == "" {
			ident.SchemaID = p.r.Config().DefaultIdentityTraitsSchemaID(ctx)
		}

		stateChangedAt := sqlxx.NullTime(time.Now())
		ident.StateChangedAt = &stateChangedAt
		if ident.State == "" {
			ident.State = identity.StateActive
		}

		if len(ident.Traits) == 0 {
			ident.Traits = identity.Traits("{}")
		}

		if err = p.InjectTraitsSchemaURL(ctx, ident); err != nil {
			return err
		}

		if err = p.validateIdentity(ctx, ident); err != nil {
			return err
		}
	}

	var succeededIDs []uuid.UUID
	var partialErr *identity.CreateIdentitiesError
	if err := p.Transaction(ctx, func(ctx context.Context, tx *pop.Connection) error {
		conn := &batch.TracerConnection{
			Tracer:     p.r.Tracer(ctx),
			Connection: tx,
		}

		succeededIDs = make([]uuid.UUID, 0, len(identities))
		failedIdentityIDs := make(map[uuid.UUID]struct{ created bool })
		partialErr = nil
		createdIdentities := make([]*identity.Identity, 0, len(identities))

		var opts []batch.CreateOpts
		if len(identities) > 1 {
			opts = append(opts, batch.WithPartialInserts)
		}
		if err := batch.Create(ctx, conn, identities, opts...); err != nil {
			if partialErr := new(batch.PartialConflictError[identity.Identity]); errors.As(err, &partialErr) {
				for _, k := range partialErr.Failed {
					failedIdentityIDs[k.ID] = struct{ created bool }{false}
				}

				// Mark all created identities that were not in the failed list as created.
				for _, ident := range identities {
					if _, ok := failedIdentityIDs[ident.ID]; !ok {
						createdIdentities = append(createdIdentities, ident)
					}
				}
			} else {
				return sqlcon.HandleError(err)
			}
		} else {
			// If no errors occurred, we can safely assume all identities were created.
			createdIdentities = identities
		}

		p.normalizeAllAddressess(ctx, createdIdentities...)

		if err = p.createVerifiableAddresses(ctx, tx, createdIdentities...); err != nil {
			if partialErr := new(batch.PartialConflictError[identity.VerifiableAddress]); errors.As(err, &partialErr) {
				for _, k := range partialErr.Failed {
					failedIdentityIDs[k.IdentityID] = struct{ created bool }{true}
				}
			} else {
				return sqlcon.HandleError(err)
			}
		}
		if err = p.createRecoveryAddresses(ctx, tx, createdIdentities...); err != nil {
			if partialErr := new(batch.PartialConflictError[identity.RecoveryAddress]); errors.As(err, &partialErr) {
				for _, k := range partialErr.Failed {
					failedIdentityIDs[k.IdentityID] = struct{ created bool }{true}
				}
			} else {
				return sqlcon.HandleError(err)
			}
		}
		if err = p.createIdentityCredentials(ctx, tx, createdIdentities...); err != nil {
			if partialErr := new(batch.PartialConflictError[identity.Credentials]); errors.As(err, &partialErr) {
				for _, k := range partialErr.Failed {
					failedIdentityIDs[k.IdentityID] = struct{ created bool }{true}
				}
			} else if partialErr := new(batch.PartialConflictError[identity.CredentialIdentifier]); errors.As(err, &partialErr) {
				for _, k := range partialErr.Failed {
					credID := k.IdentityCredentialsID
					for _, ident := range identities {
						for _, cred := range ident.Credentials {
							if cred.ID == credID {
								failedIdentityIDs[ident.ID] = struct{ created bool }{true}
							}
						}
					}
				}
			} else {
				return sqlcon.HandleError(err)
			}
		}

		// If any of the batch inserts failed on conflict, let's delete the corresponding
		// identity and return a list of failed identities in the error.
		if len(failedIdentityIDs) > 0 {
			partialErr = identity.NewCreateIdentitiesError(len(failedIdentityIDs))
			idsToBeRemoved := make([]uuid.UUID, 0, len(failedIdentityIDs))

			for _, ident := range identities {
				if info, ok := failedIdentityIDs[ident.ID]; ok {
					partialErr.AddFailedIdentity(ident, sqlcon.ErrUniqueViolation)
					if info.created {
						idsToBeRemoved = append(idsToBeRemoved, ident.ID)
					}
				} else {
					succeededIDs = append(succeededIDs, ident.ID)
				}
			}
			// Manually roll back by deleting the identities that were inserted before the
			// error occurred.
			if err := p.DeleteIdentities(ctx, idsToBeRemoved); err != nil {
				return sqlcon.HandleError(err)
			}

			return nil
		} else {
			// No failures: report all identities as created.
			for _, ident := range identities {
				succeededIDs = append(succeededIDs, ident.ID)
			}
		}

		return nil
	}); err != nil {
		return err
	}

	// Report succeeded identities as created.
	for _, identID := range succeededIDs {
		span.AddEvent(events.NewIdentityCreated(ctx, identID))
	}

	return partialErr.ErrOrNil()
}

func (p *IdentityPersister) HydrateIdentityAssociations(ctx context.Context, i *identity.Identity, expand identity.Expandables) (err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.HydrateIdentityAssociations",
		trace.WithAttributes(
			attribute.Stringer("identity.id", i.ID),
			attribute.Stringer("network.id", p.NetworkID(ctx))))
	defer otelx.End(span, &err)

	nid := p.NetworkID(ctx)

	eg, ctx := errgroup.WithContext(ctx)
	if expand.Has(identity.ExpandFieldRecoveryAddresses) {
		eg.Go(func() error {
			// We use WithContext to get a copy of the connection struct, which solves the race detector
			// from complaining incorrectly.
			//
			// https://github.com/ory/pop/issues/723
			if err := p.GetConnection(ctx).WithContext(ctx).
				Where("identity_id = ? AND nid = ?", i.ID, nid).
				Order("id ASC").
				All(&i.RecoveryAddresses); err != nil {
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
			// https://github.com/ory/pop/issues/723
			if err := p.GetConnection(ctx).WithContext(ctx).
				Order("id ASC").
				Where("identity_id = ? AND nid = ?", i.ID, nid).
				All(&i.VerifiableAddresses); err != nil {
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
			// https://github.com/ory/pop/issues/723
			creds, err := QueryForCredentials(p.GetConnection(ctx).WithContext(ctx),
				Where{"identity_credentials.identity_id = ?", []interface{}{i.ID}},
				Where{"identity_credentials.nid = ?", []interface{}{nid}})
			if err != nil {
				return err
			}
			i.Credentials = creds[i.ID]
			return
		})
	}

	if err := eg.Wait(); err != nil {
		return err
	}

	if err := i.Validate(); err != nil {
		return err
	}

	if err := identity.UpgradeCredentials(i); err != nil {
		return err
	}

	return p.InjectTraitsSchemaURL(ctx, i)
}

type queryCredentials struct {
	Identifier string `db:"cred_identifier"`
	identity.Credentials
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
func QueryForCredentials(con *pop.Connection, where ...Where) (credentialsPerIdentity map[uuid.UUID](map[identity.CredentialsType]identity.Credentials), err error) {
	// This query has been meticulously crafted to be as fast as possible.
	// If you touch it, you will likely introduce a performance regression.
	q := con.Select(
		"COALESCE(identity_credential_identifiers.identifier, '') cred_identifier",
		"identity_credentials.id",
		"identity_credentials.identity_credential_type_id",
		"identity_credentials.identity_id",
		"identity_credentials.nid",
		"identity_credentials.config",
		"identity_credentials.version",
		"identity_credentials.created_at",
		"identity_credentials.updated_at",
	).LeftJoin(identifiersTableNameWithIndexHint(con),
		"identity_credential_identifiers.identity_credential_id = identity_credentials.id AND identity_credential_identifiers.nid = identity_credentials.nid",
	).Order(
		"identity_credential_identifiers.identifier ASC",
	)
	for _, w := range where {
		q = q.Where("("+w.Condition+")", w.Args...)
	}
	var results []queryCredentials
	if err := q.All(&results); err != nil {
		return nil, sqlcon.HandleError(err)
	}

	// assemble
	credentialsPerIdentity = map[uuid.UUID](map[identity.CredentialsType]identity.Credentials){}
	for _, res := range results {

		res.Type, err = FindIdentityCredentialsTypeByID(con, res.IdentityCredentialTypeID)
		if err != nil {
			return nil, err
		}

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
		res.Identifiers = identifiers
		credentials[res.Type] = res.Credentials
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

func identifiersTableNameWithIndexHint(con *pop.Connection) string {
	ici := "identity_credential_identifiers"
	switch con.Dialect.Name() {
	case "cockroach":
		ici += "@identity_credential_identifiers_ici_nid_i_idx"
	case "sqlite3":
		ici += " INDEXED BY identity_credential_identifiers_ici_nid_i_idx"
	case "mysql":
		ici += " USE INDEX(identity_credential_identifiers_ici_nid_i_idx)"
	default:
		// good luck 🤷‍♂️
	}
	return ici
}

func paginationAttributes(params *identity.ListIdentityParameters, paginator *keysetpagination.Paginator) []attribute.KeyValue {
	attrs := []attribute.KeyValue{
		attribute.StringSlice("expand", params.Expand.ToEager()),
		attribute.Bool("use:credential_identifier_filter", params.CredentialsIdentifier != ""),
		attribute.Bool("use:credential_identifier_similar_filter", params.CredentialsIdentifierSimilar != ""),
	}
	if params.PagePagination != nil {
		attrs = append(attrs,
			attribute.Int("page", params.PagePagination.Page),
			attribute.Int("per_page", params.PagePagination.ItemsPerPage))
	} else {
		attrs = append(attrs,
			attribute.String("page_token", paginator.Token().Encode()),
			attribute.Int("page_size", paginator.Size()))
	}
	return attrs
}

// getCredentialTypeIDs returns a map of credential types to their respective IDs.
//
// If a credential type is not found, an error is returned.
func (p *IdentityPersister) getCredentialTypeIDs(ctx context.Context, credentialTypes []identity.CredentialsType) (map[identity.CredentialsType]uuid.UUID, error) {
	result := map[identity.CredentialsType]uuid.UUID{}

	for _, ct := range credentialTypes {
		typeID, err := FindIdentityCredentialsTypeByName(p.GetConnection(ctx), ct)
		if err != nil {
			return nil, err
		}
		result[ct] = typeID
	}

	return result, nil
}

func (p *IdentityPersister) ListIdentities(ctx context.Context, params identity.ListIdentityParameters) (_ []identity.Identity, nextPage *keysetpagination.Paginator, err error) {
	paginator := keysetpagination.GetPaginator(append(
		params.KeySetPagination,
		keysetpagination.WithDefaultToken(identity.DefaultPageToken()),
		keysetpagination.WithDefaultSize(250),
		keysetpagination.WithColumn("id", "ASC"))...)

	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.ListIdentities", trace.WithAttributes(append(
		paginationAttributes(&params, paginator),
		attribute.Stringer("network.id", p.NetworkID(ctx)))...))
	defer otelx.End(span, &err)

	if _, err := uuid.FromString(paginator.Token().Parse("id")["id"]); err != nil {
		return nil, nil, errors.WithStack(x.PageTokenInvalid)
	}

	nid := p.NetworkID(ctx)
	var is []identity.Identity

	if err = p.Transaction(ctx, func(ctx context.Context, con *pop.Connection) error {
		is = make([]identity.Identity, 0) // Make sure we reset this to 0 in case of retries.
		nextPage = nil

		if err := crdbx.SetTransactionReadOnly(con); err != nil {
			return err
		}

		if err := crdbx.SetTransactionConsistency(con, params.ConsistencyLevel, p.r.Config().DefaultConsistencyLevel(ctx)); err != nil {
			return err
		}

		joins := ""
		wheres := "identities.nid = ? AND identities.id > ?"
		args := []any{nid, paginator.Token().Encode()}
		limit := fmt.Sprintf("LIMIT %d", paginator.Size()+1)
		if params.PagePagination != nil {
			wheres = "identities.nid = ?"
			args = []any{nid}
			paginator := pop.NewPaginator(params.PagePagination.Page+1, params.PagePagination.ItemsPerPage)
			limit = fmt.Sprintf("LIMIT %d OFFSET %d", paginator.PerPage, paginator.Offset)
		}
		identifier := params.CredentialsIdentifier
		identifierOperator := "="
		if identifier == "" && params.CredentialsIdentifierSimilar != "" {
			identifier = x.EscapeLikePattern(params.CredentialsIdentifierSimilar) + "%"
			identifierOperator = "LIKE"
		}

		if len(identifier) > 0 {
			types, err := p.getCredentialTypeIDs(ctx, []identity.CredentialsType{
				identity.CredentialsTypeWebAuthn,
				identity.CredentialsTypePassword,
				identity.CredentialsTypeCodeAuth,
				identity.CredentialsTypeOIDC,
				identity.CredentialsTypeSAML,
			})
			if err != nil {
				return err
			}

			// When filtering by credentials identifier, we most likely are looking for a username or email. It is therefore
			// important to normalize the identifier before querying the database.

			joins = params.TransformStatement(`
			INNER JOIN identity_credentials ic ON ic.identity_id = identities.id AND ic.nid = identities.nid
			INNER JOIN identity_credential_identifiers ici ON ici.identity_credential_id = ic.id AND ici.nid = ic.nid
`)

			wheres += fmt.Sprintf(`
			AND ic.nid = ? AND ici.nid = ?
			AND ((ici.identity_credential_type_id IN (?, ?, ?) AND ici.identifier %s ?)
              OR (ici.identity_credential_type_id IN (?, ?) AND ici.identifier %s ?))
			`, identifierOperator, identifierOperator)
			args = append(args,
				nid, nid,
				types[identity.CredentialsTypeWebAuthn], types[identity.CredentialsTypePassword], types[identity.CredentialsTypeCodeAuth],
				NormalizeIdentifier(identity.CredentialsTypePassword, identifier),
				types[identity.CredentialsTypeOIDC], types[identity.CredentialsTypeSAML],
				identifier,
			)
		}

		if len(params.IdsFilter) > 0 {
			wheres += `
				AND identities.id in (?)
			`
			args = append(args, params.IdsFilter)
		} else if !params.OrganizationID.IsNil() {
			wheres += `
				AND identities.organization_id = ?
			`
			args = append(args, params.OrganizationID.String())
		}

		columns := popx.DBColumns[identity.Identity](&popx.AliasQuoter{Alias: "identities", Quoter: con.Dialect})

		query := fmt.Sprintf(`
		SELECT DISTINCT %s
		FROM identities AS identities
		%s
		WHERE
		%s
		ORDER BY identities.id ASC
		%s`,
			columns,
			joins, wheres, limit)

		if err := con.RawQuery(query, args...).All(&is); err != nil {
			return sqlcon.HandleError(err)
		}

		if params.PagePagination == nil {
			is, nextPage = keysetpagination.Result(is, paginator)
		}

		if len(is) == 0 {
			return nil
		}

		identitiesByID := make(map[uuid.UUID]*identity.Identity, len(is))
		identityIDs := make([]any, len(is))
		for k := range is {
			identitiesByID[is[k].ID] = &is[k]
			identityIDs[k] = is[k].ID
		}

		for _, e := range params.Expand {
			switch e {
			case identity.ExpandFieldCredentials:
				creds, err := QueryForCredentials(con,
					Where{"identity_credentials.nid = ?", []interface{}{nid}},
					Where{"identity_credentials.identity_id IN (?)", identityIDs})
				if err != nil {
					return err
				}
				for k := range is {
					is[k].Credentials = creds[is[k].ID]
				}
			case identity.ExpandFieldVerifiableAddresses:
				addrs := make([]identity.VerifiableAddress, 0)
				if err := con.Where("identity_id IN (?)", identityIDs).Where("nid = ?", nid).Order("id").All(&addrs); err != nil {
					return sqlcon.HandleError(err)
				}
				for _, addr := range addrs {
					identitiesByID[addr.IdentityID].VerifiableAddresses = append(identitiesByID[addr.IdentityID].VerifiableAddresses, addr)
				}
			case identity.ExpandFieldRecoveryAddresses:
				addrs := make([]identity.RecoveryAddress, 0)
				if err := con.Where("identity_id IN (?)", identityIDs).Where("nid = ?", nid).Order("id").All(&addrs); err != nil {
					return sqlcon.HandleError(err)
				}
				for _, addr := range addrs {
					identitiesByID[addr.IdentityID].RecoveryAddresses = append(identitiesByID[addr.IdentityID].RecoveryAddresses, addr)
				}
			}
		}

		return nil
	}); err != nil {
		return nil, nil, err
	}

	schemaCache := map[string]string{}
	for k := range is {
		i := &is[k]

		if u, ok := schemaCache[i.SchemaID]; ok {
			i.SchemaURL = u
		} else {
			if err := p.InjectTraitsSchemaURL(ctx, i); err != nil {
				return nil, nil, err
			}
			schemaCache[i.SchemaID] = i.SchemaURL
		}

		if err := i.Validate(); err != nil {
			return nil, nil, err
		}

		if err := identity.UpgradeCredentials(i); err != nil {
			return nil, nil, err
		}

		is[k] = *i
	}

	return is, nextPage, nil
}

func (p *IdentityPersister) UpdateIdentityColumns(ctx context.Context, i *identity.Identity, columns ...string) (err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.UpdateIdentityColumns",
		trace.WithAttributes(
			attribute.Stringer("identity.id", i.ID),
			attribute.Stringer("network.id", p.NetworkID(ctx))))
	defer otelx.End(span, &err)

	if err := p.Transaction(ctx, func(ctx context.Context, tx *pop.Connection) error {
		_, err := tx.Where("id = ? AND nid = ?", i.ID, p.NetworkID(ctx)).UpdateQuery(i, columns...)
		return sqlcon.HandleError(err)
	}); err != nil {
		return err
	}

	span.AddEvent(events.NewIdentityUpdated(ctx, i.ID))
	return nil
}

func (p *IdentityPersister) UpdateIdentity(ctx context.Context, i *identity.Identity, mods ...identity.UpdateIdentityModifier) (err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.UpdateIdentity",
		trace.WithAttributes(
			attribute.Stringer("identity.id", i.ID),
			attribute.Stringer("network.id", p.NetworkID(ctx))))
	defer otelx.End(span, &err)

	if err := p.validateIdentity(ctx, i); err != nil {
		return err
	}

	o := identity.NewUpdateIdentityOptions(mods)

	i.NID = p.NetworkID(ctx)
	i.UpdatedAt = time.Now().UTC().Truncate(time.Microsecond)
	if err := sqlcon.HandleError(p.Transaction(ctx, func(ctx context.Context, tx *pop.Connection) error {
		// This returns "ErrNoRows" if the identity does not exist
		if err := update.Generic(WithTransaction(ctx, tx), tx, p.r.Tracer(ctx).Tracer(), i); err != nil {
			return err
		}

		var identityCreds map[identity.CredentialsType]identity.Credentials
		p.normalizeAllAddressess(ctx, i)
		if o.FromDatabase() != nil {
			if o.FromDatabase().ID != i.ID {
				return errors.New("mismatched identity ID: this is a bug")
			}
			var err error
			i.RecoveryAddresses, err = updateAssociationWith(ctx, p, o.FromDatabase().RecoveryAddresses, i.RecoveryAddresses)
			if err != nil {
				return err
			}
			i.VerifiableAddresses, err = updateAssociationWith(ctx, p, o.FromDatabase().VerifiableAddresses, i.VerifiableAddresses)
			if err != nil {
				return err
			}
			identityCreds = o.FromDatabase().Credentials
		} else {
			i.RecoveryAddresses, err = updateAssociation(ctx, p, i, i.RecoveryAddresses)
			if err != nil {
				return err
			}
			i.VerifiableAddresses, err = updateAssociation(ctx, p, i, i.VerifiableAddresses)
			if err != nil {
				return err
			}

			creds, err := QueryForCredentials(tx,
				Where{"identity_credentials.identity_id = ?", []interface{}{i.ID}},
				Where{"identity_credentials.nid = ?", []interface{}{p.NetworkID(ctx)}})
			if err != nil {
				return err
			}
			if c, found := creds[i.ID]; found {
				identityCreds = c
			}
		}

		oldCredentials := make([]identity.Credentials, 0, len(identityCreds))
		for _, cred := range identityCreds {
			oldCredentials = append(oldCredentials, cred)
		}

		// Convert new credentials map to slice
		newCredentials := make([]identity.Credentials, 0, len(i.Credentials))
		for _, cred := range i.Credentials {
			newCredentials = append(newCredentials, cred)
		}

		i.Credentials, err = updateCredentialsAssociation(ctx, p, tx, i.ID, oldCredentials, newCredentials)
		return err
	})); err != nil {
		return err
	}

	span.AddEvent(events.NewIdentityUpdated(ctx, i.ID))
	return nil
}

func (p *IdentityPersister) DeleteIdentity(ctx context.Context, id uuid.UUID) (err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.DeleteIdentity",
		trace.WithAttributes(
			attribute.Stringer("identity.id", id),
			attribute.Stringer("network.id", p.NetworkID(ctx))))
	defer otelx.End(span, &err)

	tableName := new(identity.Identity).TableName(ctx)
	if p.c.Dialect.Name() == "cockroach" {
		tableName += "@primary"
	}
	nid := p.NetworkID(ctx)
	count, err := p.GetConnection(ctx).RawQuery(fmt.Sprintf("DELETE FROM %s WHERE id = ? AND nid = ?", tableName),
		id,
		nid,
	).ExecWithCount()
	if err != nil {
		return sqlcon.HandleError(err)
	}
	if count == 0 {
		return errors.WithStack(sqlcon.ErrNoRows)
	}
	span.AddEvent(events.NewIdentityDeleted(ctx, id))
	return nil
}

func (p *IdentityPersister) DeleteIdentities(ctx context.Context, ids []uuid.UUID) (err error) {
	// This function is only used internally to cleanup partially created identities,
	// when creating a batch of identities at once and some failed to be fully created.
	// This act should not be observable externally and thus we do not emit an event.

	stringIDs := make([]string, len(ids))
	for k, id := range ids {
		stringIDs[k] = id.String()
	}
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.DeleteIdentites",
		trace.WithAttributes(
			attribute.StringSlice("identity.ids", stringIDs),
			attribute.Stringer("network.id", p.NetworkID(ctx))))
	defer otelx.End(span, &err)

	placeholders := strings.TrimSuffix(strings.Repeat("?, ", len(ids)), ", ")
	args := make([]any, 0, len(ids)+1)
	for _, id := range ids {
		args = append(args, id)
	}
	args = append(args, p.NetworkID(ctx))

	tableName := new(identity.Identity).TableName(ctx)
	if p.c.Dialect.Name() == "cockroach" {
		tableName += "@primary"
	}
	count, err := p.GetConnection(ctx).RawQuery(fmt.Sprintf(
		"DELETE FROM %s WHERE id IN (%s) AND nid = ?",
		tableName,
		placeholders,
	),
		args...,
	).ExecWithCount()
	if err != nil {
		return sqlcon.HandleError(err)
	}
	if count != len(ids) {
		return errors.WithStack(sqlcon.ErrNoRows)
	}
	return nil
}

func (p *IdentityPersister) GetIdentity(ctx context.Context, id uuid.UUID, expand identity.Expandables) (_ *identity.Identity, err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.GetIdentity",
		trace.WithAttributes(
			attribute.Stringer("identity.id", id),
			attribute.Stringer("network.id", p.NetworkID(ctx)),
			attribute.StringSlice("expand", expand.ToEager())))
	defer otelx.End(span, &err)

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

func (p *IdentityPersister) FindIdentityByExternalID(ctx context.Context, externalID string, expand identity.Expandables) (res *identity.Identity, err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.FindIdentityByExternalID",
		trace.WithAttributes(
			attribute.String("identity.external_id", externalID),
			attribute.Stringer("network.id", p.NetworkID(ctx))))
	defer otelx.End(span, &err)

	var i identity.Identity
	if err := p.GetConnection(ctx).Where("external_id = ? AND nid = ?", externalID, p.NetworkID(ctx)).First(&i); err != nil {
		return nil, sqlcon.HandleError(err)
	}

	if err := p.HydrateIdentityAssociations(ctx, &i, identity.ExpandEverything); err != nil {
		return nil, err
	}

	return &i, nil
}

func (p *IdentityPersister) FindVerifiableAddressByValue(ctx context.Context, via string, value string) (_ *identity.VerifiableAddress, err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.FindVerifiableAddressByValue",
		trace.WithAttributes(
			attribute.Stringer("network.id", p.NetworkID(ctx))))
	otelx.End(span, &err)

	var address identity.VerifiableAddress
	if err := p.GetConnection(ctx).Where("nid = ? AND via = ? AND value = ?", p.NetworkID(ctx), via, stringToLowerTrim(value)).First(&address); err != nil {
		return nil, sqlcon.HandleError(err)
	}

	return &address, nil
}

func (p *IdentityPersister) FindRecoveryAddressByValue(ctx context.Context, via identity.RecoveryAddressType, value string) (_ *identity.RecoveryAddress, err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.FindRecoveryAddressByValue",
		trace.WithAttributes(
			attribute.Stringer("network.id", p.NetworkID(ctx))))
	defer otelx.End(span, &err)

	var address identity.RecoveryAddress
	if err := p.GetConnection(ctx).Where("nid = ? AND via = ? AND value = ?", p.NetworkID(ctx), via, stringToLowerTrim(value)).First(&address); err != nil {
		return nil, sqlcon.HandleError(err)
	}

	return &address, nil
}

// Find all recovery addresses for an identity if at least one of its recovery addresses matches the provided value.
func (p *IdentityPersister) FindAllRecoveryAddressesForIdentityByRecoveryAddressValue(ctx context.Context, anyRecoveryAddress string) (_ []identity.RecoveryAddress, err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.FindAllRecoveryAddressesForIdentityByRecoveryAddressValue",
		trace.WithAttributes(
			attribute.Stringer("network.id", p.NetworkID(ctx))))
	defer otelx.End(span, &err)

	var recoveryAddresses []identity.RecoveryAddress

	// SQL explanation:
	// 1. Find a row (`B`) with the value matching `anyRecoveryAddress`.
	//    This row has an identity id (`B.identity_id`).
	// 2. Find all rows (`A`) with this identity id.
	//    Meaning: find all recovery addresses for this identity.
	//    The result set includes the user provided address (`anyRecoveryAddress`).
	//    NOTE: Should we exclude from the result set the login address for more security?
	//
	// This is all done in one query with a self-join.
	// We also bound the results for safety.
	err = p.GetConnection(ctx).RawQuery(
		`
SELECT A.id, A.via, A.value, A.identity_id, A.created_at, A.updated_at, A.nid
FROM identity_recovery_addresses A
JOIN identity_recovery_addresses B
ON A.identity_id = B.identity_id
AND A.nid = B.nid
WHERE B.value = ?
AND A.nid = ?
LIMIT 10
		`,
		stringToLowerTrim(anyRecoveryAddress),
		p.NetworkID(ctx),
	).
		All(&recoveryAddresses)
	if err != nil {
		return nil, sqlcon.HandleError(err)
	}
	return recoveryAddresses, nil
}

func (p *IdentityPersister) VerifyAddress(ctx context.Context, code string) (err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.VerifyAddress",
		trace.WithAttributes(
			attribute.Stringer("network.id", p.NetworkID(ctx))))
	defer otelx.End(span, &err)

	newCode, err := otp.New()
	if err != nil {
		return err
	}

	count, err := p.GetConnection(ctx).RawQuery(
		// #nosec G201 -- TableName is static
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

func (p *IdentityPersister) UpdateVerifiableAddress(ctx context.Context, address *identity.VerifiableAddress, updateColumns ...string) (err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.UpdateVerifiableAddress",
		trace.WithAttributes(
			attribute.Stringer("identity.id", address.IdentityID),
			attribute.Stringer("network.id", p.NetworkID(ctx)),
			attribute.StringSlice("columns", updateColumns)))
	defer otelx.End(span, &err)

	address.NID = p.NetworkID(ctx)
	address.Value = stringToLowerTrim(address.Value)
	return update.Generic(ctx, p.GetConnection(ctx), p.r.Tracer(ctx).Tracer(), address, updateColumns...)
}

func (p *IdentityPersister) validateIdentity(ctx context.Context, i *identity.Identity) (err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.validateIdentity",
		trace.WithAttributes(
			attribute.Stringer("identity.id", i.ID),
			attribute.Stringer("network.id", p.NetworkID(ctx))))
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
	// This trace is more noisy than it's worth in diagnostic power.
	// ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.InjectTraitsSchemaURL")
	// defer otelx.End(span, &err)

	ss, err := p.r.IdentityTraitsSchemas(ctx)
	if err != nil {
		return err
	}
	s, err := ss.GetByID(i.SchemaID)
	if err != nil {
		return errors.WithStack(herodot.ErrMisconfiguration.WithReasonf(
			`The JSON Schema "%s" for this identity's traits could not be found.`, i.SchemaID))
	}
	i.SchemaURL = s.SchemaURL(p.r.Config().SelfPublicURL(ctx)).String()
	return nil
}

var (
	credentialTypesID   = x.NewSyncMap[uuid.UUID, identity.CredentialsType]()
	credentialTypesName = x.NewSyncMap[identity.CredentialsType, uuid.UUID]()
)

func FindIdentityCredentialsTypeByID(con *pop.Connection, id uuid.UUID) (identity.CredentialsType, error) {
	result, found := credentialTypesID.Load(id)
	if !found {
		if err := loadCredentialTypes(con); err != nil {
			return "", err
		}

		result, found = credentialTypesID.Load(id)
	}

	if !found {
		return "", errors.WithStack(herodot.ErrInternalServerError.WithReasonf("The SQL adapter failed to return the appropriate credentials_type for id %q. This is a bug in the code.", id))
	}

	return result, nil
}

func FindIdentityCredentialsTypeByName(con *pop.Connection, ct identity.CredentialsType) (uuid.UUID, error) {
	result, found := credentialTypesName.Load(ct)
	if !found {
		if err := loadCredentialTypes(con); err != nil {
			return uuid.Nil, err
		}

		result, found = credentialTypesName.Load(ct)
	}

	if !found {
		return uuid.Nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("The SQL adapter failed to return the appropriate credentials_type for name %q. This is a bug in the code.", ct))
	}

	return result, nil
}

var mux sync.Mutex

func loadCredentialTypes(con *pop.Connection) (err error) {
	ctx, span := trace.SpanFromContext(con.Context()).TracerProvider().Tracer("").Start(con.Context(), "persistence.sql.identity.loadCredentialTypes")
	defer otelx.End(span, &err)
	_ = ctx

	mux.Lock()
	defer mux.Unlock()
	var tt []identity.CredentialsTypeTable
	if err := con.WithContext(ctx).All(&tt); err != nil {
		return sqlcon.HandleError(err)
	}

	for _, t := range tt {
		credentialTypesID.Store(t.ID, t.Name)
		credentialTypesName.Store(t.Name, t.ID)
	}

	return nil
}
