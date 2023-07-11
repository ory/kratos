// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package identity

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/samber/lo"

	"github.com/tidwall/sjson"

	"github.com/tidwall/gjson"

	"github.com/ory/kratos/cipher"

	"github.com/ory/herodot"
	"github.com/ory/x/sqlxx"

	"github.com/ory/kratos/driver/config"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
)

// An Identity's State
//
// The state can either be `active` or `inactive`.
//
// swagger:model identityState
type State string

const (
	StateActive   State = "active"
	StateInactive State = "inactive"
)

func (lt State) IsValid() error {
	switch lt {
	case StateActive, StateInactive:
		return nil
	}
	return errors.New("identity state is not valid")
}

// Identity represents an Ory Kratos identity
//
// An [identity](https://www.ory.sh/docs/kratos/concepts/identity-user-model) represents a (human) user in Ory.
//
// swagger:model identity
type Identity struct {
	l *sync.RWMutex `db:"-" faker:"-"`

	// ID is the identity's unique identifier.
	//
	// The Identity ID can not be changed and can not be chosen. This ensures future
	// compatibility and optimization for distributed stores such as CockroachDB.
	//
	// required: true
	ID uuid.UUID `json:"id" faker:"-" db:"id"`

	// Credentials represents all credentials that can be used for authenticating this identity.
	Credentials map[CredentialsType]Credentials `json:"credentials,omitempty" faker:"-" db:"-"`

	// AvailableAAL defines the maximum available AAL for this identity. If the user has only a password
	// configured, the AAL will be 1. If the user has a password and a TOTP configured, the AAL will be 2.
	AvailableAAL NullableAuthenticatorAssuranceLevel `json:"-" faker:"-" db:"available_aal"`

	// // IdentifierCredentials contains the access and refresh token for oidc identifier
	// IdentifierCredentials []IdentifierCredential `json:"identifier_credentials,omitempty" faker:"-" db:"-"`

	// SchemaID is the ID of the JSON Schema to be used for validating the identity's traits.
	//
	// required: true
	SchemaID string `json:"schema_id" faker:"-" db:"schema_id"`

	// SchemaURL is the URL of the endpoint where the identity's traits schema can be fetched from.
	//
	// format: url
	// required: true
	SchemaURL string `json:"schema_url" faker:"-" db:"-"`

	// State is the identity's state.
	//
	// This value has currently no effect.
	State State `json:"state" faker:"-" db:"state"`

	// StateChangedAt contains the last time when the identity's state changed.
	StateChangedAt *sqlxx.NullTime `json:"state_changed_at,omitempty" faker:"-" db:"state_changed_at"`

	// Traits represent an identity's traits. The identity is able to create, modify, and delete traits
	// in a self-service manner. The input will always be validated against the JSON Schema defined
	// in `schema_url`.
	//
	// required: true
	Traits Traits `json:"traits" faker:"-" db:"traits"`

	// VerifiableAddresses contains all the addresses that can be verified by the user.
	//
	// Extensions:
	// ---
	// x-omitempty: true
	// ---
	VerifiableAddresses []VerifiableAddress `json:"verifiable_addresses,omitempty" faker:"-" has_many:"identity_verifiable_addresses" fk_id:"identity_id" order_by:"id asc"`

	// RecoveryAddresses contains all the addresses that can be used to recover an identity.
	//
	// Extensions:
	// ---
	// x-omitempty: true
	// ---
	RecoveryAddresses []RecoveryAddress `json:"recovery_addresses,omitempty" faker:"-" has_many:"identity_recovery_addresses" fk_id:"identity_id" order_by:"id asc"`

	// Store metadata about the identity which the identity itself can see when calling for example the
	// session endpoint. Do not store sensitive information (e.g. credit score) about the identity in this field.
	MetadataPublic sqlxx.NullJSONRawMessage `json:"metadata_public" faker:"-" db:"metadata_public"`

	// Store metadata about the user which is only accessible through admin APIs such as `GET /admin/identities/<id>`.
	MetadataAdmin sqlxx.NullJSONRawMessage `json:"metadata_admin,omitempty" faker:"-" db:"metadata_admin"`

	// CreatedAt is a helper struct field for gobuffalo.pop.
	CreatedAt time.Time `json:"created_at" db:"created_at"`

	// UpdatedAt is a helper struct field for gobuffalo.pop.
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	NID       uuid.UUID `json:"-"  faker:"-" db:"nid"`
}

// Traits represent an identity's traits. The identity is able to create, modify, and delete traits
// in a self-service manner. The input will always be validated against the JSON Schema defined
// in `schema_url`.
//
// swagger:model identityTraits
type Traits json.RawMessage

func (t *Traits) Scan(value interface{}) error {
	return sqlxx.JSONScan(t, value)
}

func (t Traits) Value() (driver.Value, error) {
	return sqlxx.JSONValue(t)
}

func (t *Traits) String() string {
	return string(*t)
}

// MarshalJSON returns m as the JSON encoding of m.
func (t Traits) MarshalJSON() ([]byte, error) {
	if t == nil {
		return []byte("null"), nil
	}
	return t, nil
}

// UnmarshalJSON sets *m to a copy of data.
func (t *Traits) UnmarshalJSON(data []byte) error {
	if t == nil {
		return errors.New("json.RawMessage: UnmarshalJSON on nil pointer")
	}
	*t = append((*t)[0:0], data...)
	return nil
}

func (i Identity) TableName(ctx context.Context) string {
	return "identities"
}

func (i *Identity) lock() *sync.RWMutex {
	if i.l == nil {
		i.l = new(sync.RWMutex)
	}
	return i.l
}

func (i *Identity) IsActive() bool {
	return i.State == StateActive
}

func (i *Identity) SetCredentials(t CredentialsType, c Credentials) {
	i.lock().Lock()
	defer i.lock().Unlock()
	if i.Credentials == nil {
		i.Credentials = make(map[CredentialsType]Credentials)
	}

	c.Type = t
	i.Credentials[t] = c
}

func (i *Identity) SetCredentialsWithConfig(t CredentialsType, c Credentials, conf interface{}) (err error) {
	i.lock().Lock()
	defer i.lock().Unlock()
	if i.Credentials == nil {
		i.Credentials = make(map[CredentialsType]Credentials)
	}

	c.Config, err = json.Marshal(conf)
	if err != nil {
		return errors.WithStack(herodot.ErrInternalServerError.WithReasonf("Unable to encode %s credentials to JSON: %s", t, err))
	}

	c.Type = t
	i.Credentials[t] = c
	return nil
}

func (i *Identity) DeleteCredentialsType(t CredentialsType) {
	i.lock().Lock()
	defer i.lock().Unlock()
	if i.Credentials == nil {
		return
	}

	delete(i.Credentials, t)
}

func (i *Identity) GetCredentialsOr(t CredentialsType, or *Credentials) *Credentials {
	c, ok := i.GetCredentials(t)
	if !ok {
		return or
	}
	return c
}

func (i *Identity) UpsertCredentialsConfig(t CredentialsType, conf []byte, version int) {
	c, ok := i.GetCredentials(t)
	if !ok {
		c = &Credentials{}
	}

	c.Type = t
	c.IdentityID = i.ID
	c.Config = conf
	c.Version = version

	i.SetCredentials(t, *c)
}

func (i *Identity) GetCredentials(t CredentialsType) (*Credentials, bool) {
	i.lock().RLock()
	defer i.lock().RUnlock()

	if c, ok := i.Credentials[t]; ok {
		return &c, true
	}

	return nil, false
}

func (i *Identity) ParseCredentials(t CredentialsType, config interface{}) (*Credentials, error) {
	i.lock().RLock()
	defer i.lock().RUnlock()

	if c, ok := i.Credentials[t]; ok {
		if err := json.Unmarshal(c.Config, config); err != nil {
			return nil, errors.WithStack(err)
		}
		return &c, nil
	}

	return nil, herodot.ErrNotFound.WithReasonf("identity does not have credential type %s", t)
}

func (i *Identity) CopyWithoutCredentials() *Identity {
	i.lock().RLock()
	defer i.lock().RUnlock()
	ii := *i
	ii.l = new(sync.RWMutex)
	ii.Credentials = nil
	return &ii
}

func NewIdentity(traitsSchemaID string) *Identity {
	if traitsSchemaID == "" {
		traitsSchemaID = config.DefaultIdentityTraitsSchemaID
	}

	stateChangedAt := sqlxx.NullTime(time.Now().UTC())
	return &Identity{
		ID:                  uuid.Nil,
		Credentials:         map[CredentialsType]Credentials{},
		Traits:              Traits("{}"),
		SchemaID:            traitsSchemaID,
		VerifiableAddresses: []VerifiableAddress{},
		State:               StateActive,
		StateChangedAt:      &stateChangedAt,
		l:                   new(sync.RWMutex),
	}
}

func (i Identity) GetID() uuid.UUID {
	return i.ID
}

func (i Identity) GetNID() uuid.UUID {
	return i.NID
}

func (i Identity) MarshalJSON() ([]byte, error) {
	type localIdentity Identity
	i.Credentials = nil
	i.MetadataAdmin = nil
	result, err := json.Marshal(localIdentity(i))
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (i *Identity) UnmarshalJSON(b []byte) error {
	type localIdentity Identity
	err := json.Unmarshal(b, (*localIdentity)(i))
	i.Credentials = nil
	i.MetadataAdmin = nil
	return err
}

func (i *Identity) SetAvailableAAL(ctx context.Context, m *Manager) (err error) {
	i.AvailableAAL = NewNullableAuthenticatorAssuranceLevel(NoAuthenticatorAssuranceLevel)
	if c, err := m.CountActiveFirstFactorCredentials(ctx, i); err != nil {
		return err
	} else if c == 0 {
		// No first factor set up - AAL is 0
		return nil
	}

	i.AvailableAAL = NewNullableAuthenticatorAssuranceLevel(AuthenticatorAssuranceLevel1)
	if c, err := m.CountActiveMultiFactorCredentials(ctx, i); err != nil {
		return err
	} else if c == 0 {
		// No second factor set up - AAL is 1
		return nil
	}

	i.AvailableAAL = NewNullableAuthenticatorAssuranceLevel(AuthenticatorAssuranceLevel2)
	return nil
}

type WithAdminMetadataInJSON Identity

func (i WithAdminMetadataInJSON) MarshalJSON() ([]byte, error) {
	type localIdentity Identity
	i.Credentials = nil
	return json.Marshal(localIdentity(i))
}

type WithCredentialsAndAdminMetadataInJSON Identity

func (i WithCredentialsAndAdminMetadataInJSON) MarshalJSON() ([]byte, error) {
	type localIdentity Identity
	return json.Marshal(localIdentity(i))
}

type WithCredentialsMetadataAndAdminMetadataInJSON Identity

func (i WithCredentialsMetadataAndAdminMetadataInJSON) MarshalJSON() ([]byte, error) {
	type localIdentity Identity
	for k, v := range i.Credentials {
		v.Config = nil
		i.Credentials[k] = v
	}
	return json.Marshal(localIdentity(i))
}

func (i *Identity) Validate() error {
	expected := i.NID
	if expected == uuid.Nil {
		return errors.WithStack(herodot.ErrInternalServerError.WithReason("Received empty nid."))
	}

	i.RecoveryAddresses = lo.Filter(i.RecoveryAddresses, func(v RecoveryAddress, key int) bool {
		return v.NID == expected
	})

	i.VerifiableAddresses = lo.Filter(i.VerifiableAddresses, func(v VerifiableAddress, key int) bool {
		return v.NID == expected
	})

	for k := range i.Credentials {
		c := i.Credentials[k]
		if c.NID != expected {
			delete(i.Credentials, k)
			continue
		}
	}

	return nil
}

// CollectVerifiableAddresses returns a slice of all verifiable addresses of the given identities.
func CollectVerifiableAddresses(i []*Identity) (res []VerifiableAddress) {
	res = make([]VerifiableAddress, 0, len(i))
	for _, id := range i {
		res = append(res, id.VerifiableAddresses...)
	}

	return res
}

// CollectRecoveryAddresses returns a slice of all recovery addresses of the given identities.
func CollectRecoveryAddresses(i []*Identity) (res []RecoveryAddress) {
	res = make([]RecoveryAddress, 0, len(i))
	for _, id := range i {
		res = append(res, id.RecoveryAddresses...)
	}

	return res
}

func (i *Identity) WithDeclassifiedCredentials(ctx context.Context, c cipher.Provider, includeCredentials []CredentialsType) (*Identity, error) {
	credsToPublish := make(map[CredentialsType]Credentials)

	for ct, original := range i.Credentials {
		if _, found := lo.Find(includeCredentials, func(i CredentialsType) bool {
			return i == ct
		}); !found {
			toPublish := original
			toPublish.Config = []byte{}
			credsToPublish[ct] = toPublish
			continue
		}

		switch ct {
		case CredentialsTypeOIDC:
			toPublish := original
			toPublish.Config = []byte{}

			for _, token := range []string{"initial_id_token", "initial_access_token", "initial_refresh_token"} {
				var i int
				var err error
				gjson.GetBytes(original.Config, "providers").ForEach(func(_, v gjson.Result) bool {
					key := fmt.Sprintf("%d.%s", i, token)
					ciphertext := v.Get(token).String()

					var plaintext []byte
					plaintext, err = c.Cipher(ctx).Decrypt(ctx, ciphertext)
					if err != nil {
						return false
					}

					toPublish.Config, err = sjson.SetBytes(toPublish.Config, "providers."+key, string(plaintext))
					if err != nil {
						return false
					}

					toPublish.Config, err = sjson.SetBytes(toPublish.Config, fmt.Sprintf("providers.%d.subject", i), v.Get("subject").String())
					if err != nil {
						return false
					}

					toPublish.Config, err = sjson.SetBytes(toPublish.Config, fmt.Sprintf("providers.%d.provider", i), v.Get("provider").String())
					if err != nil {
						return false
					}

					i++
					return true
				})

				if err != nil {
					return nil, err
				}

				credsToPublish[ct] = toPublish
			}
		default:
			credsToPublish[ct] = original
		}
	}

	ii := *i
	ii.Credentials = credsToPublish
	return &ii, nil
}

// Patch Identities Parameters
//
// swagger:parameters batchPatchIdentities
//
//nolint:deadcode,unused
//lint:ignore U1000 Used to generate Swagger and OpenAPI definitions
type batchPatchIdentitites struct {
	// in: body
	Body BatchPatchIdentitiesBody
}

// Patch Identities Body
//
// swagger:model patchIdentitiesBody
//
//lint:ignore U1000 Used to generate Swagger and OpenAPI definitions
type BatchPatchIdentitiesBody struct {
	// Identities holds the list of patches to apply
	//
	// required
	Identities []*BatchIdentityPatch `json:"identities"`

	// Future fields:
	// RemotePatchesURL string
	// Async bool
}

// Payload for patching an identity
//
// swagger:model identityPatch
//
//lint:ignore U1000 Used to generate Swagger and OpenAPI definitions
type BatchIdentityPatch struct {
	// The identity to create.
	Create *CreateIdentityBody `json:"create"`

	// The ID of this patch.
	//
	// The patch ID is optional. If specified, the ID will be returned in the
	// response, so consumers of this API can correlate the response with the
	// patch.
	ID *uuid.UUID `json:"patch_id"`
}

// swagger:enum BatchPatchAction
//
//lint:ignore U1000 Used to generate Swagger and OpenAPI definitions
type BatchPatchAction string

const (
	// Create this identity.
	ActionCreate BatchPatchAction = "create"

	// Future actions:
	//
	// Delete this identity.
	// ActionDelete BatchPatchAction = "delete"
	//
	// ActionUpdate BatchPatchAction = "update"
)

// Patch identities response
//
// swagger:model batchPatchIdentitiesResponse
//
//lint:ignore U1000 Used to generate Swagger and OpenAPI definitions
type batchPatchIdentitiesResponse struct {
	// The patch responses for the individual identities.
	Identities []*BatchIdentityPatchResponse `json:"identities"`
}

// Response for a single identity patch
//
// swagger:model identityPatchResponse
//
//lint:ignore U1000 Used to generate Swagger and OpenAPI definitions
type BatchIdentityPatchResponse struct {
	// The action for this specific patch
	Action BatchPatchAction `json:"action"`

	// The identity ID payload of this patch
	IdentityID *uuid.UUID `json:"identity,omitempty"`

	// The ID of this patch response, if an ID was specified in the patch.
	PatchID *uuid.UUID `json:"patch_id,omitempty"`
}
