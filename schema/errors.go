// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package schema

import (
	"fmt"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/pkg/errors"

	"github.com/ory/jsonschema/v3"

	"github.com/ory/kratos/text"
)

type ValidationError struct {
	*jsonschema.ValidationError
	Messages text.Messages
}

func NewRequiredError(missingPtr, missingFieldName string) error {
	return errors.WithStack(&ValidationError{
		ValidationError: &jsonschema.ValidationError{
			Message:     fmt.Sprintf("missing properties: %s", missingFieldName),
			InstancePtr: missingPtr,
			Context: &jsonschema.ValidationErrorContextRequired{
				Missing: []string{missingFieldName},
			},
		},
		Messages: new(text.Messages).Add(text.NewValidationErrorRequired(missingFieldName)),
	})
}

func NewTOTPVerifierWrongError(instancePtr string) error {
	t := text.NewErrorValidationTOTPVerifierWrong()
	return errors.WithStack(&ValidationError{
		ValidationError: &jsonschema.ValidationError{
			Message:     t.Text,
			InstancePtr: instancePtr,
		},
		Messages: new(text.Messages).Add(t),
	})
}

func NewWebAuthnVerifierWrongError(instancePtr string) error {
	t := text.NewErrorValidationTOTPVerifierWrong()
	return errors.WithStack(&ValidationError{
		ValidationError: &jsonschema.ValidationError{
			Message:     t.Text,
			InstancePtr: instancePtr,
		},
		Messages: new(text.Messages).Add(t),
	})
}

func NewLookupAlreadyUsed() error {
	t := text.NewErrorValidationLookupAlreadyUsed()
	return errors.WithStack(&ValidationError{
		ValidationError: &jsonschema.ValidationError{
			Message:     t.Text,
			InstancePtr: "#/",
		},
		Messages: new(text.Messages).Add(t),
	})
}

func NewErrorValidationLookupInvalid() error {
	t := text.NewErrorValidationLookupInvalid()
	return errors.WithStack(&ValidationError{
		ValidationError: &jsonschema.ValidationError{
			Message:     t.Text,
			InstancePtr: "#/",
		},
		Messages: new(text.Messages).Add(t),
	})
}

type ValidationErrorContextPasswordPolicyViolation struct {
	Reason string
}

func (r *ValidationErrorContextPasswordPolicyViolation) AddContext(_, _ string) {}

func (r *ValidationErrorContextPasswordPolicyViolation) FinishInstanceContext() {}

func NewPasswordPolicyViolationError(instancePtr string, reason string) error {
	return errors.WithStack(&ValidationError{
		ValidationError: &jsonschema.ValidationError{
			Message:     fmt.Sprintf("the password does not fulfill the password policy because: %s", reason),
			InstancePtr: instancePtr,
			Context: &ValidationErrorContextPasswordPolicyViolation{
				Reason: reason,
			},
		},
		Messages: new(text.Messages).Add(text.NewErrorValidationPasswordPolicyViolation(reason)),
	})
}

func NewMissingIdentifierError() error {
	return errors.WithStack(&ValidationError{
		ValidationError: &jsonschema.ValidationError{
			Message:     "could not find any identifiers",
			InstancePtr: "#/",
		},
		Messages: new(text.Messages).Add(text.NewErrorValidationIdentifierMissing()),
	})
}

type ValidationErrorContextInvalidCredentialsError struct{}

func (r *ValidationErrorContextInvalidCredentialsError) AddContext(_, _ string) {}

func (r *ValidationErrorContextInvalidCredentialsError) FinishInstanceContext() {}

func NewInvalidCredentialsError() error {
	return errors.WithStack(&ValidationError{
		ValidationError: &jsonschema.ValidationError{
			Message:     `the provided credentials are invalid, check for spelling mistakes in your password or username, email address, or phone number`,
			InstancePtr: "#/",
			Context:     &ValidationErrorContextPasswordPolicyViolation{},
		},
		Messages: new(text.Messages).Add(text.NewErrorValidationInvalidCredentials()),
	})
}

type ValidationErrorContextDuplicateCredentialsError struct {
	AvailableCredentials   []string `json:"available_credential_types"`
	AvailableOIDCProviders []string `json:"available_oidc_providers"`
	IdentifierHint         string   `json:"credential_identifier_hint"`
}

func (r *ValidationErrorContextDuplicateCredentialsError) AddContext(_, _ string) {}

func (r *ValidationErrorContextDuplicateCredentialsError) FinishInstanceContext() {}

type DuplicateCredentialsHinter interface {
	AvailableCredentials() []string
	AvailableOIDCProviders() []string
	IdentifierHint() string
	HasHints() bool
}

func NewDuplicateCredentialsError(err error) error {
	if hinter := DuplicateCredentialsHinter(nil); errors.As(err, &hinter) && hinter.HasHints() {
		oidcProviders := make([]string, 0, len(hinter.AvailableOIDCProviders()))
		for _, provider := range hinter.AvailableOIDCProviders() {
			oidcProviders = append(oidcProviders, cases.Title(language.English).String(provider))
		}

		reason := ""
		if hinter.IdentifierHint() != "" {
			reason = fmt.Sprintf("You tried signing with %s which is already in use by another account.", hinter.IdentifierHint())
		} else {
			reason = "You tried to sign up using an email, phone, or username that is already used by another account."
		}

		if len(hinter.AvailableCredentials()) > 0 {
			humanReadable := make([]string, 0, len(hinter.AvailableCredentials()))
			for _, cred := range hinter.AvailableCredentials() {
				switch cred {
				case "password":
					humanReadable = append(humanReadable, "your password")
				case "oidc":
					humanReadable = append(humanReadable, "social sign in")
				case "webauthn":
					humanReadable = append(humanReadable, "your PassKey or a security key")
				}
			}

			reason = fmt.Sprintf("%s You can sign in using %s.", reason, strings.Join(humanReadable, ", "))
			if len(hinter.AvailableOIDCProviders()) > 0 {
				reason = fmt.Sprintf("%s Use one of the following social sign in providers: %s", reason, strings.Join(oidcProviders, ", "))
			}
		} else if len(hinter.AvailableOIDCProviders()) > 0 {
			reason = fmt.Sprintf("%s You can sign in using one of the following social sign in providers: %s.", reason, strings.Join(oidcProviders, ", "))
		}

		return errors.WithStack(&ValidationError{
			ValidationError: &jsonschema.ValidationError{
				Message:     `an account with the same identifier (email, phone, username, ...) exists already`,
				InstancePtr: "#/",
				Context: &ValidationErrorContextDuplicateCredentialsError{
					AvailableCredentials:   hinter.AvailableCredentials(),
					AvailableOIDCProviders: hinter.AvailableOIDCProviders(),
					IdentifierHint:         hinter.IdentifierHint(),
				},
			},
			Messages: new(text.Messages).Add(text.NewErrorValidationDuplicateCredentialsWithHints(reason, hinter.AvailableCredentials(), hinter.AvailableOIDCProviders(), hinter.IdentifierHint())),
		})
	}

	return errors.WithStack(&ValidationError{
		ValidationError: &jsonschema.ValidationError{
			Message:     `an account with the same identifier (email, phone, username, ...) exists already`,
			InstancePtr: "#/",
			Context:     &ValidationErrorContextDuplicateCredentialsError{},
		},
		Messages: new(text.Messages).Add(text.NewErrorValidationDuplicateCredentials()),
	})
}

func NewNoLoginStrategyResponsible() error {
	return errors.WithStack(&ValidationError{
		ValidationError: &jsonschema.ValidationError{
			Message:     `could not find a strategy to login with`,
			InstancePtr: "#/",
		},
		Messages: new(text.Messages).Add(text.NewErrorValidationLoginNoStrategyFound()),
	})
}

func NewNoRegistrationStrategyResponsible() error {
	return errors.WithStack(&ValidationError{
		ValidationError: &jsonschema.ValidationError{
			Message:     `could not find a strategy to sign up with`,
			InstancePtr: "#/",
		},
		Messages: new(text.Messages).Add(text.NewErrorValidationRegistrationNoStrategyFound()),
	})
}

func NewNoSettingsStrategyResponsible() error {
	return errors.WithStack(&ValidationError{
		ValidationError: &jsonschema.ValidationError{
			Message:     `could not find a strategy to update settings with`,
			InstancePtr: "#/",
		},
		Messages: new(text.Messages).Add(text.NewErrorValidationSettingsNoStrategyFound()),
	})
}

func NewNoRecoveryStrategyResponsible() error {
	return errors.WithStack(&ValidationError{
		ValidationError: &jsonschema.ValidationError{
			Message:     `could not find a strategy to recover your account with`,
			InstancePtr: "#/",
		},
		Messages: new(text.Messages).Add(text.NewErrorValidationRecoveryNoStrategyFound()),
	})
}

func NewNoVerificationStrategyResponsible() error {
	return errors.WithStack(&ValidationError{
		ValidationError: &jsonschema.ValidationError{
			Message:     `could not find a strategy to verify your account with`,
			InstancePtr: "#/",
		},
		Messages: new(text.Messages).Add(text.NewErrorValidationVerificationNoStrategyFound()),
	})
}

func NewAddressNotVerifiedError() error {
	return errors.WithStack(&ValidationError{
		ValidationError: &jsonschema.ValidationError{
			Message:     `account address not yet verified`,
			InstancePtr: "#/",
		},
		Messages: new(text.Messages).Add(text.NewErrorValidationAddressNotVerified()),
	})
}

func NewNoTOTPDeviceRegistered() error {
	return errors.WithStack(&ValidationError{
		ValidationError: &jsonschema.ValidationError{
			Message:     `you have no TOTP device set up`,
			InstancePtr: "#/",
		},
		Messages: new(text.Messages).Add(text.NewErrorValidationNoTOTPDevice()),
	})
}

func NewNoLookupDefined() error {
	return errors.WithStack(&ValidationError{
		ValidationError: &jsonschema.ValidationError{
			Message:     `you have no backup recovery codes set up`,
			InstancePtr: "#/",
		},
		Messages: new(text.Messages).Add(text.NewErrorValidationNoLookup()),
	})
}

func NewNoWebAuthnRegistered() error {
	return errors.WithStack(&ValidationError{
		ValidationError: &jsonschema.ValidationError{
			Message:     `you have no WebAuthn device set up`,
			InstancePtr: "#/",
		},
		Messages: new(text.Messages).Add(text.NewErrorValidationNoWebAuthnDevice()),
	})
}

func NewHookValidationError(instancePtr, message string, messages text.Messages) *ValidationError {
	return &ValidationError{
		ValidationError: &jsonschema.ValidationError{
			Message:     message,
			InstancePtr: instancePtr,
		},
		Messages: messages,
	}
}

type ValidationListError struct {
	Validations []*ValidationError
}

func (e ValidationListError) Error() string {
	var detailError string
	for pos, validationErr := range e.Validations {
		detailError = detailError + fmt.Sprintf("\n(%d) %s", pos, validationErr.Error())
	}
	return fmt.Sprintf("%d validation errors occurred:%s", len(e.Validations), detailError)
}

func (e *ValidationListError) Add(v *ValidationError) {
	e.Validations = append(e.Validations, v)
}

func (e ValidationListError) HasErrors() bool {
	return len(e.Validations) > 0
}

func (e *ValidationListError) WithError(instancePtr, message string, details text.Messages) {
	e.Validations = append(e.Validations, &ValidationError{
		ValidationError: &jsonschema.ValidationError{
			Message:     message,
			InstancePtr: instancePtr,
		},
		Messages: details,
	})
}

func NewValidationListError(errs []*ValidationError) error {
	return errors.WithStack(&ValidationListError{Validations: errs})
}

func NewNoWebAuthnCredentials() error {
	return errors.WithStack(&ValidationError{
		ValidationError: &jsonschema.ValidationError{
			Message:     `account does not exist or has no security key set up`,
			InstancePtr: "#/",
		},
		Messages: new(text.Messages).Add(text.NewErrorValidationSuchNoWebAuthnUser()),
	})
}
