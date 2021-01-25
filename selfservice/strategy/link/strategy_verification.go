package link

import (
	"net/http"
	"net/url"
	"time"

	"github.com/ory/kratos/ui/node"

	"github.com/ory/herodot"

	"github.com/ory/x/pkgerx"

	"github.com/gofrs/uuid"
	"github.com/julienschmidt/httprouter"
	"github.com/markbates/pkger"
	"github.com/pkg/errors"

	"github.com/ory/x/decoderx"
	"github.com/ory/x/sqlcon"
	"github.com/ory/x/sqlxx"
	"github.com/ory/x/urlx"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/schema"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/verification"
	"github.com/ory/kratos/selfservice/form"
	"github.com/ory/kratos/text"
	"github.com/ory/kratos/x"
)

const (
	RouteVerification = "/self-service/verification/methods/link"
)

func (s *Strategy) VerificationStrategyID() string {
	return verification.StrategyVerificationLinkName
}

func (s *Strategy) RegisterPublicVerificationRoutes(public *x.RouterPublic) {
	public.POST(RouteVerification, s.handleVerification)
	public.GET(RouteVerification, s.handleVerification)
}

func (s *Strategy) RegisterAdminVerificationRoutes(admin *x.RouterAdmin) {
}

func (s *Strategy) PopulateVerificationMethod(r *http.Request, req *verification.Flow) error {
	f := form.NewHTMLForm(req.AppendTo(urlx.AppendPaths(s.d.Config(r.Context()).SelfPublicURL(), RouteVerification)).String())

	f.SetCSRF(s.d.GenerateCSRFToken(r))
	f.GetNodes().Upsert(
		// v0.5: form.Field{Name: "email", Type: "email", Required: true}
		node.NewInputField("email", nil, node.VerificationLinkGroup, node.InputAttributeTypeEmail, node.WithRequiredInputAttribute),
	)

	req.Methods[s.VerificationStrategyID()] = &verification.FlowMethod{
		Method: s.VerificationStrategyID(),
		Config: &verification.FlowMethodConfig{FlowMethodConfigurator: &FlowMethod{HTMLForm: f}},
	}
	return nil
}

func (s *Strategy) decodeVerification(r *http.Request, decodeBody bool) (*completeSelfServiceVerificationFlowWithLinkMethodParameters, error) {
	var body completeSelfServiceVerificationFlowWithLinkMethod

	if decodeBody {
		if err := s.dx.Decode(r, &body,
			decoderx.MustHTTPRawJSONSchemaCompiler(
				pkgerx.MustRead(pkger.Open("github.com/ory/kratos:/selfservice/strategy/link/.schema/email.schema.json")),
			),
			decoderx.HTTPDecoderSetValidatePayloads(false),
			decoderx.HTTPDecoderJSONFollowsFormFormat()); err != nil {
			return nil, err
		}
	}

	q := r.URL.Query()
	return &completeSelfServiceVerificationFlowWithLinkMethodParameters{
		Flow:  q.Get("flow"),
		Token: q.Get("token"),
		Body:  body,
	}, nil
}

// handleVerificationError is a convenience function for handling all types of errors that may occur (e.g. validation error).
func (s *Strategy) handleVerificationError(w http.ResponseWriter, r *http.Request, f *verification.Flow, body *completeSelfServiceVerificationFlowWithLinkMethodParameters, err error) {
	if f != nil {
		config, err := f.MethodToForm(s.VerificationStrategyID())
		if err != nil {
			s.d.VerificationFlowErrorHandler().WriteFlowError(w, r, s.VerificationStrategyID(), f, err)
			return
		}

		config.Reset()
		config.SetCSRF(s.d.GenerateCSRFToken(r))

		config.GetNodes().Upsert(
			// v0.5: form.Field{Name: "email", Type: "email", Required: true, Value: body.Body.Email}
			node.NewInputField("email", body.Body.Email, node.VerificationLinkGroup, node.InputAttributeTypeEmail, node.WithRequiredInputAttribute),
		)
	}

	s.d.VerificationFlowErrorHandler().WriteFlowError(w, r, s.VerificationStrategyID(), f, err)
}

// swagger:parameters completeSelfServiceVerificationFlowWithLinkMethod
type completeSelfServiceVerificationFlowWithLinkMethodParameters struct {
	// in: body
	Body completeSelfServiceVerificationFlowWithLinkMethod

	// Verification Token
	//
	// The verification token which completes the verification request. If the token
	// is invalid (e.g. expired) an error will be shown to the end-user.
	//
	// in: query
	Token string `json:"token"`

	// The Flow ID
	//
	// format: uuid
	// in: query
	Flow string `json:"flow"`
}

func (m *completeSelfServiceVerificationFlowWithLinkMethodParameters) GetFlow() uuid.UUID {
	return x.ParseUUID(m.Flow)
}

type completeSelfServiceVerificationFlowWithLinkMethod struct {
	// Email to Verify
	//
	// Needs to be set when initiating the flow. If the email is a registered
	// verification email, a verification link will be sent. If the email is not known,
	// a email with details on what happened will be sent instead.
	//
	// format: email
	// in: body
	Email string `json:"email"`

	// Sending the anti-csrf token is only required for browser login flows.
	CSRFToken string `form:"csrf_token" json:"csrf_token"`
}

// swagger:route POST /self-service/verification/methods/link public completeSelfServiceVerificationFlowWithLinkMethod
//
// Complete Verification Flow with Link Method
//
// Use this endpoint to complete a verification flow using the link method. This endpoint
// behaves differently for API and browser flows and has several states:
//
// - `choose_method` expects `flow` (in the URL query) and `email` (in the body) to be sent
//   and works with API- and Browser-initiated flows.
//	 - For API clients it either returns a HTTP 200 OK when the form is valid and HTTP 400 OK when the form is invalid
//     and a HTTP 302 Found redirect with a fresh verification flow if the flow was otherwise invalid (e.g. expired).
//	 - For Browser clients it returns a HTTP 302 Found redirect to the Verification UI URL with the Verification Flow ID appended.
// - `sent_email` is the success state after `choose_method` and allows the user to request another verification email. It
//   works for both API and Browser-initiated flows and returns the same responses as the flow in `choose_method` state.
// - `passed_challenge` expects a `token` to be sent in the URL query and given the nature of the flow ("sending a verification link")
//   does not have any API capabilities. The server responds with a HTTP 302 Found redirect either to the Settings UI URL
//   (if the link was valid) and instructs the user to update their password, or a redirect to the Verification UI URL with
//   a new Verification Flow ID which contains an error message that the verification link was invalid.
//
// More information can be found at [ORY Kratos Email and Phone Verification Documentation](https://www.ory.sh/docs/kratos/selfservice/flows/verify-email-account-activation).
//
//     Consumes:
//     - application/json
//     - application/x-www-form-urlencoded
//
//     Produces:
//     - application/json
//
//     Schemes: http, https
//
//     Responses:
//       400: verificationFlow
//       302: emptyResponse
//       500: genericError
func (s *Strategy) handleVerification(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	if !s.d.Config(r.Context()).SelfServiceStrategy(s.VerificationStrategyID()).Enabled {
		s.handleVerificationError(w, r, nil, nil, errors.WithStack(herodot.ErrBadRequest.WithReasonf("Verification using this method is not allowed because it was disabled.")))
		return
	}

	body, err := s.decodeVerification(r, false)
	if err != nil {
		s.handleVerificationError(w, r, nil, body, err)
		return
	}

	if len(body.Token) > 0 {
		s.verificationUseToken(w, r, body)
		return
	}

	f, err := s.d.VerificationFlowPersister().GetVerificationFlow(r.Context(), body.GetFlow())
	if err != nil {
		s.handleVerificationError(w, r, f, body, err)
		return
	}

	if err := f.Valid(); err != nil {
		s.handleVerificationError(w, r, f, body, err)
		return
	}

	switch f.State {
	case verification.StateChooseMethod:
		fallthrough
	case verification.StateEmailSent:
		// Do nothing (continue with execution after this switch statement)
		s.verificationHandleFormSubmission(w, r, f)
		return
	case verification.StatePassedChallenge:
		s.retryVerificationFlowWithMessage(w, r, f.Type, text.NewErrorValidationVerificationRetrySuccess())
		return
	default:
		s.retryVerificationFlowWithMessage(w, r, f.Type, text.NewErrorValidationVerificationStateFailure())
		return
	}
}

func (s *Strategy) verificationHandleFormSubmission(w http.ResponseWriter, r *http.Request, f *verification.Flow) {
	var body = new(completeSelfServiceVerificationFlowWithLinkMethodParameters)
	body, err := s.decodeVerification(r, true)
	if err != nil {
		s.handleVerificationError(w, r, f, body, err)
		return
	}

	if len(body.Body.Email) == 0 {
		s.handleVerificationError(w, r, f, body, schema.NewRequiredError("#/email", "email"))
		return
	}

	if err := flow.VerifyRequest(r, f.Type, s.d.Config(r.Context()).DisableAPIFlowEnforcement(), s.d.GenerateCSRFToken, body.Body.CSRFToken); err != nil {
		s.handleVerificationError(w, r, f, body, err)
		return
	}

	if err := s.d.LinkSender().SendVerificationLink(r.Context(), f, identity.VerifiableAddressTypeEmail, body.Body.Email); err != nil {
		if !errors.Is(err, ErrUnknownAddress) {
			s.handleVerificationError(w, r, f, body, err)
			return
		}
		// Continue execution
	}

	config, err := f.MethodToForm(s.VerificationStrategyID())
	if err != nil {
		s.handleVerificationError(w, r, f, body, err)
		return
	}

	config.Reset()
	config.SetCSRF(s.d.GenerateCSRFToken(r))
	config.GetNodes().Upsert(
		// v0.5: form.Field{Name: "email", Type: "email", Required: true, Value: body.Body.Email}
		node.NewInputField("email", body.Body.Email, node.VerificationLinkGroup, node.InputAttributeTypeEmail, node.WithRequiredInputAttribute),
	)

	f.Active = sqlxx.NullString(s.VerificationStrategyID())
	f.State = verification.StateEmailSent
	f.Messages.Set(text.NewVerificationEmailSent())
	if err := s.d.VerificationFlowPersister().UpdateVerificationFlow(r.Context(), f); err != nil {
		s.handleVerificationError(w, r, f, body, err)
		return
	}

	if f.Type == flow.TypeBrowser {
		http.Redirect(w, r, f.AppendTo(s.d.Config(r.Context()).SelfServiceFlowVerificationUI()).String(), http.StatusFound)
		return
	}

	updatedFlow, err := s.d.VerificationFlowPersister().GetVerificationFlow(r.Context(), f.ID)
	if err != nil {
		s.handleVerificationError(w, r, f, body, err)
		return
	}

	s.d.Writer().Write(w, r, updatedFlow)
}

// nolint:deadcode,unused
// swagger:parameters selfServiceBrowserVerify
type selfServiceBrowserVerifyParameters struct {
	// required: true
	// in: query
	Token string `json:"token"`
}

func (s *Strategy) verificationUseToken(w http.ResponseWriter, r *http.Request, body *completeSelfServiceVerificationFlowWithLinkMethodParameters) {
	token, err := s.d.VerificationTokenPersister().UseVerificationToken(r.Context(), body.Token)
	if err != nil {
		if errors.Is(err, sqlcon.ErrNoRows) {
			s.retryVerificationFlowWithMessage(w, r, flow.TypeBrowser, text.NewErrorValidationVerificationTokenInvalidOrAlreadyUsed())
			return
		}

		s.handleVerificationError(w, r, nil, body, err)
		return
	}

	var f *verification.Flow
	if !token.FlowID.Valid {
		f, err = verification.NewFlow(time.Until(token.ExpiresAt), s.d.GenerateCSRFToken(r), r, s.d.VerificationStrategies(), flow.TypeBrowser)
		if err != nil {
			s.handleVerificationError(w, r, nil, body, err)
			return
		}

		if err := s.d.VerificationFlowPersister().CreateVerificationFlow(r.Context(), f); err != nil {
			s.handleVerificationError(w, r, nil, body, err)
			return
		}
	} else {
		f, err = s.d.VerificationFlowPersister().GetVerificationFlow(r.Context(), token.FlowID.UUID)
		if err != nil {
			s.handleVerificationError(w, r, nil, body, err)
			return
		}
	}

	if err := token.Valid(); err != nil {
		s.handleVerificationError(w, r, f, body, err)
		return
	}

	f.Messages.Clear()
	f.State = verification.StatePassedChallenge
	if err := s.d.VerificationFlowPersister().UpdateVerificationFlow(r.Context(), f); err != nil {
		s.handleVerificationError(w, r, f, body, err)
		return
	}

	address := token.VerifiableAddress
	address.Verified = true
	address.VerifiedAt = sqlxx.NullTime(time.Now().UTC())
	address.Status = identity.VerifiableAddressStatusCompleted
	if err := s.d.PrivilegedIdentityPool().UpdateVerifiableAddress(r.Context(), address); err != nil {
		s.handleVerificationError(w, r, f, body, err)
		return
	}

	http.Redirect(w, r, s.d.Config(r.Context()).SelfServiceFlowVerificationReturnTo(f.
		AppendTo(s.d.Config(r.Context()).SelfServiceFlowVerificationUI())).String(), http.StatusFound)
}

func (s *Strategy) retryVerificationFlowWithMessage(w http.ResponseWriter, r *http.Request, ft flow.Type, message *text.Message) {
	s.d.Logger().WithRequest(r).WithField("message", message).Debug("A verification flow is being retried because a validation error occurred.")

	req, err := verification.NewFlow(s.d.Config(r.Context()).SelfServiceFlowVerificationRequestLifespan(), s.d.GenerateCSRFToken(r), r, s.d.VerificationStrategies(), ft)
	if err != nil {
		s.d.SelfServiceErrorManager().Forward(r.Context(), w, r, err)
		return
	}

	req.Messages.Add(message)
	if err := s.d.VerificationFlowPersister().CreateVerificationFlow(r.Context(), req); err != nil {
		s.d.SelfServiceErrorManager().Forward(r.Context(), w, r, err)
		return
	}

	if ft == flow.TypeBrowser {
		http.Redirect(w, r, req.AppendTo(s.d.Config(r.Context()).SelfServiceFlowVerificationUI()).String(), http.StatusFound)
		return
	}

	http.Redirect(w, r, urlx.CopyWithQuery(urlx.AppendPaths(s.d.Config(r.Context()).SelfPublicURL(),
		verification.RouteGetFlow), url.Values{"id": {req.ID.String()}}).String(), http.StatusFound)
}
