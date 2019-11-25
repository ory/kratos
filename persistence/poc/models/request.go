package models

import (
	"encoding/json"
	"time"

	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/uuid"
	"github.com/gobuffalo/validate"
)

// type MethodType string

type Request struct {
	ID        uuid.UUID         `json:"id" db:"id" rw:"r"`
	CreatedAt time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt time.Time         `json:"updated_at" db:"updated_at"`
	PPMethods Methods           `json:"-" has_many:"methods" fk_id:"request_id"`
	Methods   map[string]Method `json:"methods" db:"-"`
}

// String is not required by pop and may be deleted
func (r Request) String() string {
	jr, _ := json.Marshal(r)
	return string(jr)
}

func (r *Request) BeforeSave(_ *pop.Connection) error {
	r.PPMethods = make(Methods, len(r.Methods))
	for _, m := range r.Methods {
		r.PPMethods = append(r.PPMethods, m)
	}
	r.Methods = nil
	return nil
}

func (r *Request) AfterFind(_ *pop.Connection) error {
	r.Methods = make(map[string]Method)
	for _, m := range r.PPMethods {
		r.Methods[m.ID.String()] = m
	}
	r.PPMethods = nil
	return nil
}

// Requests is not required by pop and may be deleted
type Requests []Request

// String is not required by pop and may be deleted
func (r Requests) String() string {
	jr, _ := json.Marshal(r)
	return string(jr)
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
// This method is not required and may be deleted.
func (r *Request) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateCreate gets run every time you call "pop.ValidateAndCreate" method.
// This method is not required and may be deleted.
func (r *Request) ValidateCreate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateUpdate gets run every time you call "pop.ValidateAndUpdate" method.
// This method is not required and may be deleted.
func (r *Request) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}
