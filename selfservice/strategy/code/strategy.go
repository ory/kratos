// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package code

import (
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
		x.TracingProvider

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
		session.PersistenceProvider

		sessiontokenexchange.PersistenceProvider

		continuity.ManagementProvider
	}

	Strategy struct {
		deps strategyDependencies
		dx   *decoderx.HTTP
	}

	codeIdentifier struct {
		Identifier string `json:"identifier"`
	}
)

func NewStrategy(deps any) *Strategy {
	return &Strategy{deps: deps.(strategyDependencies), dx: decoderx.NewHTTP()}
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
			ds, err := s.deps.Config().DefaultIdentityTraitsSchemaURL(r.Context())
			if err != nil {
				return err
			}

			identifierLabel, err := login.GetIdentifierLabelFromSchema(r.Context(), ds.String())
			if err != nil {
				return err
			}

			nodes.Upsert(node.NewInputField("identifier", "", node.DefaultGroup, node.InputAttributeTypeText, node.WithRequiredInputAttribute).WithMetaLabel(identifierLabel))
		} else if f.GetFlowName() == flow.RegistrationFlow {
			ds, err := s.deps.Config().DefaultIdentityTraitsSchemaURL(r.Context())
			if err != nil {
				return err
			}

			// set the traits on the default group so that the ui can render them
			// this prevents having multiple of the same ui fields on the same ui form
			traitNodes, err := container.NodesFromJSONSchema(r.Context(), node.DefaultGroup, ds.String(), "", nil)
			if err != nil {
				return err
			}

			for _, n := range traitNodes {
				nodes.Upsert(n)
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
					// we don't need the user to change the values here
					// for better UX let's make them disabled
					// when there are errors we won't hide the fields
					if len(n.Messages) == 0 {
						if input, ok := n.Attributes.(*node.InputAttributes); ok {
							input.Type = "hidden"
							n.Attributes = input
						}
					}
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

				if n.Group == node.DefaultGroup {
					// we don't need the user to change the values here
					// for better UX let's make them disabled
					// when there are errors we won't hide the fields
					if len(n.Messages) == 0 {
						if input, ok := n.Attributes.(*node.InputAttributes); ok {
							input.Type = "hidden"
							n.Attributes = input
						}
					}
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
func (s *Strategy) NewCodeUINodes(r *http.Request, f flow.Flow, data any) error {
	if err := s.PopulateMethod(r, f); err != nil {
		return err
	}

	prefix := "" // The login flow does not process traits
	if f.GetFlowName() == flow.RegistrationFlow {
		// The registration form does however
		prefix = "traits"
	}

	cont, err := container.NewFromStruct("", node.CodeGroup, data, prefix)
	if err != nil {
		return err
	}

	for _, n := range cont.Nodes {
		// we only set the value and not the whole field because we want to keep types from the initial form generation
		f.GetUI().GetNodes().SetValueAttribute(n.ID(), n.Attributes.GetValue())
	}

	return nil
}

func SetDefaultFlowState(f flow.Flow, resend string) {
	// By Default the flow should be in the 'choose method' state.
	if f.GetState() == "" {
		f.SetState(flow.StateChooseMethod)
	}

	if strings.EqualFold(resend, "code") {
		f.SetState(flow.StateChooseMethod)
	}
}

const CodeLength = 6

func GenerateCode() string {
	return randx.MustString(CodeLength, randx.Numeric)
}
