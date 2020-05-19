package recovery

import (
	"net/http"
	"time"

	"github.com/gobuffalo/pop/v5"
	"github.com/gofrs/uuid"
	"github.com/pkg/errors"

	"github.com/ory/x/sqlxx"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/x"
)

type State string

const (
	StateBlank     = ""
	StatePending   = "pending"
	StateSent      = "sent"
	StateConfirmed = "confirmed"
	StateSuccess   = "success"
)

func NextState(current State) State {
	switch current {
	case StateBlank:
		return StatePending
	case StatePending:
		return StateSent
	case StateSent:
		return StateConfirmed
	case StateConfirmed:
		return StateSuccess
	}
	return StateBlank
}

// Request presents a recovery request
//
// This request is used when an identity wants to recover their account.
//
// We recommend reading the [Account Recovery Documentation](../self-service/flows/password-reset-account-recovery)
//
// swagger:model recoveryRequest
type Request struct {
	// ID represents the request's unique ID. When performing the recovery flow, this
	// represents the id in the recovery ui's query parameter: http://<urls.recovery_ui>?request=<id>
	//
	// required: true
	// type: string
	// format: uuid
	ID uuid.UUID `json:"id" db:"id" faker:"uuid" rw:"r"`

	// ExpiresAt is the time (UTC) when the request expires. If the user still wishes to update the setting,
	// a new request has to be initiated.
	//
	// required: true
	ExpiresAt time.Time `json:"expires_at" faker:"time_type" db:"expires_at"`

	// IssuedAt is the time (UTC) when the request occurred.
	//
	// required: true
	IssuedAt time.Time `json:"issued_at" faker:"time_type" db:"issued_at"`

	// RequestURL is the initial URL that was requested from ORY Kratos. It can be used
	// to forward information contained in the URL's path or query for example.
	//
	// required: true
	RequestURL string `json:"request_url" db:"request_url"`

	// Active, if set, contains the registration method that is being used. It is initially
	// not set.
	Active sqlxx.NullString `json:"active,omitempty" db:"active_method"`

	// Methods contains context for all account recovery methods. If a registration request has been
	// processed, but for example the password is incorrect, this will contain error messages.
	//
	// required: true
	Methods map[string]*RequestMethod `json:"methods" faker:"recovery_request_methods" db:"-"`

	// MethodsRaw is a helper struct field for gobuffalo.pop.
	MethodsRaw RequestMethodsRaw `json:"-" faker:"-" has_many:"selfservice_recovery_request_methods" fk_id:"selfservice_recovery_request_id"`

	// RecoveryAddress links this request to a recovery address.
	RecoveryAddress *identity.RecoveryAddress `json:"-" belongs_to:"identity_recovery_addresses" fk_id:"RecoveryAddressID"`

	// State represents the state of this request. Can be one of:
	//
	// - pending
	// - sent
	// - confirmed
	// - success
	//
	// required: true
	State State `json:"state" faker:"-" db:"state"`

	// CSRFToken contains the anti-csrf token associated with this request.
	CSRFToken string `json:"-" db:"csrf_token"`

	// CreatedAt is a helper struct field for gobuffalo.pop.
	CreatedAt time.Time `json:"-" faker:"-" db:"created_at"`
	// UpdatedAt is a helper struct field for gobuffalo.pop.
	UpdatedAt time.Time `json:"-" faker:"-" db:"updated_at"`
	// RecoveryAddressID is a helper struct field for gobuffalo.pop.
	RecoveryAddressID uuid.UUID `json:"-" faker:"-" db:"identity_recovery_address_id"`
}

func NewRequest(exp time.Duration,csrf string, r *http.Request) *Request {
	return &Request{
		ID:         x.NewUUID(),
		ExpiresAt:  time.Now().UTC().Add(exp),
		IssuedAt:   time.Now().UTC(),
		RequestURL: x.RequestURL(r).String(),
		Methods:    map[string]*RequestMethod{},
		State:      NextState(StateBlank),
		CSRFToken:  csrf,
	}
}

func (r *Request) TableName() string {
	return "selfservice_recovery_requests"
}

func (r *Request) GetID() uuid.UUID {
	return r.ID
}

func (r *Request) Valid(s *session.Session) error {
	if r.ExpiresAt.Before(time.Now().UTC()) {
		return errors.WithStack(ErrRequestExpired.
			WithReasonf("The recovery request expired %.2f minutes ago, please try again.",
				-time.Since(r.ExpiresAt).Minutes()))
	}
	return nil
}

func (r *Request) BeforeSave(_ *pop.Connection) error {
	r.MethodsRaw = make([]RequestMethod, 0, len(r.Methods))
	for _, m := range r.Methods {
		r.MethodsRaw = append(r.MethodsRaw, *m)
	}
	r.Methods = nil
	return nil
}

func (r *Request) AfterSave(c *pop.Connection) error {
	return r.AfterFind(c)
}

func (r *Request) AfterFind(_ *pop.Connection) error {
	r.Methods = make(RequestMethods)
	for key := range r.MethodsRaw {
		m := r.MethodsRaw[key] // required for pointer dereference
		r.Methods[m.Method] = &m
	}
	r.MethodsRaw = nil
	return nil
}
