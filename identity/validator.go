package identity

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ory/x/sqlfields"

	"github.com/pkg/errors"

	"github.com/ory/herodot"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/schema"
)

type (
	validatorDependencies interface {
		IdentityTraitsSchemas(ctx context.Context) (schema.Schemas, error)
		config.Provider
	}
	Validator struct {
		v *schema.Validator
		d validatorDependencies
	}
	ValidationProvider interface {
		IdentityValidator() *Validator
	}
)

func NewValidator(d validatorDependencies) *Validator {
	return &Validator{v: schema.NewValidator(), d: d}
}

func (v *Validator) ValidateWithRunner(ctx context.Context, i *Identity, runners ...schema.Extension) error {
	runner, err := schema.NewExtensionRunner(ctx, runners...)
	if err != nil {
		return err
	}

	ss, err := v.d.IdentityTraitsSchemas(ctx)
	if err != nil {
		return err
	}

	s, err := ss.GetByID(i.SchemaID)
	if err != nil {
		return err
	}

	var metadataPublic *sqlfields.NullJSONRawMessage
	if i.MetadataPublic.Valid {
		metadataPublic = &i.MetadataPublic
	}

	validationFields := struct {
		Traits         Traits                        `json:"traits"`
		MetadataPublic *sqlfields.NullJSONRawMessage `json:"metadata_public,omitempty"`
		MetadataAdmin  *sqlfields.NullJSONRawMessage `json:"metadata_admin,omitempty"`
	}{
		Traits:         i.Traits,
		MetadataPublic: metadataPublic,
		MetadataAdmin:  i.MetadataAdmin,
	}
	rawIdentity, err := json.Marshal(validationFields)
	if err != nil {
		return errors.WithStack(herodot.ErrInternalServerError.WithError(err.Error()).WithReason("Unable to marshal identity for validation."))
	}
	fmt.Println(string(rawIdentity))
	fmt.Printf("%#v %#v\n", validationFields.MetadataAdmin, validationFields.MetadataPublic)

	return v.v.Validate(ctx, s.URL.String(), rawIdentity, schema.WithExtensionRunner(runner))
}

func (v *Validator) Validate(ctx context.Context, i *Identity) error {
	return v.ValidateWithRunner(ctx, i,
		NewSchemaExtensionCredentials(i),
		NewSchemaExtensionVerification(i, v.d.Config().SelfServiceFlowVerificationRequestLifespan(ctx)),
		NewSchemaExtensionRecovery(i),
	)
}
