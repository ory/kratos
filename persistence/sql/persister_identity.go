package sql

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/ory/x/stringslice"

	"github.com/ory/kratos/corp"

	"github.com/ory/jsonschema/v3"
	"github.com/ory/x/sqlxx"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/otp"
	"github.com/ory/kratos/x"

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

func (p *Persister) ListVerifiableAddresses(ctx context.Context, page, itemsPerPage int) (a []identity.VerifiableAddress, err error) {
	if err := p.GetConnection(ctx).Where("nid = ?", corp.ContextualizeNID(ctx, p.nid)).Order("id DESC").Paginate(page, x.MaxItemsPerPage(itemsPerPage)).All(&a); err != nil {
		return nil, sqlcon.HandleError(err)
	}

	return a, err
}

func (p *Persister) ListRecoveryAddresses(ctx context.Context, page, itemsPerPage int) (a []identity.RecoveryAddress, err error) {
	if err := p.GetConnection(ctx).Where("nid = ?", corp.ContextualizeNID(ctx, p.nid)).Order("id DESC").Paginate(page, x.MaxItemsPerPage(itemsPerPage)).All(&a); err != nil {
		return nil, sqlcon.HandleError(err)
	}

	return a, err
}

func (p *Persister) FindByCredentialsIdentifier(ctx context.Context, ct identity.CredentialsType, match string) (*identity.Identity, *identity.Credentials, error) {
	nid := corp.ContextualizeNID(ctx, p.nid)

	var cts []identity.CredentialsTypeTable
	if err := p.GetConnection(ctx).All(&cts); err != nil {
		return nil, nil, sqlcon.HandleError(err)
	}

	var find struct {
		IdentityID uuid.UUID `db:"identity_id"`
	}

	// Force case-insensitivity for identifiers
	if ct == identity.CredentialsTypePassword {
		match = strings.ToLower(match)
	}

	// #nosec G201
	if err := p.GetConnection(ctx).RawQuery(fmt.Sprintf(`SELECT
    ic.identity_id
FROM %s ic
         INNER JOIN %s ict on ic.identity_credential_type_id = ict.id
         INNER JOIN %s ici on ic.id = ici.identity_credential_id
WHERE ici.identifier = ?
  AND ic.nid = ?
  AND ici.nid = ?
  AND ict.name = ?`,
		corp.ContextualizeTableName(ctx, "identity_credentials"),
		corp.ContextualizeTableName(ctx, "identity_credential_types"),
		corp.ContextualizeTableName(ctx, "identity_credential_identifiers"),
	),
		match,
		nid,
		nid,
		ct,
	).First(&find); err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
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
	var m identity.CredentialsTypeTable
	if err := p.GetConnection(ctx).Where("name = ?", ct).First(&m); err != nil {
		return nil, sqlcon.HandleError(err)
	}
	return &m, nil
}

func (p *Persister) createIdentityCredentials(ctx context.Context, i *identity.Identity) error {
	c := p.GetConnection(ctx)

	nid := corp.ContextualizeNID(ctx, p.nid)
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
		cred.CredentialTypeID = ct.ID
		if err := c.Create(&cred); err != nil {
			return sqlcon.HandleError(err)
		}

		for _, ids := range cred.Identifiers {
			// Force case-insensitivity for identifiers
			if cred.Type == identity.CredentialsTypePassword {
				ids = strings.ToLower(ids)
			}

			if len(ids) == 0 {
				return errors.WithStack(herodot.ErrInternalServerError.WithReasonf("Unable to create identity credentials with missing or empty identifier."))
			}

			if err := c.Create(&identity.CredentialIdentifier{
				Identifier:                ids,
				IdentityCredentialsID:     cred.ID,
				IdentityCredentialsTypeID: ct.ID,
				NID:                       corp.ContextualizeNID(ctx, p.nid),
			}); err != nil {
				return sqlcon.HandleError(err)
			}
		}

		i.Credentials[k] = cred
	}

	return nil
}

func (p *Persister) createVerifiableAddresses(ctx context.Context, i *identity.Identity) error {
	for k := range i.VerifiableAddresses {
		i.VerifiableAddresses[k].IdentityID = i.ID
		i.VerifiableAddresses[k].NID = corp.ContextualizeNID(ctx, p.nid)
		if err := p.GetConnection(ctx).Create(&i.VerifiableAddresses[k]); err != nil {
			return err
		}
	}
	return nil
}

func (p *Persister) createRecoveryAddresses(ctx context.Context, i *identity.Identity) error {
	for k := range i.RecoveryAddresses {
		i.RecoveryAddresses[k].IdentityID = i.ID
		i.RecoveryAddresses[k].NID = corp.ContextualizeNID(ctx, p.nid)
		if err := p.GetConnection(ctx).Create(&i.RecoveryAddresses[k]); err != nil {
			return err
		}
	}
	return nil
}

func (p *Persister) findVerifiableAddresses(ctx context.Context, i *identity.Identity) error {
	var addresses []identity.VerifiableAddress
	if err := p.GetConnection(ctx).Where("identity_id = ? AND nid = ?", i.ID, corp.ContextualizeNID(ctx, p.nid)).Order("id ASC").All(&addresses); err != nil {
		return err
	}
	i.VerifiableAddresses = addresses
	return nil
}

func (p *Persister) findRecoveryAddresses(ctx context.Context, i *identity.Identity) error {
	var addresses []identity.RecoveryAddress
	if err := p.GetConnection(ctx).Where("identity_id = ? AND nid = ?", i.ID, corp.ContextualizeNID(ctx, p.nid)).Order("id ASC").All(&addresses); err != nil {
		return err
	}
	i.RecoveryAddresses = addresses
	return nil
}

func (p *Persister) CountIdentities(ctx context.Context) (int64, error) {
	count, err := p.c.WithContext(ctx).Where("nid = ?", corp.ContextualizeNID(ctx, p.nid)).Count(new(identity.Identity))
	if err != nil {
		return 0, sqlcon.HandleError(err)
	}
	return int64(count), nil
}

func (p *Persister) CreateIdentity(ctx context.Context, i *identity.Identity) error {
	i.NID = corp.ContextualizeNID(ctx, p.nid)

	if i.SchemaID == "" {
		i.SchemaID = config.DefaultIdentityTraitsSchemaID
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

		if err := p.createVerifiableAddresses(ctx, i); err != nil {
			return sqlcon.HandleError(err)
		}

		if err := p.createRecoveryAddresses(ctx, i); err != nil {
			return sqlcon.HandleError(err)
		}

		return p.createIdentityCredentials(ctx, i)
	})
}

func (p *Persister) ListIdentities(ctx context.Context, page, perPage int) ([]identity.Identity, error) {
	is := make([]identity.Identity, 0)

	/* #nosec G201 TableName is static */
	if err := sqlcon.HandleError(p.GetConnection(ctx).Where("nid = ?", corp.ContextualizeNID(ctx, p.nid)).
		EagerPreload("VerifiableAddresses", "RecoveryAddresses").
		Paginate(page, perPage).Order("id DESC").
		All(&is)); err != nil {
		return nil, err
	}

	schemaCache := map[string]string{}

	for k := range is {
		i := &is[k]
		if err := i.ValidateNID(); err != nil {
			return nil, sqlcon.HandleError(err)
		}

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

func (p *Persister) ListIdentitiesFiltered(ctx context.Context, values url.Values, page, perPage int) ([]identity.Identity, error) {
	is := make([]identity.Identity, 0, perPage)

//`[a-zA-Z0-9\.]+`
	/* #nosec G201 TableName is static */
	if err := sqlcon.HandleError(p.GetConnection(ctx).Where("identities.nid = ?", corp.ContextualizeNID(ctx, p.nid)).
		LeftJoin("identity_verifiable_addresses verifiable_addresses", "verifiable_addresses.identity_id=identities.id").
		LeftJoin("identity_recovery_addresses recovery_addresses", "recovery_addresses.identity_id=identities.id").
		EagerPreload("VerifiableAddresses", "RecoveryAddresses").
		Scope(p.buildScope(ctx, values)).
		Paginate(page, perPage).Order("identities.id DESC").
		All(&is)); err != nil {
		return nil, err
	}

	schemaCache := map[string]string{}

	for k := range is {
		i := &is[k]
		if err := i.ValidateNID(); err != nil {
			return nil, sqlcon.HandleError(err)
		}

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
	if err := p.validateIdentity(ctx, i); err != nil {
		return err
	}

	i.NID = corp.ContextualizeNID(ctx, p.nid)
	return sqlcon.HandleError(p.Transaction(ctx, func(ctx context.Context, tx *pop.Connection) error {
		if count, err := tx.Where("id = ? AND nid = ?", i.ID, corp.ContextualizeNID(ctx, p.nid)).Count(i); err != nil {
			return err
		} else if count == 0 {
			return sql.ErrNoRows
		}

		for _, tn := range []string{
			new(identity.Credentials).TableName(ctx),
			new(identity.VerifiableAddress).TableName(ctx),
			new(identity.RecoveryAddress).TableName(ctx),
		} {
			/* #nosec G201 TableName is static */
			if err := tx.RawQuery(fmt.Sprintf(
				`DELETE FROM %s WHERE identity_id = ? AND nid = ?`, tn), i.ID, corp.ContextualizeNID(ctx, p.nid)).Exec(); err != nil {
				return err
			}
		}

		if err := p.update(WithTransaction(ctx, tx), i); err != nil {
			return err
		}

		if err := p.createVerifiableAddresses(ctx, i); err != nil {
			return err
		}

		if err := p.createRecoveryAddresses(ctx, i); err != nil {
			return err
		}

		return p.createIdentityCredentials(ctx, i)
	}))
}

func (p *Persister) DeleteIdentity(ctx context.Context, id uuid.UUID) error {
	return p.delete(ctx, new(identity.Identity), id)
}

func (p *Persister) GetIdentity(ctx context.Context, id uuid.UUID) (*identity.Identity, error) {
	var i identity.Identity
	if err := p.GetConnection(ctx).Where("id = ? AND nid = ?", id, corp.ContextualizeNID(ctx, p.nid)).First(&i); err != nil {
		return nil, sqlcon.HandleError(err)
	}

	i.Credentials = nil

	if err := p.findVerifiableAddresses(ctx, &i); err != nil {
		return nil, sqlcon.HandleError(err)
	}

	if err := p.findRecoveryAddresses(ctx, &i); err != nil {
		return nil, sqlcon.HandleError(err)
	}

	if err := p.injectTraitsSchemaURL(ctx, &i); err != nil {
		return nil, err
	}

	return &i, nil
}

func (p *Persister) GetIdentityConfidential(ctx context.Context, id uuid.UUID) (*identity.Identity, error) {
	var i identity.Identity

	nid := corp.ContextualizeNID(ctx, p.nid)
	if err := p.GetConnection(ctx).Where("id = ? AND nid = ?", id, nid).First(&i); err != nil {
		return nil, sqlcon.HandleError(err)
	}

	var creds identity.CredentialsCollection
	if err := p.GetConnection(ctx).Where("identity_id = ? AND nid = ?", id, nid).All(&creds); err != nil {
		return nil, sqlcon.HandleError(err)
	}

	i.Credentials = make(map[identity.CredentialsType]identity.Credentials)
	for k := range creds {
		cred := &creds[k]

		var ct identity.CredentialsTypeTable
		if err := p.GetConnection(ctx).Find(&ct, cred.CredentialTypeID); err != nil {
			return nil, sqlcon.HandleError(err)
		}
		cred.Type = ct.Name

		var cids identity.CredentialIdentifierCollection
		if err := p.GetConnection(ctx).Where("identity_credential_id = ? AND nid = ?", cred.ID, nid).All(&cids); err != nil {
			return nil, sqlcon.HandleError(err)
		}

		cred.Identifiers = make([]string, len(cids))
		for kk, cid := range cids {
			cred.Identifiers[kk] = cid.Identifier
		}

		i.Credentials[cred.Type] = *cred
	}

	if err := p.findRecoveryAddresses(ctx, &i); err != nil {
		return nil, err
	}
	if err := p.findVerifiableAddresses(ctx, &i); err != nil {
		return nil, err
	}

	if err := p.injectTraitsSchemaURL(ctx, &i); err != nil {
		return nil, err
	}

	return &i, nil
}

func (p *Persister) FindVerifiableAddressByValue(ctx context.Context, via identity.VerifiableAddressType, value string) (*identity.VerifiableAddress, error) {
	var address identity.VerifiableAddress
	if err := p.GetConnection(ctx).Where("nid = ? AND via = ? AND LOWER(value) = ?", corp.ContextualizeNID(ctx, p.nid), via, strings.ToLower(value)).First(&address); err != nil {
		return nil, sqlcon.HandleError(err)
	}

	return &address, nil
}

func (p *Persister) FindRecoveryAddressByValue(ctx context.Context, via identity.RecoveryAddressType, value string) (*identity.RecoveryAddress, error) {
	var address identity.RecoveryAddress
	if err := p.GetConnection(ctx).Where("nid = ? AND via = ? AND LOWER(value) = ?", corp.ContextualizeNID(ctx, p.nid), via, strings.ToLower(value)).First(&address); err != nil {
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
			"UPDATE %s SET status = ?, verified = true, verified_at = ?, code = ? WHERE nid = ? AND code = ? AND expires_at > ?",
			new(identity.VerifiableAddress).TableName(ctx),
		),
		identity.VerifiableAddressStatusCompleted,
		time.Now().UTC().Round(time.Second),
		newCode,
		corp.ContextualizeNID(ctx, p.nid),
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
	address.NID = corp.ContextualizeNID(ctx, p.nid)
	return p.update(ctx, address)
}

func (p *Persister) validateIdentity(ctx context.Context, i *identity.Identity) error {
	if err := p.r.IdentityValidator().ValidateWithRunner(ctx, i); err != nil {
		if _, ok := errorsx.Cause(err).(*jsonschema.ValidationError); ok {
			return errors.WithStack(herodot.ErrBadRequest.WithReasonf("%s", err))
		}
		return err
	}

	return nil
}

func (p *Persister) injectTraitsSchemaURL(ctx context.Context, i *identity.Identity) error {
	s, err := p.r.IdentityTraitsSchemas(ctx).GetByID(i.SchemaID)
	if err != nil {
		return errors.WithStack(herodot.ErrInternalServerError.WithReasonf(
			`The JSON Schema "%s" for this identity's traits could not be found.`, i.SchemaID))
	}
	i.SchemaURL = s.SchemaURL(p.r.Config(ctx).SelfPublicURL(nil)).String()
	return nil
}

func (p *Persister) getJsonSearchQuery(ctx context.Context, field string, values []string) pop.ScopeFunc {
	return func(q *pop.Query) *pop.Query {
		switch p.Connection(ctx).Dialect.Name() {
		case "sqlite3", "mysql", "mariadb":
			field, innerField := extractFieldAndInnerFields(field)
			for _, value := range values {
				if innerField == "" {
					q = q.Where(fmt.Sprintf(`json_extract(%s, '$') = ?`, field), value)
				} else {
					stmt := fmt.Sprintf(`json_extract(%s, '$.%s') = ?`, field, innerField)
					q = q.Where(stmt, value)
				}
			}
			return q
		case "postgres", "cockroach":
			field, innerField := extractFieldAndInnerFields(field)
			for _, value := range values {
				if innerField == "" {
					q = q.Where(fmt.Sprintf(`%s @> '"%s"'`, field, value))
				} else {
					q = q.Where(fmt.Sprintf(`%s @> '{"%s":"%s"}'`, field, innerField, value))
				}
			}
			return q
		}
		return q
	}
}

func (p *Persister) searchCredentialQuery(field string, values []string) pop.ScopeFunc {
	var innerField string
	_, innerField = extractFieldAndInnerFields(field)
	switch innerField {
	case "type":
		return func(q *pop.Query) *pop.Query { return q.Where("credential_types.name IN (?)", values) }
	case "identifier":
		return func(q *pop.Query) *pop.Query { return q.Where("credential_identifiers.identifier IN (?)", values) }
	default:
		return func(q *pop.Query) *pop.Query { return q }
	}
}

var quoteChar = map[string]uint8{
	"cockroach": '"',
	"mariadb":   '`',
	"mysql":     '`',
	"postgres":  '"',
	"sqlite3":   '"',
}

func (p *Persister) Quote(ctx context.Context, key string) string {
	n := p.Connection(ctx).Dialect.Name()
	c, ok := quoteChar[n]
	if !ok {
		// guess panic is OK here as the error is not fixable without a new release of Kratos
		panic("DSN is of unknown dialect " + n)
	}

	parts := strings.Split(key, ".")

	for i, part := range parts {
		part = strings.Trim(part, `"`)
		part = strings.Trim(part, "`")
		part = strings.TrimSpace(part)

		parts[i] = fmt.Sprintf(`%c%v%c`, c, part, c)
	}

	return strings.Join(parts, ".")
}
func (p *Persister) buildScope(ctx context.Context, queryValues url.Values) pop.ScopeFunc {
	return func(q *pop.Query) *pop.Query {

		for field, values := range queryValues {
			if stringslice.Has([]string{"page", "per_page"}, field) {
				continue
			}
			if ! p.validateFields(field) {
				p.r.Logger().Warning(`field ignored. does not respect this patterns [a-zA-Z0-9\._]+`)
				continue
			}
			if ! p.validateValues(values) {
				p.r.Logger().Warning(`values ignored. does not respect this patterns [%]+`)
				continue
			}
			if stringslice.Has([]string{"with_credentials"}, field) {
				q = q.LeftJoin("identity_credentials ic", "identities.id=ic.identity_id")
				q = q.LeftJoin("identity_credential_types credential_types", "credential_types.id=ic.identity_credential_type_id")
				q = q.LeftJoin("identity_credential_identifiers credential_identifiers", "credential_identifiers.identity_credential_id=ic.id")
				//q = q.LeftJoin("identity_credentials credentials", "credentials.identity_id=identities.id")
				continue
			}
			if strings.HasPrefix(field, "traits") {
				q = q.Scope(p.getJsonSearchQuery(ctx, field, values))
				continue
			}
			if strings.HasPrefix(field, "credentials") {
				q = q.Scope(p.searchCredentialQuery(field, values))
				continue
			}
			//q = q.Where("? = ?", field, values[0])
			field = p.Quote(ctx, field)
			q = q.Where(fmt.Sprintf("%s IN (?)", field), values)
		}
		return q
	}
}

func extractFieldAndInnerFields(field string) (string, string) {
	if !strings.Contains(field, ".") {
		return field, ""
	}
	dotIndex := strings.Index(field, ".")
	return field[:dotIndex], field[dotIndex+1:]
}

func (p *Persister) validateFields(field string) bool {
	res, err := regexp.MatchString(`[a-zA-Z0-9\._]+`, field)
	if err != nil {
		p.r.Logger().Errorf("validation field failed : %s",err.Error())
		return false
	}
	return res
}

func (p *Persister) validateValues(values []string) bool {
	prohibited := `[%]+`
	for _, value := range values {
		ok, err := regexp.MatchString(prohibited, value)
		if err != nil {
			p.r.Logger().Errorf("unable to check values parameter : %s", err.Error())
			return false
		}
		if ok {
			return false
		}
	}
	return true
}
