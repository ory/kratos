// nolint:deadcode,unused
package identity

// swagger:parameters createIdentity upsertIdentity
type swaggerParametersCreateIdentity struct {
	// in: body
	// required: true
	Body Identity
}

// swagger:parameters createIdentity
type createIdentityParameters struct {
	// required: true
	// in: body
	Body *Identity
}

// swagger:parameters updateIdentity
type updateIdentityParameters struct {
	// ID must be set to the ID of identity you want to update
	//
	// required: true
	// in: path
	ID string `json:"id"`

	// required: true
	// in: body
	Body *Identity
}

// swagger:parameters deleteIdentity
type deleteIdentityParameters struct {
	// ID is the identity's ID.
	//
	// required: true
	// in: path
	ID string `json:"id"`
}
