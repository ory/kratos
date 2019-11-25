package models

import (
	"encoding/json"
	"time"

	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/validate"
	"github.com/gofrs/uuid"
)

type LoginRequest struct {
	ID             uuid.UUID           `json:"id" db:"id"`
	ExpiresAt      time.Time           `json:"expires_at" faker:"time_type" db:"expires_at"`
	CreatedAt      time.Time           `json:"issued_at" faker:"time_type" db:"created_at"`
	UpdatedAt      time.Time           `json:"updated_at" faker:"time_type" db:"updated_at"`
	RequestURL     string              `json:"request_url" db:"request_url"`
	Active         CredentialsType     `json:"active,omitempty" db:"active"`
	Methods        LoginRequestMethods `json:"methods" faker:"login_request_methods" db:"methods"`
	RequestHeaders HTTPHeader          `json:"-" faker:"http_header" db:"request_headers"`
}

// String is not required by pop and may be deleted
func (l LoginRequest) String() string {
	jl, _ := json.Marshal(l)
	return string(jl)
}

// LoginRequests is not required by pop and may be deleted
type LoginRequests []LoginRequest

// String is not required by pop and may be deleted
func (l LoginRequests) String() string {
	jl, _ := json.Marshal(l)
	return string(jl)
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
// This method is not required and may be deleted.
func (l *LoginRequest) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateCreate gets run every time you call "pop.ValidateAndCreate" method.
// This method is not required and may be deleted.
func (l *LoginRequest) ValidateCreate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateUpdate gets run every time you call "pop.ValidateAndUpdate" method.
// This method is not required and may be deleted.
func (l *LoginRequest) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}
