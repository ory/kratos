// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package code

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/pkg/errors"

	"github.com/ory/herodot"
	"github.com/ory/kratos/continuity"
	"github.com/ory/kratos/courier"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/schema"
	"github.com/ory/kratos/selfservice/errorx"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/selfservice/flow/recovery"
	"github.com/ory/kratos/selfservice/flow/registration"
	"github.com/ory/kratos/selfservice/flow/settings"
	"github.com/ory/kratos/selfservice/flow/verification"
	"github.com/ory/kratos/selfservice/sessiontokenexchange"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/text"
	"github.com/ory/kratos/ui/container"
	"github.com/ory/kratos/ui/node"
	"github.com/ory/kratos/x"
	"github.com/ory/x/decoderx"
	"github.com/ory/x/randx"
	"github.com/ory/x/urlx"
)

var (
	_ recovery.Strategy      = new(Strategy)
	_ recovery.AdminHandler  = new(Strategy)
	_ recovery.PublicHandler = new(Strategy)
)

var (
	_ verification.Strategy      = new(Strategy)
	_ verification.AdminHandler  = new(Strategy)
	_ verification.PublicHandler = new(Strategy)
)

var (
	_ login.Strategy        = new(Strategy)
	_ registration.Strategy = new(Strategy)
)

type (
	// FlowMethod contains the configuration for this selfservice strategy.
	FlowMethod struct {
		*container.Container
	}

	strategyDependencies interface {
		x.CSRFProvider
		x.CSRFTokenGeneratorProvider
		x.WriterProvider
		x.LoggingProvider

		config.Provider

		session.HandlerProvider
		session.ManagementProvider
		settings.HandlerProvider
		settings.FlowPersistenceProvider

		identity.ValidationProvider
		identity.ManagementProvider
		identity.PoolProvider
		identity.PrivilegedPoolProvider

		courier.Provider

		errorx.ManagementProvider

		recovery.ErrorHandlerProvider
		recovery.FlowPersistenceProvider
		recovery.StrategyProvider
		recovery.HookExecutorProvider

		verification.FlowPersistenceProvider
		verification.StrategyProvider
		verification.HookExecutorProvider

		login.StrategyProvider
		login.FlowPersistenceProvider

		registration.StrategyProvider
		registration.FlowPersistenceProvider

		RecoveryCodePersistenceProvider
		VerificationCodePersistenceProvider
		SenderProvider

		RegistrationCodePersistenceProvider
		LoginCodePersistenceProvider

		schema.IdentityTraitsProvider

		sessiontokenexchange.PersistenceProvider

		continuity.ManagementProvider
	}

	Strategy struct {
		deps strategyDependencies
		dx   *decoderx.HTTP
	}
)

func NewStrategy(deps strategyDependencies) *Strategy {
	return &Strategy{deps: deps, dx: decoderx.NewHTTP()}
}

func (s *Strategy) ID() identity.CredentialsType {
	return identity.CredentialsTypeCodeAuth
}

func (s *Strategy) NodeGroup() node.UiNodeGroup {
	return node.CodeGroup
}

func (s *Strategy) PopulateMethod(r *http.Request, f flow.Flow) error {
	if string(f.GetState()) == "" {
		f.SetState(flow.StateChooseMethod)
	}

	f.GetUI().ResetMessages()

	nodes := f.GetUI().Nodes

	switch f.GetState() {
	case flow.StateChooseMethod:

		if f.GetFlowName() == flow.VerificationFlow || f.GetFlowName() == flow.RecoveryFlow {
			nodes.Append(
				node.NewInputField("email", nil, node.CodeGroup, node.InputAttributeTypeEmail, node.WithRequiredInputAttribute).
					WithMetaLabel(text.NewInfoNodeInputEmail()),
			)
		} else if f.GetFlowName() == flow.LoginFlow {
			// we use the identifier label here since we don't know what
			// type of field the identifier is
			nodes.Upsert(
				node.NewInputField("identifier", nil, node.DefaultGroup, node.InputAttributeTypeText, node.WithRequiredInputAttribute).
					WithMetaLabel(text.NewInfoNodeLabelID()),
			)
		} else if f.GetFlowName() == flow.RegistrationFlow {
			ds, err := s.deps.Config().DefaultIdentityTraitsSchemaURL(r.Context())
			if err != nil {
				return err
			}

			// set the traits on the default group so that the ui can render them
			// this prevents having multiple of the same ui fields on the same ui form
			traitNodes, err := container.NodesFromJSONSchema(r.Context(), node.CodeGroup, ds.String(), "", nil)
			if err != nil {
				return err
			}

			for _, n := range traitNodes {
				nodes.Append(n)
			}
		}

		var codeMetaLabel *text.Message

		switch f.GetFlowName() {
		case flow.VerificationFlow, flow.RecoveryFlow:
			codeMetaLabel = text.NewInfoNodeLabelSubmit()
		case flow.LoginFlow:
			codeMetaLabel = text.NewInfoSelfServiceLoginCode()
		case flow.RegistrationFlow:
			codeMetaLabel = text.NewInfoSelfServiceRegistrationRegisterCode()
		}

		methodButton := node.NewInputField("method", s.ID(), node.CodeGroup, node.InputAttributeTypeSubmit).
			WithMetaLabel(codeMetaLabel)

		nodes.Append(methodButton)

		f.GetUI().Nodes = nodes

	case flow.StateEmailSent:
		// fresh ui node group
		freshNodes := node.Nodes{}
		var route string
		var codeMetaLabel *text.Message
		var message *text.Message

		var resendNode *node.Node

		switch f.GetFlowName() {
		case flow.RecoveryFlow:
			route = recovery.RouteSubmitFlow
			codeMetaLabel = text.NewInfoNodeLabelRecoveryCode()
			message = text.NewRecoveryEmailWithCodeSent()

			resendNode = node.NewInputField("email", nil, node.CodeGroup, node.InputAttributeTypeEmail, node.WithRequiredInputAttribute).
				WithMetaLabel(text.NewInfoNodeResendOTP())
		case flow.VerificationFlow:
			route = verification.RouteSubmitFlow
			codeMetaLabel = text.NewInfoNodeLabelVerificationCode()
			message = text.NewVerificationEmailWithCodeSent()

		case flow.LoginFlow:
			route = login.RouteSubmitFlow
			codeMetaLabel = text.NewInfoNodeLabelLoginCode()
			message = text.NewLoginEmailWithCodeSent()

			// preserve the login identifier that were submitted
			// so we can retry the code flow with the same data
			for _, n := range f.GetUI().Nodes {
				if n.Group == node.DefaultGroup {
					freshNodes = append(freshNodes, n)
				}
			}

			resendNode = node.NewInputField("resend", "code", node.CodeGroup, node.InputAttributeTypeSubmit).
				WithMetaLabel(text.NewInfoNodeResendOTP())

		case flow.RegistrationFlow:
			route = registration.RouteSubmitFlow
			codeMetaLabel = text.NewInfoNodeLabelRegistrationCode()
			message = text.NewRegistrationEmailWithCodeSent()

			// in the registration flow we need to preserve the trait fields that were submitted
			// so we can retry the code flow with the same data
			for _, n := range f.GetUI().Nodes {
				if t, ok := n.Attributes.(*node.InputAttributes); ok && t.Type == node.InputAttributeTypeSubmit {
					continue
				}

				if n.Group == node.CodeGroup {
					freshNodes = append(freshNodes, n)
				}
			}

			resendNode = node.NewInputField("resend", "code", node.CodeGroup, node.InputAttributeTypeSubmit).
				WithMetaLabel(text.NewInfoNodeResendOTP())
		default:
			return errors.WithStack(herodot.ErrBadRequest.WithReason("received an unexpected flow type"))
		}

		// Hidden field Required for the re-send code button
		// !!important!!: this field must be appended before the code submit button since upsert will replace the first node with the same name
		freshNodes.Upsert(
			node.NewInputField("method", s.NodeGroup(), node.CodeGroup, node.InputAttributeTypeHidden),
		)

		// code input field
		freshNodes.Upsert(node.NewInputField("code", nil, node.CodeGroup, node.InputAttributeTypeText, node.WithRequiredInputAttribute).
			WithMetaLabel(codeMetaLabel))

		// code submit button
		freshNodes.
			Append(node.NewInputField("method", s.ID(), node.CodeGroup, node.InputAttributeTypeSubmit).
				WithMetaLabel(text.NewInfoNodeLabelSubmit()))

		if resendNode != nil {
			freshNodes.Append(resendNode)
		}

		f.GetUI().Nodes = freshNodes

		f.GetUI().Method = "POST"
		f.GetUI().Action = flow.AppendFlowTo(urlx.AppendPaths(s.deps.Config().SelfPublicURL(r.Context()), route), f.GetID()).String()

		// Set the request's CSRF token
		if f.GetType() == flow.TypeBrowser {
			f.GetUI().SetCSRF(s.deps.GenerateCSRFToken(r))
		}

		f.GetUI().Messages.Set(message)

	case flow.StatePassedChallenge:
		fallthrough
	default:
		return errors.WithStack(herodot.ErrBadRequest.WithReason("received an unexpected flow state"))
	}

	// no matter the flow type or state we need to set the CSRF token
	if f.GetType() == flow.TypeBrowser {
		f.GetUI().SetCSRF(s.deps.GenerateCSRFToken(r))
	}
	return nil
}

// NewCodeUINodes creates a fresh UI for the code flow.
// this is used with the `recovery`, `verification`, `registration` and `login` flows.
func (s *Strategy) NewCodeUINodes(r *http.Request, f flow.Flow, data json.RawMessage) error {
	if err := s.PopulateMethod(r, f); err != nil {
		return err
	}

	// on Registration flow we need to populate the form with the values from the initial form generation
	if f.GetFlowName() == flow.RegistrationFlow {
		for _, n := range container.NewFromJSON("", node.CodeGroup, data, "traits").Nodes {
			// we only set the value and not the whole field because we want to keep types from the initial form generation
			f.GetUI().GetNodes().SetValueAttribute(n.ID(), n.Attributes.GetValue())
		}
	} else if f.GetFlowName() == flow.LoginFlow {
		// on Login flow we need to populate the form with the values from the initial form generation
		for _, n := range container.NewFromJSON("", node.DefaultGroup, data, "").Nodes {
			f.GetUI().GetNodes().SetValueAttribute(n.ID(), n.Attributes.GetValue())
		}
	}

	return nil
}

type (
	FlowType interface {
		*login.Flow | *registration.Flow | *recovery.Flow | *verification.Flow
		flow.Flow
	}
	FlowPayload interface {
		*updateLoginFlowWithCodeMethod | *updateRegistrationFlowWithCodeMethod | *updateRecoveryFlowWithCodeMethod | *updateVerificationFlowWithCodeMethod
	}
	CreateCodeState[T FlowType, P FlowPayload]           func(context.Context, T, *Strategy, P) error
	ValidateCodeState[T FlowType, P FlowPayload]         func(context.Context, T, *Strategy, P) error
	AlreadyValidatedCodeState[T FlowType, P FlowPayload] func(context.Context, T, *Strategy, P) error
	CodeStateManager[T FlowType, P FlowPayload]          struct {
		f                         T
		payload                   P
		strategy                  *Strategy
		createCodeState           CreateCodeState[T, P]
		verifyCodeState           ValidateCodeState[T, P]
		alreadyValidatedCodeState AlreadyValidatedCodeState[T, P]
	}
)

func NewCodeStateManager[T FlowType, P FlowPayload](f T, s *Strategy, payload P) *CodeStateManager[T, P] {
	return &CodeStateManager[T, P]{
		f:        f,
		strategy: s,
		payload:  payload,
	}
}

func (c *CodeStateManager[T, P]) SetCreateCodeHandler(fn CreateCodeState[T, P]) {
	c.createCodeState = fn
}

func (c *CodeStateManager[T, P]) SetCodeVerifyHandler(fn ValidateCodeState[T, P]) {
	c.verifyCodeState = fn
}

func (c *CodeStateManager[T, P]) SetCodeDoneHandler(fn AlreadyValidatedCodeState[T, P]) {
	c.alreadyValidatedCodeState = fn
}

func (c *CodeStateManager[T, P]) validatePayload(ctx context.Context) error {
	switch v := any(c.payload).(type) {
	case *updateLoginFlowWithCodeMethod:
		if len(v.Identifier) == 0 {
			return errors.WithStack(schema.NewRequiredError("#/identifier", "identifier"))
		}
	case *updateRegistrationFlowWithCodeMethod:
		if len(v.Traits) == 0 {
			return errors.WithStack(schema.NewRequiredError("#/traits", "traits"))
		}
	case *updateRecoveryFlowWithCodeMethod:
		if len(v.Email) == 0 {
			return errors.WithStack(schema.NewRequiredError("#/email", "email"))
		}
	case *updateVerificationFlowWithCodeMethod:
		if len(v.Email) == 0 {
			return errors.WithStack(schema.NewRequiredError("#/email", "email"))
		}
	default:
		return errors.WithStack(herodot.ErrBadRequest.WithReason("received unexpected flow payload type"))
	}
	return nil
}

func (c *CodeStateManager[T, P]) getResend() string {
	switch v := any(c.payload).(type) {
	case *updateLoginFlowWithCodeMethod:
		return v.Resend
	case *updateRegistrationFlowWithCodeMethod:
		return v.Resend
	}
	return ""
}

func (c *CodeStateManager[T, P]) getCode() string {
	switch v := any(c.payload).(type) {
	case *updateLoginFlowWithCodeMethod:
		return v.Code
	case *updateRegistrationFlowWithCodeMethod:
		return v.Code
	case *updateRecoveryFlowWithCodeMethod:
		return v.Code
	case *updateVerificationFlowWithCodeMethod:
		return v.Code
	}
	return ""
}

func (c *CodeStateManager[T, P]) Run(ctx context.Context) error {
	// By Default the flow should be in the 'choose method' state.
	if c.f.GetState() == "" {
		c.f.SetState(flow.StateChooseMethod)
	}

	if strings.EqualFold(c.getResend(), "code") {
		c.f.SetState(flow.StateChooseMethod)
	}

	switch c.f.GetState() {
	case flow.StateChooseMethod:
		// we are in the first submission state of the flow

		if err := c.validatePayload(ctx); err != nil {
			return err
		}

		if err := c.createCodeState(ctx, c.f, c.strategy, c.payload); err != nil {
			return err
		}

	case flow.StateEmailSent:
		// we are in the second submission state of the flow
		// we need to check the code and update the identity
		if len(c.getCode()) == 0 {
			return errors.WithStack(schema.NewRequiredError("#/code", "code"))
		}

		if err := c.validatePayload(ctx); err != nil {
			return err
		}

		if err := c.verifyCodeState(ctx, c.f, c.strategy, c.payload); err != nil {
			return err
		}
	case flow.StatePassedChallenge:
		return c.alreadyValidatedCodeState(ctx, c.f, c.strategy, c.payload)
	default:
		return errors.WithStack(errors.New("Unknown flow state"))
	}
	return nil
}

func (s *Strategy) NextFlowState(f flow.Flow) {
	switch f.GetState() {
	case flow.StateChooseMethod:
		f.SetState(flow.StateEmailSent)
	case flow.StateEmailSent:
		f.SetState(flow.StatePassedChallenge)
	case flow.StatePassedChallenge:
		f.SetState(flow.StatePassedChallenge)
	default:
		f.SetState(flow.StateChooseMethod)
	}
}

const CodeLength = 6

func GenerateCode() string {
	return randx.MustString(CodeLength, randx.Numeric)
}
