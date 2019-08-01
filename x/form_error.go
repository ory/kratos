package x

import (
	"fmt"

	"github.com/pkg/errors"
)

type FormErrorCode int

const (
	ErrAuthnPasswordCredentialsInvalidCode FormErrorCode = iota + 1
	ErrSignupBadPayloadCode
	ErrSignupPasswordPolicyViolationCode
	ErrSignupIdentifierExistsCode
	ErrAuthnOIDCSubjectNotFoundCode
)

var (
	ErrAuthnPasswordCredentialsInvalid = FormError{
		MessageField: "The provided credentials are invalid. Check for spelling mistakes in your password or username, email address, or phone number.",
		CodeField:    ErrAuthnPasswordCredentialsInvalidCode,
		debug:        errors.New("the provided credentials are invalid"),
	}
	ErrSignupInvalidPayload = FormError{
		MessageField: "The form data contains invalid fields or other malformed data.",
		CodeField:    ErrSignupBadPayloadCode,
		debug:        errors.New("the form data contains invalid fields or other malformed data"),
	}
	ErrSignupIdentifierExists = FormError{
		MessageField: "A user with the given identifier (user id, email, phone, or username) exists already.",
		CodeField:    ErrSignupIdentifierExistsCode,
		debug:        errors.New("a user with the given identifier (email, phone, or username) exists already"),
	}
	ErrSignupPasswordPolicyViolation = FormError{
		MessageField: "The provided password does not meet the required password standard.",
		CodeField:    ErrSignupPasswordPolicyViolationCode,
		FieldField:   "password",
		debug:        errors.New("the provided password does not meet the required password standard"),
	}
	ErrAuthnOIDCSubjectNoRegistered = FormError{
		MessageField: "That account is not signed up yet. Please sign up first.",
		CodeField:    ErrAuthnOIDCSubjectNotFoundCode,
		debug:        errors.New("no identity could be found for that subject"),
	}
)

type FormError struct {
	MessageField string        `json:"message"`
	CodeField    FormErrorCode `json:"code,omitempty"`
	FieldField   string        `json:"field,omitempty"`
	debug        error         `json:"-"`
}

type stackTracer interface {
	StackTrace() errors.StackTrace
}

// StackTrace returns the error's stack trace.
func (e *FormError) StackTrace() errors.StackTrace {
	if e.debug == nil {
		return nil
	}

	if ee, ok := e.debug.(stackTracer); ok {
		return ee.StackTrace()
	}

	return nil
}

func (e *FormError) WithError(debug error) *FormError {
	err := *e
	err.debug = debug
	return &err
}

func (e *FormError) Error() string {
	if e.debug == nil {
		return ""
	}
	return e.debug.Error()
}

func (e *FormError) WithMessage(m string) *FormError {
	err := *e
	err.MessageField = m
	return &err
}

func (e *FormError) WithMessagef(m string, v ...interface{}) *FormError {
	err := *e
	err.MessageField = fmt.Sprintf(m, v...)
	return &err
}

func (e *FormError) WithField(field string) *FormError {
	err := *e
	err.FieldField = field
	return &err
}

func (e *FormError) Copy() *FormError {
	err := *e
	return &err
}
