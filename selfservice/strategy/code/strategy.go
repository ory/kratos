// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package code

import (
	"context"
	"encoding/json"
	"net/http"
	"sort"
	"strings"

	"github.com/ory/kratos/x/nosurfx"

	"github.com/samber/lo"

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
	_ recovery.Strategy      = (*Strategy)(nil)
	_ recovery.AdminHandler  = (*Strategy)(nil)
	_ recovery.PublicHandler = (*Strategy)(nil)
)

var (
	_ verification.Strategy      = (*Strategy)(nil)
	_ verification.AdminHandler  = (*Strategy)(nil)
	_ verification.PublicHandler = (*Strategy)(nil)
)

var (
	_ login.Strategy                    = (*Strategy)(nil)
	_ registration.Strategy             = (*Strategy)(nil)
	_ identity.ActiveCredentialsCounter = (*Strategy)(nil)
)

type (
	// FlowMethod contains the configuration for this selfservice strategy.
	FlowMethod struct {
		*container.Container
	}

	strategyDependencies interface {
		nosurfx.CSRFProvider
		nosurfx.CSRFTokenGeneratorProvider
		x.WriterProvider
		x.LoggingProvider
		x.TracingProvider
		x.TransactionPersistenceProvider

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

func (s *Strategy) CountActiveFirstFactorCredentials(ctx context.Context, cc map[identity.CredentialsType]identity.Credentials) (int, error) {
	codeConfig := s.deps.Config().SelfServiceCodeStrategy(ctx)
	if codeConfig.PasswordlessEnabled {
		// Login with code for passwordless is enabled
		return 1, nil
	}

	return 0, nil
}

func (s *Strategy) CountActiveMultiFactorCredentials(ctx context.Context, cc map[identity.CredentialsType]identity.Credentials) (int, error) {
	codeConfig := s.deps.Config().SelfServiceCodeStrategy(ctx)
	if !codeConfig.MFAEnabled {
		return 0, nil
	}

	// Get code credentials if they exist
	creds, ok := cc[identity.CredentialsTypeCodeAuth]
	if !ok {
		return 0, nil
	}

	// Check if the credentials config is valid JSON
	if !gjson.Valid(string(creds.Config)) {
		return 0, nil
	}

	// Check for v0 format with address_type field
	if gjson.GetBytes(creds.Config, "address_type").Exists() {
		addressType := gjson.GetBytes(creds.Config, "address_type").String()
		if addressType != "" {
			return 1, nil
		}
		return 0, nil
	}

	var conf identity.CredentialsCode
	if err := json.Unmarshal(creds.Config, &conf); err != nil {
		return 0, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("Unable to unmarshal credentials config: %s", err))
	}

	// If no addresses configured, return 0
	if len(conf.Addresses) == 0 {
		return 0, nil
	}

	// Count valid addresses configured for MFA
	validAddresses := 0
	for _, addr := range conf.Addresses {
		if addr.Address != "" {
			validAddresses++
		}
	}

	return validAddresses, nil
}

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
		if err := s.populateEmailSentFlow(r.Context(), f); err != nil {
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
	switch f := f.(type) {
	case *recovery.Flow, *verification.Flow:
		f.GetUI().Nodes.Append(
			node.NewInputField("email", nil, node.CodeGroup, node.InputAttributeTypeEmail, node.WithRequiredInputAttribute).
				WithMetaLabel(text.NewInfoNodeInputEmail()),
		)
		f.GetUI().Nodes.Append(
			node.NewInputField("method", s.ID(), node.CodeGroup, node.InputAttributeTypeSubmit).
				WithMetaLabel(text.NewInfoNodeLabelContinue()),
		)
	case *login.Flow:
		ds, err := f.IdentitySchema.URL(ctx, s.deps.Config())
		if err != nil {
			return err
		}
		if f.RequestedAAL == identity.AuthenticatorAssuranceLevel2 {
			via := r.URL.Query().Get("via")

			sess, err := s.deps.SessionManager().FetchFromRequest(r.Context(), r)
			if err != nil {
				return err
			}

			// We need to load the identity's credentials.
			if len(sess.Identity.Credentials) == 0 {
				if err := s.deps.PrivilegedIdentityPool().HydrateIdentityAssociations(ctx, sess.Identity, identity.ExpandCredentials); err != nil {
					return err
				}
			}

			// The via parameter lets us hint at the OTP address to use for 2fa.
			if via == "" {
				addresses, found, err := FindCodeAddressCandidates(sess.Identity, s.deps.Config().SelfServiceCodeMethodMissingCredentialFallbackEnabled(ctx))
				if err != nil {
					return err
				} else if !found {
					return nil
				}

				sort.SliceStable(addresses, func(i, j int) bool {
					return addresses[i].To < addresses[j].To && addresses[i].Via < addresses[j].Via
				})

				for _, address := range addresses {
					f.GetUI().Nodes.Append(node.NewInputField("address", address.To, node.CodeGroup, node.InputAttributeTypeSubmit).
						WithMetaLabel(text.NewInfoSelfServiceLoginAAL2CodeAddress(string(address.Via), address.To)))
				}
			} else {
				value := gjson.GetBytes(sess.Identity.Traits, via).String()
				if value == "" {
					return errors.WithStack(herodot.ErrBadRequest.WithReasonf("No value found for trait %s in the current identity.", via))
				}

				// TODO Remove this normalization once the via parameter is deprecated.
				//
				// Here we need to normalize the via parameter to the actual address. This is necessary because otherwise
				// we won't find the address in the list of addresses.
				//
				// Since we don't know if the via parameter is an email address or a phone number, we need to normalize for both.
				value = x.GracefulNormalization(value)

				addresses, found, err := FindCodeAddressCandidates(sess.Identity, s.deps.Config().SelfServiceCodeMethodMissingCredentialFallbackEnabled(ctx))
				if err != nil {
					return err
				} else if !found {
					return nil
				}

				address, found := lo.Find(addresses, func(item Address) bool {
					return item.To == value
				})
				if !found {
					return errors.WithStack(herodot.ErrBadRequest.WithReasonf("You can only reference a trait that matches a verification email address in the via parameter, or a registered credential."))
				}

				f.GetUI().Nodes.Append(node.NewInputField("address", address.To, node.CodeGroup, node.InputAttributeTypeSubmit).
					WithMetaLabel(text.NewInfoSelfServiceLoginAAL2CodeAddress(string(address.Via), address.To)))
			}
		} else {
			identifierLabel, err := login.GetIdentifierLabelFromSchema(ctx, ds.String())
			if err != nil {
				return err
			}

			f.GetUI().Nodes.Upsert(node.NewInputField("identifier", "", node.DefaultGroup, node.InputAttributeTypeText, node.WithRequiredInputAttribute).WithMetaLabel(identifierLabel))
			f.GetUI().Nodes.Append(
				node.NewInputField("method", s.ID(), node.CodeGroup, node.InputAttributeTypeSubmit).WithMetaLabel(text.NewInfoSelfServiceLoginCode()),
			)
		}
	}

	return nil
}

func (s *Strategy) populateEmailSentFlow(ctx context.Context, f flow.Flow) error {
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
		message = text.NewLoginCodeSent()

		// preserve the login identifier that was submitted
		// so we can retry the code flow with the same data
		for _, n := range f.GetUI().Nodes {
			if n.ID() == "identifier" || n.ID() == "address" {
				if input, ok := n.Attributes.(*node.InputAttributes); ok {
					input.Type = "hidden"
					n.Attributes = input
					input.Name = "identifier"
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

		resendNode = nodeRegistrationResendNode()

		// Insert a back button if we have a two-step registration screen, so that the
		// user can navigate back to the credential selection screen.
		if s.deps.Config().SelfServiceFlowRegistrationTwoSteps(ctx) {
			backNode = nodeRegistrationSelectCredentialsNode()
		}
	default:
		return errors.WithStack(herodot.ErrBadRequest.WithReason("received an unexpected flow type"))
	}

	// Hidden field Required for the re-send code button
	// !!important!!: this field must be appended before the code submit button since upsert will replace the first node with the same name
	freshNodes.Upsert(nodeCodeInputFieldHidden())

	// code input field
	freshNodes.Upsert(nodeCodeInputField().WithMetaLabel(codeMetaLabel))

	// code submit button
	freshNodes.Append(nodeContinueButton())

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
