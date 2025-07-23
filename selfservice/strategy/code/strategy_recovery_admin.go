// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package code

import (
	"context"
	"net/http"
	"net/url"
	"time"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/trace"

	"github.com/ory/herodot"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/recovery"
	"github.com/ory/kratos/selfservice/strategy"
	"github.com/ory/kratos/text"
	"github.com/ory/kratos/ui/node"
	"github.com/ory/kratos/x"
	"github.com/ory/kratos/x/events"
	"github.com/ory/kratos/x/redir"
	"github.com/ory/pop/v6"
	"github.com/ory/x/decoderx"
	"github.com/ory/x/sqlcon"
	"github.com/ory/x/urlx"
)

const (
	RouteAdminCreateRecoveryCode = "/recovery/code"
)

func (s *Strategy) RegisterPublicRecoveryRoutes(public *x.RouterPublic) {
	s.deps.CSRFHandler().IgnorePath(RouteAdminCreateRecoveryCode)
	public.POST(RouteAdminCreateRecoveryCode, redir.RedirectToAdminRoute(s.deps))
}

func (s *Strategy) RegisterAdminRecoveryRoutes(admin *x.RouterAdmin) {
	wrappedCreateRecoveryCode := strategy.IsDisabled(s.deps, s.RecoveryStrategyID(), s.createRecoveryCodeForIdentity)
	admin.POST(RouteAdminCreateRecoveryCode, wrappedCreateRecoveryCode)
}

// Create Recovery Code for Identity Parameters
//
// swagger:parameters createRecoveryCodeForIdentity
//
//nolint:deadcode,unused
//lint:ignore U1000 Used to generate Swagger and OpenAPI definitions
type createRecoveryCodeForIdentity struct {
	// in: body
	Body createRecoveryCodeForIdentityBody
}

// Create Recovery Code for Identity Request Body
//
// swagger:model createRecoveryCodeForIdentityBody
type createRecoveryCodeForIdentityBody struct {
	// Identity to Recover
	//
	// The identity's ID you wish to recover.
	//
	// required: true
	IdentityID uuid.UUID `json:"identity_id"`

	// Code Expires In
	//
	// The recovery code will expire after that amount of time has passed. Defaults to the configuration value of
	// `selfservice.methods.code.config.lifespan`.
	//
	//
	// pattern: ^([0-9]+(ns|us|ms|s|m|h))*$
	// example:
	//	- 1h
	//	- 1m
	//	- 1s
	ExpiresIn string `json:"expires_in"`

	// Flow Type
	//
	// The flow type for the recovery flow. Defaults to browser.
	//
	// required: false
	FlowType *flow.Type `json:"flow_type"`
}

// Recovery Code for Identity
//
// Used when an administrator creates a recovery code for an identity.
//
// swagger:model recoveryCodeForIdentity
//
//nolint:deadcode,unused
//lint:ignore U1000 Used to generate Swagger and OpenAPI definitions
type recoveryCodeForIdentity struct {
	// RecoveryLink with flow
	//
	// This link opens the recovery UI with an empty `code` field.
	//
	// required: true
	// format: uri
	RecoveryLink string `json:"recovery_link"`

	// RecoveryCode is the code that can be used to recover the account
	//
	// required: true
	RecoveryCode string `json:"recovery_code"`

	// Expires At is the timestamp of when the recovery flow expires
	//
	// The timestamp when the recovery code expires.
	ExpiresAt time.Time `json:"expires_at"`
}

// swagger:route POST /admin/recovery/code identity createRecoveryCodeForIdentity
//
// # Create a Recovery Code
//
// This endpoint creates a recovery code which should be given to the user in order for them to recover
// (or activate) their account.
//
//	Consumes:
//	- application/json
//
//	Produces:
//	- application/json
//
//	Schemes: http, https
//
//	Security:
//	  oryAccessToken:
//
//	Responses:
//	  201: recoveryCodeForIdentity
//	  400: errorGeneric
//	  404: errorGeneric
//	  default: errorGeneric
func (s *Strategy) createRecoveryCodeForIdentity(w http.ResponseWriter, r *http.Request) {
	var p createRecoveryCodeForIdentityBody
	if err := s.dx.Decode(r, &p, decoderx.HTTPJSONDecoder()); err != nil {
		s.deps.Writer().WriteError(w, r, err)
		return
	}

	ctx := r.Context()
	config := s.deps.Config()

	expiresIn := config.SelfServiceCodeMethodLifespan(ctx)
	if len(p.ExpiresIn) > 0 {
		// If an expiration of the code was supplied use it instead of the default duration
		var err error
		expiresIn, err = time.ParseDuration(p.ExpiresIn)
		if err != nil {
			s.deps.Writer().WriteError(w, r, errors.WithStack(herodot.
				ErrBadRequest.
				WithReasonf(`Unable to parse "expires_in" whose format should match "[0-9]+(ns|us|ms|s|m|h)" but did not: %s`, p.ExpiresIn)))
			return
		}
	}

	if expiresIn <= 0 {
		s.deps.Writer().WriteError(w, r, errors.WithStack(herodot.ErrBadRequest.WithReasonf(`Value from "expires_in" must result to a future time: %s`, expiresIn)))
		return
	}

	flowType := flow.TypeBrowser
	if p.FlowType != nil {
		flowType = *p.FlowType
	}
	if !flowType.Valid() {
		s.deps.Writer().WriteError(w, r, errors.WithStack(herodot.ErrBadRequest.WithReasonf(`Value from "flow_type" is not valid: %q`, flowType)))
		return
	}

	recoveryFlow, err := recovery.NewFlow(config, expiresIn, s.deps.GenerateCSRFToken(r), r, s, flowType)
	if err != nil {
		s.deps.Writer().WriteError(w, r, err)
		return
	}
	recoveryFlow.DangerousSkipCSRFCheck = true
	recoveryFlow.State = flow.StateEmailSent
	recoveryFlow.UI.Nodes = node.Nodes{}
	recoveryFlow.UI.Nodes.Append(node.NewInputField("code", nil, node.CodeGroup, node.InputAttributeTypeText, node.WithRequiredInputAttribute, node.WithInputAttributes(func(a *node.InputAttributes) {
		a.Pattern = "[0-9]+"
		a.MaxLength = CodeLength
	})).
		WithMetaLabel(text.NewInfoNodeLabelRecoveryCode()),
	)
	rawCode := GenerateCode()

	recoveryFlow.UI.Nodes.
		Append(node.NewInputField("method", s.RecoveryStrategyID(), node.CodeGroup, node.InputAttributeTypeSubmit).
			WithMetaLabel(text.NewInfoNodeLabelContinue()))

	id, err := s.deps.IdentityPool().GetIdentity(ctx, p.IdentityID, identity.ExpandDefault)
	if notFoundErr := sqlcon.ErrNoRows; errors.As(err, &notFoundErr) {
		s.deps.Writer().WriteError(w, r, notFoundErr.WithReasonf("could not find identity"))
		return
	} else if err != nil {
		s.deps.Writer().WriteError(w, r, err)
		return
	}

	if err := s.deps.TransactionalPersisterProvider().Transaction(ctx, func(ctx context.Context, c *pop.Connection) error {
		if err := s.deps.RecoveryFlowPersister().CreateRecoveryFlow(ctx, recoveryFlow); err != nil {
			return err
		}

		if _, err := s.deps.RecoveryCodePersister().CreateRecoveryCode(ctx, &CreateRecoveryCodeParams{
			RawCode:    rawCode,
			CodeType:   RecoveryCodeTypeAdmin,
			ExpiresIn:  expiresIn,
			FlowID:     recoveryFlow.ID,
			IdentityID: id.ID,
		}); err != nil {
			return err
		}

		return nil
	}); err != nil {
		s.deps.Writer().WriteError(w, r, err)
		return
	}

	trace.SpanFromContext(r.Context()).AddEvent(
		events.NewRecoveryInitiatedByAdmin(ctx, recoveryFlow.ID, id.ID, flowType.String(), "code"),
	)

	s.deps.Audit().
		WithField("identity_id", id.ID).
		WithSensitiveField("recovery_code", rawCode).
		Info("A recovery code has been created.")

	body := &recoveryCodeForIdentity{
		ExpiresAt: recoveryFlow.ExpiresAt.UTC(),
		RecoveryLink: urlx.CopyWithQuery(
			s.deps.Config().SelfServiceFlowRecoveryUI(ctx),
			url.Values{
				"flow": {recoveryFlow.ID.String()},
			}).String(),
		RecoveryCode: rawCode,
	}

	s.deps.Writer().WriteCode(w, r, http.StatusCreated, body, herodot.UnescapedHTML)
}
