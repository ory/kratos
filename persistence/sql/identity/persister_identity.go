// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package identity

import (
	"cmp"
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"maps"
	"reflect"
	"slices"
	"sort"
	"strings"
	"sync/atomic"
	"time"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

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
	"github.com/ory/x/logrusx"
	"github.com/ory/x/otelx"
	"github.com/ory/x/pagination/keysetpagination"
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
	logrusx.Provider
	config.Provider
	contextx.Provider
	otelx.Provider
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

// PreferExactMatch returns the element from results whose value (extracted by getValue)
// matches originalValue. If no exact match exists, the first element is returned.
// Used by IN(normalized, original) queries to prefer the non-normalized match.
func PreferExactMatch[T any](results []T, originalValue string, getValue func(T) string) T {
	for _, r := range results {
		if getValue(r) == originalValue {
			return r
		}
	}
	return results[0]
}

// NormalizeIdentifier takes a credential type to determine the type of
// formatting to apply to the 'match' string. Password, Code, and
// WebAuthn types perform graceful normalization which include lower
// casing for email formats.
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

	case identity.CredentialsTypeDeviceAuthn:
		return match

	case identity.CredentialsTypePassword, identity.CredentialsTypeCodeAuth, identity.CredentialsTypeWebAuthn:
		return x.GracefulNormalization(match)
	default:
		return match
	}
}

func (p *IdentityPersister) FindIdentityByCredentialIdentifier(ctx context.Context, identifier string, caseSensitive bool, expand identity.Expandables) (_ *identity.Identity, err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.FindIdentityByCredentialIdentifier",
		trace.WithAttributes(
			attribute.Stringer("network.id", p.NetworkID(ctx))))
	defer otelx.End(span, &err)

	var res []struct {
		IdentityID uuid.UUID `db:"identity_id"`
		Identifier string    `db:"identifier"`
	}

	var normalizedIdentifier string
	if !caseSensitive {
		normalizedIdentifier = NormalizeIdentifier(identity.CredentialsTypePassword, identifier)
	} else {
		normalizedIdentifier = identifier
	}

	nid := p.NetworkID(ctx)

	// Query with both normalized and non-normalized
	// identifiers for backward compatibility.
	// LIMIT is 1 when both forms are identical, 2 when they differ.
	limit := 1
	if normalizedIdentifier != identifier {
		limit = 2
	}
	err = p.GetConnection(ctx).RawQuery(`
SELECT ic.identity_id, ici.identifier
FROM identity_credentials ic
INNER JOIN identity_credential_identifiers ici
ON ic.id = ici.identity_credential_id
WHERE ici.identifier IN (?,?)
AND ic.nid = ?
AND ici.nid = ?
LIMIT ?`,
		normalizedIdentifier,
		identifier,
		nid,
		nid,
		limit,
	).All(&res)
	if err != nil {
		return nil, sqlcon.HandleError(err)
	}

	if len(res) == 0 {
		return nil, sqlcon.HandleError(sqlcon.ErrNoRows())
	}

	result := PreferExactMatch(res, identifier, func(r struct {
		IdentityID uuid.UUID `db:"identity_id"`
		Identifier string    `db:"identifier"`
	},
	) string {
		return r.Identifier
	})
	if len(res) > 1 {
		for _, r := range res {
			if r.IdentityID != result.IdentityID {
				p.r.Logger().Warnf("Possible duplicate identities with IDs [%q, %q]", result.IdentityID, r.IdentityID)
				break
			}
		}
	}

	// Record result count for observability
	span.SetAttributes(
		attribute.String("identity.id", result.IdentityID.String()),
		attribute.Int("identity.results", len(res)),
	)

	return p.GetIdentity(ctx, result.IdentityID, expand)
}

func (p *IdentityPersister) FindByCredentialsIdentifier(ctx context.Context, ct identity.CredentialsType, match string) (_ *identity.Identity, _ *identity.Credentials, err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.FindByCredentialsIdentifier",
		trace.WithAttributes(
			attribute.Stringer("network.id", p.NetworkID(ctx))))
	defer otelx.End(span, &err)

	nid := p.NetworkID(ctx)

	var res []struct {
		IdentityID uuid.UUID `db:"identity_id"`
		Identifier string    `db:"identifier"`
	}

	// Force case-insensitivity and trimming for identifiers
	normalizedMatch := NormalizeIdentifier(ct, match)

	credentialTypeID, err := FindIdentityCredentialsTypeByName(p.GetConnection(ctx), ct)
	if err != nil {
		return nil, nil, err
	}

	// Query with both normalized and non-normalized
	// identifiers for backward compatibility.
	// LIMIT is 1 when both forms are identical, 2 when they differ.
	limit := 1
	if normalizedMatch != match {
		limit = 2
	}
	if err := p.GetConnection(ctx).RawQuery(`
			SELECT
				ic.identity_id, ici.identifier
			FROM identity_credentials ic
					INNER JOIN identity_credential_identifiers ici
					ON ic.id = ici.identity_credential_id AND ici.identity_credential_type_id = ?
			WHERE ici.identifier IN (?, ?)
			AND ic.nid = ?
			AND ici.nid = ?
			LIMIT ?`,
		credentialTypeID,
		normalizedMatch,
		match,
		nid,
		nid,
		limit,
	).All(&res); err != nil {
		return nil, nil, sqlcon.HandleError(err)
	}

	if len(res) == 0 {
		return nil, nil, sqlcon.HandleError(sqlcon.ErrNoRows())
	}

	result := PreferExactMatch(res, match, func(r struct {
		IdentityID uuid.UUID `db:"identity_id"`
		Identifier string    `db:"identifier"`
	},
	) string {
		return r.Identifier
	})
	if len(res) > 1 {
		for _, r := range res {
			if r.IdentityID != result.IdentityID {
				p.r.Logger().Warnf("Possible duplicate identities with IDs [%q, %q]", result.IdentityID, r.IdentityID)
				break
			}
		}
	}

	// Record result count for observability
	span.SetAttributes(
		attribute.String("identity.id", result.IdentityID.String()),
		attribute.Int("identity.results", len(res)),
	)

	i, err := p.GetIdentityConfidential(ctx, result.IdentityID)
	if err != nil {
		return nil, nil, err
	}

	creds, ok := i.GetCredentials(ct)
	if !ok {
		return nil, nil, errors.WithStack(herodot.ErrInternalServerError().WithReasonf("The SQL adapter failed to return the appropriate credentials_type \"%s\". This is a bug in the code.", ct))
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

	credTypeID, err := FindIdentityCredentialsTypeByName(con, identity.CredentialsTypeWebAuthn)
	if err != nil {
		return nil, err
	}

	if err := con.RawQuery(fmt.Sprintf(`
SELECT %s
FROM identities
INNER JOIN identity_credentials
    ON  identities.id = identity_credentials.identity_id
    AND identities.nid = identity_credentials.nid
    AND identity_credentials.identity_credential_type_id = ?
WHERE identity_credentials.config ->> '%s' = ? AND identity_credentials.config ->> '%s' IS NOT NULL
  AND identities.nid = ?
LIMIT 1`, columns,
		jsonPath, jsonPath),
		credTypeID,
		base64.StdEncoding.EncodeToString(userHandle),
		p.NetworkID(ctx),
	).First(&id); err != nil {
		return nil, sqlcon.HandleError(err)
	}

	return &id, nil
}

func (p *IdentityPersister) createIdentityCredentials(ctx context.Context, extraColumns []identity.ExtraColumn, identities ...*identity.Identity) (err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.createIdentityCredentials",
		trace.WithAttributes(
			attribute.Int("num_identities", len(identities)),
			attribute.Stringer("network.id", p.NetworkID(ctx))))
	defer otelx.End(span, &err)

	type credentialSlot struct {
		ident *identity.Identity
		key   identity.CredentialsType
	}

	var (
		nid         = p.NetworkID(ctx)
		traceConn   = &batch.TracerConnection{Tracer: p.r.Tracer(ctx), Connection: p.GetConnection(ctx)}
		credentials []*identity.Credentials
		slots       = map[*identity.Credentials]credentialSlot{}
		identifiers []*identity.CredentialIdentifier
	)

	var opts []batch.CreateOpts
	if len(extraColumns) > 0 {
		// Extra columns are values the caller already knows for columns that are
		// not part of the OSS model (e.g. crdb_region on CockroachDB
		// multi-region). Writing them with the insert avoids the server-side
		// fallback (a column default or a lookup derived from a foreign key),
		// which for region columns costs a cross-region round trip per statement.
		opts = append(opts, batch.WithExtraColumns(extraColumns))
	}
	if len(identities) > 1 {
		opts = append(opts, batch.WithPartialInserts)
	}

	for _, ident := range identities {
		for k := range ident.Credentials {
			cred := new(ident.Credentials[k])

			if len(cred.Config) == 0 {
				cred.Config = sqlxx.JSONRawMessage("{}")
			}

			ct, err := FindIdentityCredentialsTypeByName(p.GetConnection(ctx), cred.Type)
			if err != nil {
				return err
			}

			cred.IdentityID = ident.ID
			cred.NID = nid
			cred.IdentityCredentialTypeID = ct

			// TOTP and lookup-secret AAL2 logins resolve the credential by
			// joining identity_credential_identifiers, using the identity ID
			// as the identifier (see selfservice/strategy/{totp,lookup}/
			// login.go and settings.go). The admin import path does not
			// provide an identifier, and at import time the identity ID may
			// still be the zero value (CockroachDB assigns it via
			// gen_random_uuid() during the identity insert). At this point
			// ident.ID is the persisted identity ID, so default the
			// identifier here to keep AAL2 login working for imported
			// credentials. See https://github.com/ory/kratos/issues/4561.
			if len(cred.Identifiers) == 0 &&
				(cred.Type == identity.CredentialsTypeTOTP ||
					cred.Type == identity.CredentialsTypeLookup) {
				cred.Identifiers = []string{ident.ID.String()}
			}

			credentials = append(credentials, cred)
			slots[cred] = credentialSlot{ident: ident, key: k}
		}
	}

	err = batch.Create(ctx, traceConn, credentials, opts...)

	for _, cred := range credentials {
		slot := slots[cred]
		slot.ident.Credentials[slot.key] = *cred
	}

	if err != nil {
		return err
	}

	for _, cred := range credentials {
		for _, identifier := range cred.Identifiers {
			// Force case-insensitivity and trimming for identifiers
			identifier = NormalizeIdentifier(cred.Type, identifier)

			if identifier == "" {
				return errors.WithStack(herodot.ErrMisconfiguration().WithReasonf(
					"Unable to create identity credentials with missing or empty identifier."))
			}

			ct, err := FindIdentityCredentialsTypeByName(p.GetConnection(ctx), cred.Type)
			if err != nil {
				return err
			}

			identifiers = append(identifiers, &identity.CredentialIdentifier{
				Identifier:                identifier,
				IdentityID:                new(cred.IdentityID),
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

func (p *IdentityPersister) createVerifiableAddresses(ctx context.Context, conn *pop.Connection, extraColumns []identity.ExtraColumn, identities ...*identity.Identity) (err error) {
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
	if len(extraColumns) > 0 {
		// Extra columns are values the caller already knows for columns that are
		// not part of the OSS model (e.g. crdb_region on CockroachDB
		// multi-region). Writing them with the insert avoids the server-side
		// fallback (a column default or a lookup derived from a foreign key),
		// which for region columns costs a cross-region round trip per statement.
		opts = append(opts, batch.WithExtraColumns(extraColumns))
	}
	if len(identities) > 1 {
		opts = append(opts, batch.WithPartialInserts)
	}

	return batch.Create(ctx, &batch.TracerConnection{Tracer: p.r.Tracer(ctx), Connection: conn}, work, opts...)
}

type differ interface {
	Signature() string
	GetID() uuid.UUID
}

func updateAssociationWith[T differ](ctx context.Context, p *IdentityPersister, extraColumns []identity.ExtraColumn, fromDatabase, updateTo []T,
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
		var opts []batch.CreateOpts
		if len(extraColumns) > 0 {
			// Extra columns are values the caller already knows for columns that are
			// not part of the OSS model (e.g. crdb_region on CockroachDB
			// multi-region). Writing them with the insert avoids the server-side
			// fallback (a column default or a lookup derived from a foreign key),
			// which for region columns costs a cross-region round trip per statement.
			opts = append(opts, batch.WithExtraColumns(extraColumns))
		}
		if err := batch.Create(ctx,
			&batch.TracerConnection{
				Tracer:     p.r.Tracer(ctx),
				Connection: p.GetConnection(ctx),
			},
			toCreate,
			opts...,
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

func updateAssociation[T differ](ctx context.Context, p *IdentityPersister, extraColumns []identity.ExtraColumn, i *identity.Identity, inID []T,
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

	return updateAssociationWith(ctx, p, extraColumns, inDB, inID)
}

func (p *IdentityPersister) updateCredentialsAssociation(ctx context.Context, extraColumns []identity.ExtraColumn, identityID uuid.UUID, fromDatabase []identity.Credentials, updateTo []identity.Credentials) (result map[identity.CredentialsType]identity.Credentials, err error) {
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
		conn := p.GetConnection(ctx)
		q := "DELETE FROM identity_credentials WHERE nid = ? AND id IN (?)"
		if conn.Dialect.Name() == "cockroach" {
			q = "DELETE FROM identity_credentials@primary WHERE nid = ? AND id IN (?)"
		}
		if err := conn.RawQuery(q, nid, credsToDeleteIDs).Exec(); err != nil {
			return nil, sqlcon.HandleError(err)
		}
	}

	// Create new credentials that aren't already in the database
	credsToCreate := make(map[identity.CredentialsType]identity.Credentials, len(newCreds))
	for _, c := range newCreds {
		credsToCreate[c.Type] = *c
	}

	if len(credsToCreate) > 0 {
		if err := p.createIdentityCredentials(ctx, extraColumns, &identity.Identity{
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

func (p *IdentityPersister) normalizeAllAddresses(ctx context.Context, identities ...*identity.Identity) {
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
		v.Value = x.GracefulNormalization(v.Value)
		v.Via = cmp.Or(v.Via, identity.AddressTypeEmail)
		if len(v.Status) == 0 {
			if v.Verified {
				v.Status = identity.VerifiableAddressStatusCompleted
			} else {
				v.Status = identity.VerifiableAddressStatusPending
			}
		}

		// If verified is true but no timestamp is set, we default to time.Now
		if v.Verified && (v.VerifiedAt == nil || time.Time(*v.VerifiedAt).IsZero()) {
			v.VerifiedAt = new(sqlxx.NullTime(time.Now()))
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
		id.RecoveryAddresses[k].Value = x.GracefulNormalization(id.RecoveryAddresses[k].Value)
		id.RecoveryAddresses[k].Via = cmp.Or(id.RecoveryAddresses[k].Via, identity.AddressTypeEmail)
	}
}

func (p *IdentityPersister) createRecoveryAddresses(ctx context.Context, conn *pop.Connection, extraColumns []identity.ExtraColumn, identities ...*identity.Identity) (err error) {
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
	if len(extraColumns) > 0 {
		// Extra columns are values the caller already knows for columns that are
		// not part of the OSS model (e.g. crdb_region on CockroachDB
		// multi-region). Writing them with the insert avoids the server-side
		// fallback (a column default or a lookup derived from a foreign key),
		// which for region columns costs a cross-region round trip per statement.
		opts = append(opts, batch.WithExtraColumns(extraColumns))
	}
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

	return p.CreateIdentities(ctx, []*identity.Identity{ident})
}

func (p *IdentityPersister) CreateIdentities(ctx context.Context, identities []*identity.Identity, opts ...identity.CreateIdentitiesModifier) (err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.CreateIdentities",
		trace.WithAttributes(
			attribute.Int("identities.count", len(identities)),
			attribute.Stringer("network.id", p.NetworkID(ctx))))
	defer otelx.End(span, &err)

	options := identity.NewCreateIdentitiesOptions(opts)

	for _, ident := range identities {
		ident.NID = p.NetworkID(ctx)

		if ident.SchemaID == "" {
			ident.SchemaID = p.r.Config().DefaultIdentityTraitsSchemaID(ctx)
		}

		stateChangedAt := sqlxx.NullTime(time.Now().UTC())
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

		var batchOpts []batch.CreateOpts
		if extras := options.ExtraColumns; len(extras) > 0 {
			batchOpts = append(batchOpts, batch.WithExtraColumns(extras))
		}
		if len(identities) > 1 {
			batchOpts = append(batchOpts, batch.WithPartialInserts)
		}
		if err := batch.Create(ctx, conn, identities, batchOpts...); err != nil {
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

		p.normalizeAllAddresses(ctx, createdIdentities...)

		if err = p.createVerifiableAddresses(ctx, tx, options.ExtraColumns, createdIdentities...); err != nil {
			if partialErr := new(batch.PartialConflictError[identity.VerifiableAddress]); errors.As(err, &partialErr) {
				for _, k := range partialErr.Failed {
					failedIdentityIDs[k.IdentityID] = struct{ created bool }{true}
				}
			} else {
				return sqlcon.HandleError(err)
			}
		}
		if err = p.createRecoveryAddresses(ctx, tx, options.ExtraColumns, createdIdentities...); err != nil {
			if partialErr := new(batch.PartialConflictError[identity.RecoveryAddress]); errors.As(err, &partialErr) {
				for _, k := range partialErr.Failed {
					failedIdentityIDs[k.IdentityID] = struct{ created bool }{true}
				}
			} else {
				return sqlcon.HandleError(err)
			}
		}
		if err = p.createIdentityCredentials(ctx, options.ExtraColumns, createdIdentities...); err != nil {
			if partialErr := new(batch.PartialConflictError[identity.Credentials]); errors.As(err, &partialErr) {
				for _, k := range partialErr.Failed {
					failedIdentityIDs[k.IdentityID] = struct{ created bool }{true}
				}
			} else if partialErr := new(batch.PartialConflictError[identity.CredentialIdentifier]); errors.As(err, &partialErr) {
				for _, k := range partialErr.Failed {
					// The failed identifier carries the owning identity ID
					// directly, so map it back without scanning every
					// identity's credentials by ID.
					if k.IdentityID != nil {
						failedIdentityIDs[*k.IdentityID] = struct{ created bool }{true}
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
					partialErr.AddFailedIdentity(ident, sqlcon.ErrUniqueViolation())
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

	if expand.Has(identity.ExpandFieldRecoveryAddresses) {
		if err := p.GetConnection(ctx).
			Where("identity_id = ? AND nid = ?", i.ID, nid).
			Order("id ASC").
			All(&i.RecoveryAddresses); err != nil {
			return sqlcon.HandleError(err)
		}
	}

	if expand.Has(identity.ExpandFieldVerifiableAddresses) {
		if err := p.GetConnection(ctx).
			Order("id ASC").
			Where("identity_id = ? AND nid = ?", i.ID, nid).
			All(&i.VerifiableAddresses); err != nil {
			return sqlcon.HandleError(err)
		}
	}

	if expand.Has(identity.ExpandFieldCredentials) {
		creds, err := QueryForCredentials(p.GetConnection(ctx),
			Where{"identity_credentials.identity_id = ?", []interface{}{i.ID}},
			Where{"identity_credentials.nid = ?", []interface{}{nid}})
		if err != nil {
			return err
		}
		i.Credentials = creds[i.ID]
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
	if (params.ColumnsTransformer == nil) != (params.RowScanner == nil) {
		return nil, nil, errors.New("ListIdentityParameters: ColumnsTransformer and RowScanner must be set together")
	}

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
				identity.CredentialsTypeDeviceAuthn,
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
		if params.ColumnsTransformer != nil {
			columns = params.ColumnsTransformer(columns)
		}

		// DISTINCT is only needed when the credentials-identifier filter adds the
		// INNER JOINs below: a single identity can match multiple identifier rows,
		// which would otherwise produce duplicates. Without the joins, identities.id
		// is the primary key, so every row is already unique and DISTINCT is pure
		// overhead (a distinct processor over the wide traits/metadata columns).
		distinct := ""
		if joins != "" {
			distinct = "DISTINCT "
		}

		query := fmt.Sprintf(`
		SELECT %s%s
		FROM identities AS identities
		%s
		WHERE
		%s
		ORDER BY identities.id ASC
		%s`,
			distinct, columns,
			joins, wheres, limit)

		if params.RowScanner != nil {
			is, err = params.RowScanner(con, query, args)
			if err != nil {
				return sqlcon.HandleError(err)
			}
		} else if err := con.RawQuery(query, args...).All(&is); err != nil {
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

// credentialsConfigLockTimeout bounds how long UpdateCredentialsConfig waits
// for the credential-row lock, so a flood of requests against one row cannot
// park waiters on pooled connections until the pool is exhausted. Enforced
// via context cancellation, which aborts a blocked statement and rolls back
// the transaction. It does not apply to SQLite (see UpdateCredentialsConfig).
const credentialsConfigLockTimeout = 5 * time.Second

// UpdateCredentialsConfig atomically read-modify-writes a single
// identity_credentials row's config under READ COMMITTED, holding an exclusive
// row lock (SELECT ... FOR UPDATE) across the whole read-mutate-write cycle:
// concurrent updates to the same row serialize and mutate always observes the
// latest committed config.
//
// The isolation level is deliberate. Under CockroachDB SERIALIZABLE,
// SELECT ... FOR UPDATE is only best-effort — a waiter may not queue behind the
// holder — whereas under READ COMMITTED it is a durable, replicated lock that
// behaves like the textbook Postgres/MySQL row lock. So here the lock, not the
// isolation level, is the correctness mechanism: it is the sole defense against
// a lost update, and any change that weakens it (reintroducing a join, an index
// hint that changes the plan, dropping FOR UPDATE) is a silent correctness bug
// rather than a throughput regression. This is safe because the operation is a
// single-row read-modify-write with no cross-row invariant, so the other READ
// COMMITTED anomalies (read skew, phantoms, write skew) cannot arise. On a
// cluster without READ COMMITTED the request is upgraded back to SERIALIZABLE,
// which is fail-safe here; on SQLite pop serializes whole transactions on an
// in-process mutex, which subsumes the row lock, and the isolation option is
// ignored.
//
// mutate maps the current config JSON to the new one; a structurally equal
// result skips the write. It must be pure: it may run more than once if the
// database retries the transaction. Calls inside a surrounding transaction
// are rejected — the lock and its timeout must be scoped to the transaction
// opened here.
//
// opts may narrow the row set with ExtraColumns, applied as equality
// predicates to both statements. WithDerivedIdentifiers additionally syncs
// the credential's identifier rows to the set derived from the post-mutation
// config, inside the same transaction and lock; like mutate, derive may run
// more than once on database retries.
func (p *IdentityPersister) UpdateCredentialsConfig(ctx context.Context, identityID uuid.UUID, ct identity.CredentialsType, mutate func(config []byte) ([]byte, error), opts ...identity.UpdateCredentialsConfigModifier) (err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.UpdateCredentialsConfig",
		trace.WithAttributes(
			attribute.Stringer("identity.id", identityID),
			attribute.Stringer("network.id", p.NetworkID(ctx)),
			attribute.String("credentials.type", string(ct))))
	defer otelx.End(span, &err)

	// An ambient transaction would be reused by popx, extending the lock to
	// its lifetime — fail loudly instead.
	if popx.InTransaction(ctx) {
		return errors.WithStack(herodot.ErrInternalServerError().WithReason("UpdateCredentialsConfig must not be called inside a surrounding transaction: its row lock and lock timeout are scoped to the transaction it opens itself"))
	}

	// The lock-wait budget does not apply to SQLite: pop serializes whole
	// SQLite transactions on an in-process per-connection mutex, so the row
	// lock the budget bounds on the cluster databases does not exist — and a
	// mutation queued on that mutex behind unrelated transactions on a loaded
	// machine (parallel test runs) would burn the budget waiting without
	// holding any pooled connection. The caller's context still cancels.
	if p.c.Dialect.Name() != "sqlite3" {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, credentialsConfigLockTimeout)
		defer cancel()
	}

	nid := p.NetworkID(ctx)

	o := identity.NewUpdateCredentialsConfigOptions(opts)
	var extraSQLBuilder strings.Builder
	var extraArgs []any
	for _, col := range o.ExtraColumns {
		_, _ = fmt.Fprintf(&extraSQLBuilder, " AND %s = ?", col.K)
		extraArgs = append(extraArgs, col.V)
	}
	extraSQL := extraSQLBuilder.String()

	// READ COMMITTED so SELECT ... FOR UPDATE takes a real, durable row lock on
	// CockroachDB (see the method doc). popx applies the isolation for
	// Postgres/MySQL/CockroachDB and ignores it for SQLite.
	txOpts := &sql.TxOptions{Isolation: sql.LevelReadCommitted}
	var wrote bool
	if err := popx.TransactionWithOptions(ctx, p.c.WithContext(ctx), txOpts, func(ctx context.Context, tx *pop.Connection) error {
		wrote = false
		// Resolve the type id up front (cached) so the locking SELECT touches
		// only identity_credentials: a join under FOR UPDATE would lock the
		// shared type row and serialize all updates of that type globally.
		typeID, err := FindIdentityCredentialsTypeByName(tx, ct)
		if err != nil {
			return err
		}

		// No index hint: the locking SELECT matches
		// identity_credentials_identity_id_idx. Forcing @primary would make
		// it a full-scan locking read that locks every scanned row.
		selectQuery := `
		SELECT ic.id, ic.config
		FROM identity_credentials ic
		WHERE ic.identity_id = ?
		  AND ic.nid = ?
		  AND ic.identity_credential_type_id = ?` + extraSQL + `
		FOR UPDATE`
		selectArgs := append([]any{identityID, nid, typeID}, extraArgs...)
		if tx.Dialect.Name() == "sqlite3" {
			// SQLite has no FOR UPDATE; pop serializes whole SQLite
			// transactions on an in-process mutex instead.
			selectQuery = `
			SELECT ic.id, ic.config
			FROM identity_credentials ic
			WHERE ic.identity_id = ?
			  AND ic.nid = ?
			  AND ic.identity_credential_type_id = ?` + extraSQL
		}

		var row struct {
			ID     uuid.UUID            `db:"id"`
			Config sqlxx.JSONRawMessage `db:"config"`
		}
		if err := tx.RawQuery(selectQuery, selectArgs...).First(&row); err != nil {
			return sqlcon.HandleError(err)
		}

		newConfig, err := mutate(row.Config)
		if err != nil {
			return err
		}

		// A no-op mutation needs no write; the lock still linearizes it with
		// concurrent writers.
		if !jsonContentEqual(row.Config, newConfig) {
			updateQuery := `
			UPDATE identity_credentials
			SET config = ?
			WHERE id = ? AND nid = ?` + extraSQL

			updateArgs := append([]any{sqlxx.JSONRawMessage(newConfig), row.ID, nid}, extraArgs...)
			if err := tx.RawQuery(updateQuery, updateArgs...).Exec(); err != nil {
				return sqlcon.HandleError(err)
			}
			wrote = true
		}

		// The identifier rows are derived state of the config; sync them under
		// the same lock even when the config write was skipped as a no-op, so
		// a previously diverged set converges.
		if o.DeriveIdentifiers != nil {
			derived, err := o.DeriveIdentifiers(newConfig)
			if err != nil {
				return err
			}
			conn := &batch.TracerConnection{Tracer: p.r.Tracer(ctx), Connection: tx}
			proto := identity.CredentialIdentifier{
				IdentityID:                new(identityID),
				IdentityCredentialsID:     row.ID,
				IdentityCredentialsTypeID: typeID,
				NID:                       nid,
			}
			changed, err := syncDerivedIdentifiers(ctx, conn, ct, proto, derived, o.ExtraColumns)
			if err != nil {
				return err
			}
			// An identifier-only change still updates the identity.
			wrote = wrote || changed
		}
		return nil
	}); err != nil {
		return err
	}

	// Only a real write is an identity update; the no-op path changed nothing.
	if wrote {
		span.AddEvent(events.NewIdentityUpdated(ctx, identityID))
	}
	return nil
}

// jsonContentEqual reports whether a and b encode structurally equal JSON
// documents. Byte equality is not enough: databases normalize stored JSON
// (jsonb key order on PostgreSQL/CockroachDB, binary JSON on MySQL), so a
// re-marshaled but semantically unchanged config rarely matches byte-for-byte.
func jsonContentEqual(a, b []byte) bool {
	var av, bv any
	if json.Unmarshal(a, &av) != nil || json.Unmarshal(b, &bv) != nil {
		return false
	}
	return reflect.DeepEqual(av, bv)
}

// syncDerivedIdentifiers reconciles a credential row's identifier rows with
// the derived set, inside the caller's locked transaction, and reports
// whether any row changed. proto is the template for inserted rows: the
// credential/type/identity/network references are already set, and each
// insert copies it and fills in ID and Identifier. Identifiers are
// normalized like createIdentityCredentials does; an unchanged set writes
// nothing, a changed set is replaced wholesale. extra narrows the read and
// the delete like the caller's config statements and is written explicitly
// on the inserts (the cloud multi-region persister pins crdb_region; the
// explicit write spares the infer_rbr_region_col_using_constraint lookup,
// mirroring UpdateIdentity).
func syncDerivedIdentifiers(ctx context.Context, conn *batch.TracerConnection, ct identity.CredentialsType, proto identity.CredentialIdentifier, derived []string, extra []identity.ExtraColumn) (changed bool, err error) {
	tx := conn.Connection

	var extraSQLBuilder strings.Builder
	extraArgs := make([]any, 0, len(extra))
	for _, col := range extra {
		_, _ = fmt.Fprintf(&extraSQLBuilder, " AND %s = ?", col.K)
		extraArgs = append(extraArgs, col.V)
	}
	extraSQL := extraSQLBuilder.String()

	target := make([]string, 0, len(derived))
	for _, identifier := range derived {
		identifier = NormalizeIdentifier(ct, identifier)
		if identifier == "" {
			return false, errors.WithStack(herodot.ErrMisconfiguration().WithReason("Unable to sync identity credential identifiers with missing or empty identifier."))
		}
		target = append(target, identifier)
	}
	slices.Sort(target)
	target = slices.Compact(target)

	var rows []struct {
		Identifier string `db:"identifier"`
	}
	if err := tx.RawQuery(
		`SELECT identifier FROM identity_credential_identifiers WHERE identity_credential_id = ? AND nid = ?`+extraSQL,
		append([]any{proto.IdentityCredentialsID, proto.NID}, extraArgs...)...).All(&rows); err != nil {
		return false, sqlcon.HandleError(err)
	}
	current := make([]string, len(rows))
	for i, r := range rows {
		current[i] = r.Identifier
	}
	slices.Sort(current)

	if slices.Equal(current, target) {
		return false, nil
	}

	// Replace wholesale, deleting first so values kept across the change do
	// not trip the (nid, type, identifier) unique index on insert.
	if err := tx.RawQuery(
		`DELETE FROM identity_credential_identifiers WHERE identity_credential_id = ? AND nid = ?`+extraSQL,
		append([]any{proto.IdentityCredentialsID, proto.NID}, extraArgs...)...).Exec(); err != nil {
		return false, sqlcon.HandleError(err)
	}
	identifiers := make([]*identity.CredentialIdentifier, len(target))
	for i, identifier := range target {
		ci := proto
		ci.ID = x.NewUUID()
		ci.Identifier = identifier
		identifiers[i] = &ci
	}
	// One batched INSERT keeps the lock hold time flat in the identifier
	// count. A duplicate identifier owned by another identity surfaces here
	// as sqlcon.ErrUniqueViolation (batch.Create normalizes driver errors
	// via sqlcon.HandleError) and rolls back the whole transaction.
	if err := batch.Create(ctx, conn, identifiers, batch.WithExtraColumns(extra)); err != nil {
		return false, err
	}
	return true, nil
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

	span.SetAttributes(attribute.Bool("update.minimize_diff", o.FromDatabase() != nil))

	i.NID = p.NetworkID(ctx)
	i.UpdatedAt = time.Now().UTC().Truncate(time.Microsecond)
	if err := sqlcon.HandleError(p.Transaction(ctx, func(ctx context.Context, tx *pop.Connection) error {
		// This returns "ErrNoRows" if the identity does not exist
		if err := update.Generic(WithTransaction(ctx, tx), tx, p.r.Tracer(ctx).Tracer(), i); err != nil {
			return err
		}

		var identityCreds map[identity.CredentialsType]identity.Credentials
		p.normalizeAllAddresses(ctx, i)
		if o.FromDatabase() != nil {
			if o.FromDatabase().ID != i.ID {
				return errors.New("mismatched identity ID: this is a bug")
			}
			var err error
			i.RecoveryAddresses, err = updateAssociationWith(ctx, p, o.ExtraColumns(), o.FromDatabase().RecoveryAddresses, i.RecoveryAddresses)
			if err != nil {
				return err
			}
			i.VerifiableAddresses, err = updateAssociationWith(ctx, p, o.ExtraColumns(), o.FromDatabase().VerifiableAddresses, i.VerifiableAddresses)
			if err != nil {
				return err
			}
			identityCreds = o.FromDatabase().Credentials
		} else {
			i.RecoveryAddresses, err = updateAssociation(ctx, p, o.ExtraColumns(), i, i.RecoveryAddresses)
			if err != nil {
				return err
			}
			i.VerifiableAddresses, err = updateAssociation(ctx, p, o.ExtraColumns(), i, i.VerifiableAddresses)
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

			if len(identityCreds) > 0 {
				// Create temporary identity to run migrations
				tempIdentity := &identity.Identity{
					ID:          i.ID,
					Credentials: identityCreds,
				}
				if err := identity.UpgradeCredentials(tempIdentity); err != nil {
					return err
				}
				identityCreds = tempIdentity.Credentials
			}
		}

		// Excluded credential types are invisible to the diff: their rows are
		// neither deleted, recreated, nor updated. Their database entries are
		// kept aside and merged back into the returned identity below.
		excludedTypes := o.ExcludedCredentialTypes()
		excludedCreds := make(map[identity.CredentialsType]identity.Credentials, len(excludedTypes))

		oldCredentials := make([]identity.Credentials, 0, len(identityCreds))
		for ct, cred := range identityCreds {
			if slices.Contains(excludedTypes, ct) {
				excludedCreds[ct] = cred
				continue
			}
			oldCredentials = append(oldCredentials, cred)
		}

		// Upgrade incoming credentials to ensure proper comparison with upgraded DB credentials
		if err := identity.UpgradeCredentials(i); err != nil {
			return err
		}

		// Convert new credentials map to slice
		newCredentials := make([]identity.Credentials, 0, len(i.Credentials))
		for ct, cred := range i.Credentials {
			if slices.Contains(excludedTypes, ct) {
				continue
			}
			newCredentials = append(newCredentials, cred)
		}

		updatedCreds, err := p.updateCredentialsAssociation(ctx, o.ExtraColumns(), i.ID, oldCredentials, newCredentials)
		if err != nil {
			return err
		}
		// The excluded types' rows were left untouched; surface their database
		// state on the returned identity instead of the in-memory copy.
		maps.Copy(updatedCreds, excludedCreds)
		i.Credentials = updatedCreds
		return nil
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
		return errors.WithStack(sqlcon.ErrNoRows())
	}
	span.AddEvent(events.NewIdentityDeleted(ctx, id))
	return nil
}

func (p *IdentityPersister) DeleteIdentities(ctx context.Context, ids []uuid.UUID) (err error) {
	// This function is only used internally to cleanup partially created identities,
	// when creating a batch of identities at once and some failed to be fully created.
	// This act should not be observable externally and thus we do not emit an event.

	if len(ids) == 0 {
		return nil
	}

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
		return errors.WithStack(sqlcon.ErrNoRows())
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

	if err := p.HydrateIdentityAssociations(ctx, &i, expand); err != nil {
		return nil, err
	}

	return &i, nil
}

func (p *IdentityPersister) FindVerifiableAddressByValue(ctx context.Context, via string, value string) (_ *identity.VerifiableAddress, err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.FindVerifiableAddressByValue",
		trace.WithAttributes(
			attribute.Stringer("network.id", p.NetworkID(ctx))))
	defer otelx.End(span, &err)

	normalized := x.GracefulNormalization(value)
	limit := 1
	if normalized != value {
		limit = 2
	}
	var addresses []identity.VerifiableAddress
	if err := p.GetConnection(ctx).RawQuery(
		"SELECT * FROM identity_verifiable_addresses WHERE nid = ? AND via = ? AND value IN (?,?) LIMIT ?",
		p.NetworkID(ctx), via, value, normalized, limit,
	).All(&addresses); err != nil {
		return nil, sqlcon.HandleError(err)
	}

	if len(addresses) == 0 {
		return nil, sqlcon.HandleError(sqlcon.ErrNoRows())
	}

	addr := PreferExactMatch(addresses, value, func(a identity.VerifiableAddress) string { return a.Value })
	return &addr, nil
}

func (p *IdentityPersister) FindRecoveryAddressByValue(ctx context.Context, via, value string) (_ *identity.RecoveryAddress, err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.FindRecoveryAddressByValue",
		trace.WithAttributes(
			attribute.Stringer("network.id", p.NetworkID(ctx))))
	defer otelx.End(span, &err)

	normalized := x.GracefulNormalization(value)
	limit := 1
	if normalized != value {
		limit = 2
	}
	var addresses []identity.RecoveryAddress
	if err := p.GetConnection(ctx).RawQuery(
		"SELECT * FROM identity_recovery_addresses WHERE nid = ? AND via = ? AND value IN (?,?) LIMIT ?",
		p.NetworkID(ctx), via, value, normalized, limit,
	).All(&addresses); err != nil {
		return nil, sqlcon.HandleError(err)
	}

	if len(addresses) == 0 {
		return nil, sqlcon.HandleError(sqlcon.ErrNoRows())
	}

	addr := PreferExactMatch(addresses, value, func(a identity.RecoveryAddress) string { return a.Value })
	return &addr, nil
}

// FindAllRecoveryAddressValuesForIdentityByRecoveryAddressValue returns the
// values of all recovery addresses for an identity if at least one of those
// addresses matches the provided value.
func (p *IdentityPersister) FindAllRecoveryAddressValuesForIdentityByRecoveryAddressValue(ctx context.Context, anyRecoveryAddress string) (recoveryAddresses []string, err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.FindAllRecoveryAddressValuesForIdentityByRecoveryAddressValue",
		trace.WithAttributes(
			attribute.Stringer("network.id", p.NetworkID(ctx))))
	defer otelx.End(span, &err)

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
	via := identity.AddressTypeSMS
	if strings.ContainsRune(anyRecoveryAddress, '@') {
		via = identity.AddressTypeEmail
	}

	nid := p.NetworkID(ctx)
	err = p.GetConnection(ctx).RawQuery(`
SELECT A.value
FROM identity_recovery_addresses A
JOIN identity_recovery_addresses B
  ON A.identity_id = B.identity_id
  AND A.nid = B.nid
WHERE A.nid = ?
  AND B.via = ?
  AND B.value IN (?, ?)
ORDER BY A.value
LIMIT 10
		`,
		nid,
		via,
		x.GracefulNormalization(anyRecoveryAddress),
		anyRecoveryAddress,
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
			new(identity.VerifiableAddress).TableName(),
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
		return sqlcon.HandleError(sqlcon.ErrNoRows())
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
	address.Value = x.GracefulNormalization(address.Value)
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
			return errors.WithStack(herodot.ErrBadRequest().WithReasonf("%s", err))
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
		return errors.WithStack(herodot.ErrMisconfiguration().WithReasonf(
			`The JSON Schema "%s" for this identity's traits could not be found.`, i.SchemaID))
	}
	i.SchemaURL = s.SchemaURL(p.r.Config().SelfPublicURL(ctx)).String()
	return nil
}

type credentialTypesCache struct {
	byID   map[uuid.UUID]identity.CredentialsType
	byName map[identity.CredentialsType]uuid.UUID
}

// credentialTypesCachePtr holds the loaded cache. It is nil until the first
// successful load. Reads and writes are lock-free (atomic.Pointer).
// In the rare case where multiple goroutines see a nil pointer simultaneously,
// they each load from the database independently and store the result; the
// last store wins, but all results are identical. After the first successful
// store, all subsequent calls take the fast path.
var credentialTypesCachePtr atomic.Pointer[credentialTypesCache]

// ensureCredentialTypesLoaded loads the credential type mappings from the
// database on the first successful call. If the load fails the pointer stays
// nil and the next call retries, so a transient DB error does not permanently
// break the process.
func ensureCredentialTypesLoaded(con *pop.Connection) (*credentialTypesCache, error) {
	if c := credentialTypesCachePtr.Load(); c != nil {
		return c, nil
	}
	c, err := loadCredentialTypes(con)
	if err != nil {
		return nil, err
	}
	credentialTypesCachePtr.Store(c)
	return c, nil
}

func FindIdentityCredentialsTypeByID(con *pop.Connection, id uuid.UUID) (identity.CredentialsType, error) {
	c, err := ensureCredentialTypesLoaded(con)
	if err != nil {
		return "", err
	}

	result, found := c.byID[id]
	if !found {
		return "", errors.WithStack(herodot.ErrInternalServerError().WithReasonf(`No identity credential type id "%s" could be found, this is a code bug.`, id))
	}

	return result, nil
}

func FindIdentityCredentialsTypeByName(con *pop.Connection, ct identity.CredentialsType) (uuid.UUID, error) {
	c, err := ensureCredentialTypesLoaded(con)
	if err != nil {
		return uuid.Nil, err
	}

	result, found := c.byName[ct]
	if !found {
		return uuid.Nil, errors.WithStack(herodot.ErrInternalServerError().WithReasonf(`No identity credential type "%s" could be found, this is a code bug.`, ct))
	}

	return result, nil
}

func loadCredentialTypes(con *pop.Connection) (_ *credentialTypesCache, err error) {
	ctx, span := trace.SpanFromContext(con.Context()).TracerProvider().Tracer("").Start(con.Context(), "persistence.sql.identity.loadCredentialTypes")
	defer otelx.End(span, &err)
	_ = ctx

	var tt []identity.CredentialsTypeTable
	if err := con.WithContext(ctx).All(&tt); err != nil {
		return nil, sqlcon.HandleError(err)
	}

	if len(tt) == 0 {
		return nil, errors.WithStack(herodot.ErrInternalServerError().WithReasonf("The SQL adapter failed to return any credentials_type. This is a bug in the code/db."))
	}

	c := &credentialTypesCache{
		byID:   make(map[uuid.UUID]identity.CredentialsType, 16),
		byName: make(map[identity.CredentialsType]uuid.UUID, 16),
	}
	for _, t := range tt {
		c.byID[t.ID] = t.Name
		c.byName[t.Name] = t.ID
	}

	return c, nil
}
