// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oidc

import (
	"io"
	"net/http"

	"github.com/pkg/errors"

	"github.com/ory/herodot"
	"github.com/ory/jsonschema/v3"
	"github.com/ory/kratos/schema"
	"github.com/ory/kratos/text"
	"github.com/ory/x/logrusx"
)

var (
	ErrScopeMissing = herodot.ErrBadRequest.
			WithError("authentication failed because a required scope was not granted").
			WithReasonf(`Unable to finish because one or more permissions were not granted. Please retry and accept all permissions.`)

	ErrIDTokenMissing = herodot.ErrBadRequest.
				WithError("authentication failed because id_token is missing").
				WithReasonf(`Authentication failed because no id_token was returned. Please accept the "openid" permission and try again.`)

	ErrProviderNoAPISupport = herodot.ErrBadRequest.
				WithError("request failed because oidc provider does not implement API flows").
				WithReasonf(`Request failed because oidc provider does not implement API flows.`)
)

func logUpstreamError(l *logrusx.Logger, resp *http.Response) error {
	if resp.StatusCode == http.StatusOK {
		return nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		l = l.WithError(err)
	}

	l.WithField("response_code", resp.StatusCode).WithField("response_body", string(body)).Error("The upstream OIDC provider returned a non 200 status code.")
	return errors.WithStack(herodot.ErrInternalServerError.WithReasonf("OpenID Connect provider returned a %d status code but 200 is expected.", resp.StatusCode))
}

type ValidationErrorContextOIDCPolicyViolation struct {
	Reason string
}

func (r *ValidationErrorContextOIDCPolicyViolation) AddContext(_, _ string) {}

func (r *ValidationErrorContextOIDCPolicyViolation) FinishInstanceContext() {}

func NewUserNotFoundError() error {
	return errors.WithStack(&schema.ValidationError{
		ValidationError: &jsonschema.ValidationError{
			Message:     `user with the provided credentials not found`,
			InstancePtr: "#/",
			Context:     &ValidationErrorContextOIDCPolicyViolation{},
		},
		Messages: new(text.Messages).Add(text.NewErrorValidationOIDCUserNotFound()),
	})
}
