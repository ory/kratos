package verify

import (
	"net/http"
	"net/url"
	"time"

	"github.com/gofrs/uuid"

	"github.com/ory/x/urlx"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/form"
	"github.com/ory/kratos/x"
)

// Request presents a verification request
//
// This request is used when an identity wants to verify an out-of-band communication
// channel such as an email address or a phone number.
//
// For more information head over to: https://www.ory.sh/docs/kratos/selfservice/flows/verify-email-account-activation
//
// swagger:model verificationRequest
type Request struct {
	// ID represents the request's unique ID. When performing the profile management flow, this
	// represents the id in the profile ui's query parameter: http://<urls.profile_ui>?request=<id>
	//
	// type: string
	// format: uuid
	ID uuid.UUID `json:"id" db:"id" faker:"uuid" rw:"r"`

	// ExpiresAt is the time (UTC) when the request expires. If the user still wishes to update the profile,
	// a new request has to be initiated.
	ExpiresAt time.Time `json:"expires_at" faker:"time_type" db:"expires_at"`

	// IssuedAt is the time (UTC) when the request occurred.
	IssuedAt time.Time `json:"issued_at" faker:"time_type" db:"issued_at"`

	// RequestURL is the initial URL that was requested from ORY Kratos. It can be used
	// to forward information contained in the URL's path or query for example.
	RequestURL string `json:"request_url" db:"request_url"`

	// TraitsForm contains form fields, errors, and so on.
	Form *form.HTMLForm `json:"form" faker:"-" db:"form"`

	Via identity.VerifiableAddressType `json:"via" db:"via"`

	// CSRFToken contains the anti-csrf token associated with this request.
	CSRFToken string `json:"-" db:"csrf_token"`

	// Success, if true, implies that the request was completed successfully.
	Success bool `json:"success" db:"success"`

	// CreatedAt is a helper struct field for gobuffalo.pop.
	CreatedAt time.Time `json:"-" faker:"-" db:"created_at"`
	// UpdatedAt is a helper struct field for gobuffalo.pop.
	UpdatedAt time.Time `json:"-" faker:"-" db:"updated_at"`
}

func (r Request) TableName() string {
	return "selfservice_verification_requests"
}

func NewRequest(
	exp time.Duration, r *http.Request, via identity.VerifiableAddressType, action *url.URL, generator form.CSRFGenerator) *Request {
	source := urlx.Copy(r.URL)
	source.Host = r.Host

	if len(source.Scheme) == 0 {
		source.Scheme = "http"
		if r.TLS != nil {
			source.Scheme = "https"
		}
	}

	id := x.NewUUID()
	csrf := generator(r)

	f := form.NewHTMLForm(urlx.CopyWithQuery(action, url.Values{"request": {id.String()}}).String())
	f.SetCSRF(csrf)
	f.SetField(form.Field{
		Name:     "to_verify",
		Type:     via.HTMLFormInputType(),
		Required: true,
	})

	return &Request{
		ID:         id,
		ExpiresAt:  time.Now().UTC().Add(exp),
		IssuedAt:   time.Now().UTC(),
		RequestURL: source.String(),
		Form:       f,
		CSRFToken:  csrf,
		Via:        via,
	}
}

func (r *Request) Valid() error {
	if r.ExpiresAt.Before(time.Now()) {
		return newErrRequestRequired(time.Since(r.ExpiresAt).Minutes())
	}
	return nil
}
