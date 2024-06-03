// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package code

import (
	"context"
	"net/http"
	"strings"

	"github.com/pkg/errors"
	"github.com/tidwall/gjson"

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

		schema.IdentitySchemaProvider
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
	codeConfig := s.deps.Config().SelfServiceCodeStrategy(r.Context())
	switch f := f.(type) {
	case *login.Flow:
		if f.RequestedAAL == identity.AuthenticatorAssuranceLevel2 {
			if !codeConfig.MFAEnabled {
				// if the flow is requesting AAL2 but MFA is not enabled for the code strategy, we return nil so that
				// other strategies can fulfil the request
				return nil
			}
		} else if !codeConfig.PasswordlessEnabled {
			// if the flow is a normal login flow but passwordless is not enabled for the code strategy,
			// we return nil so that other strategies can fulfil the request
			return nil
		}
	case *registration.Flow:
		// for registration flows, we don't have AAL requirements, so we just check that passwordless is enabled.
		if !codeConfig.PasswordlessEnabled {
			return nil
		}
	}

	if string(f.GetState()) == "" {
		f.SetState(flow.StateChooseMethod)
	}

	f.GetUI().ResetMessages()

	switch f.GetState() {
	case flow.StateChooseMethod:
		if err := s.populateChooseMethodFlow(r, f); err != nil {
			return err
		}
	case flow.StateEmailSent:
		fallthrough
	case flow.StateSMSSent:
		if err := s.populateEmailOrSMSSentFlow(r.Context(), f); err != nil {
			return err
		}
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

func (s *Strategy) populateChooseMethodFlow(r *http.Request, f flow.Flow) error {
	ctx := r.Context()
	var codeMetaLabel *text.Message
	switch f := f.(type) {
	case *recovery.Flow, *verification.Flow:
		f.GetUI().Nodes.Append(
			node.NewInputField("email", nil, node.CodeGroup, node.InputAttributeTypeEmail, node.WithRequiredInputAttribute).
				WithMetaLabel(text.NewInfoNodeInputEmail()),
		)
		f.GetUI().Nodes.Append(
			node.NewInputField("phone", nil, node.CodeGroup, node.InputAttributeTypeTel, node.WithRequiredInputAttribute).
				WithMetaLabel(text.NewInfoNodeInputPhone()),
		)
		codeMetaLabel = text.NewInfoNodeLabelSubmit()
	case *login.Flow:
		ds, err := s.deps.Config().DefaultIdentityTraitsSchemaURL(ctx)
		if err != nil {
			return err
		}
		if f.RequestedAAL == identity.AuthenticatorAssuranceLevel2 {
			via := r.URL.Query().Get("via")
			if via == "" {
				return errors.WithStack(herodot.ErrBadRequest.WithReason("AAL2 login via code requires the `via` query parameter"))
			}

			sess, err := s.deps.SessionManager().FetchFromRequest(r.Context(), r)
			if err != nil {
				return err
			}

			allSchemas, err := s.deps.IdentityTraitsSchemas(ctx)
			if err != nil {
				return err
			}
			iSchema, err := allSchemas.GetByID(sess.Identity.SchemaID)
			if err != nil {
				return err
			}
			identifierLabel, err := login.GetIdentifierLabelFromSchemaWithField(ctx, iSchema.RawURL, via)
			if err != nil {
				return err
			}

			value := gjson.GetBytes(sess.Identity.Traits, via).String()
			if value == "" {
				return errors.WithStack(herodot.ErrBadRequest.WithReasonf("No value found for trait %s in the current identity", via))
			}

			codeMetaLabel = text.NewInfoSelfServiceLoginCodeMFA()
			idNode := node.NewInputField("identifier", "", node.DefaultGroup, node.InputAttributeTypeText, node.WithRequiredInputAttribute).WithMetaLabel(identifierLabel)
			idNode.Messages.Add(text.NewInfoSelfServiceLoginCodeMFAHint(MaskAddress(value)))
			f.GetUI().Nodes.Upsert(idNode)
		} else {
			codeMetaLabel = text.NewInfoSelfServiceLoginCode()
			identifierLabel, err := login.GetIdentifierLabelFromSchema(ctx, ds.String())
			if err != nil {
				return err
			}

			f.GetUI().Nodes.Upsert(node.NewInputField("identifier", "", node.DefaultGroup, node.InputAttributeTypeText, node.WithRequiredInputAttribute).WithMetaLabel(identifierLabel))
		}
	case *registration.Flow:
		codeMetaLabel = text.NewInfoSelfServiceRegistrationRegisterCode()
		ds, err := s.deps.Config().DefaultIdentityTraitsSchemaURL(ctx)
		if err != nil {
			return err
		}

		// set the traits on the default group so that the ui can render them
		// this prevents having multiple of the same ui fields on the same ui form
		traitNodes, err := container.NodesFromJSONSchema(ctx, node.DefaultGroup, ds.String(), "", nil)
		if err != nil {
			return err
		}

		for _, n := range traitNodes {
			f.GetUI().Nodes.Upsert(n)
		}
	}

	methodButton := node.NewInputField("method", s.ID(), node.CodeGroup, node.InputAttributeTypeSubmit).
		WithMetaLabel(codeMetaLabel)

	f.GetUI().Nodes.Append(methodButton)

	return nil
}

func (s *Strategy) populateEmailOrSMSSentFlow(ctx context.Context, f flow.Flow) error {
	// fresh ui node group
	freshNodes := node.Nodes{}
	var route string
	var codeMetaLabel *text.Message
	var message *text.Message

	var resendNode *node.Node
	var backNode *node.Node

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

		// preserve the login identifier that was submitted
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

		// Insert a back button if we have a two-step registration screen, so that the
		// user can navigate back to the credential selection screen.
		if s.deps.Config().SelfServiceFlowRegistrationTwoSteps(ctx) {
			backNode = node.NewInputField(
				"screen",
				"credential-selection",
				node.ProfileGroup,
				node.InputAttributeTypeSubmit,
			).WithMetaLabel(text.NewInfoRegistrationBack())
		}

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

	if backNode != nil {
		freshNodes.Append(backNode)
	}

	f.GetUI().Nodes = freshNodes

	f.GetUI().Method = "POST"
	f.GetUI().Action = flow.AppendFlowTo(urlx.AppendPaths(s.deps.Config().SelfPublicURL(ctx), route), f.GetID()).String()

	f.GetUI().Messages.Set(message)
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

// MaskAddress masks an address by replacing the middle part with asterisks.
//
// If the address contains an @, the part before the @ is masked by taking the first 2 characters and adding 4 *
// (if the part before the @ is less than 2 characters the full value is used).
// Otherwise, the first 3 characters and last two characters are taken and 4 * are added in between.
//
// Examples:
// - foo@bar -> fo****@bar
// - foobar -> fo****ar
// - f@bar -> f@bar
// - fo@bar -> fo****@bar
// - +12345678910 -> +12****10
func MaskAddress(input string) string {
	if strings.Contains(input, "@") {
		pre, post, found := strings.Cut(input, "@")
		if !found || len(pre) < 2 {
			return input
		}
		return pre[:2] + strings.Repeat("*", 4) + "@" + post
	}
	if len(input) < 6 {
		return input
	}
	return input[:3] + strings.Repeat("*", 4) + input[len(input)-2:]
}
