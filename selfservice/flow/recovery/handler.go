package recovery

import (
	"net/http"
	"net/url"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/justinas/nosurf"
	"github.com/pkg/errors"

	"github.com/ory/x/urlx"

	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/errorx"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/x"
)

const (
	PublicRecoveryInitPath    = "/self-service/browser/flows/recovery"
	PublicRecoveryRequestPath = "/self-service/browser/flows/requests/recovery"
)

type (
	HandlerProvider interface {
		RecoveryHandler() *Handler
	}
	handlerDependencies interface {
		errorx.ManagementProvider
		identity.ManagementProvider
		identity.PrivilegedPoolProvider
		session.HandlerProvider
		StrategyProvider
		RequestPersistenceProvider
		x.CSRFTokenGeneratorProvider
		x.WriterProvider
	}
	Handler struct {
		d handlerDependencies
		c configuration.Provider
	}
)

func NewHandler(d handlerDependencies, c configuration.Provider) *Handler {
	return &Handler{c: c, d: d}
}

func (h *Handler) RegisterPublicRoutes(public *x.RouterPublic) {
	redirect := session.RedirectOnAuthenticated(h.c)
	public.GET(PublicRecoveryInitPath, h.d.SessionHandler().IsNotAuthenticated(h.init, redirect))
	public.GET(PublicRecoveryRequestPath, h.publicFetch)
}

func (h *Handler) RegisterAdminRoutes(admin *x.RouterAdmin) {
	admin.GET(PublicRecoveryRequestPath, h.adminFetch)
}

// swagger:route GET /self-service/browser/flows/recovery public initializeSelfServiceRecoveryFlow
//
// Initialize browser-based account recovery flow
//
// This endpoint initializes a browser-based account recovery flow. Once initialized, the browser will be redirected to
// `selfservice.flows.recovery.ui_url` with the request ID set as a query parameter. If a valid user session exists, the request
// is aborted.
//
// > This endpoint is NOT INTENDED for API clients and only works
// with browsers (Chrome, Firefox, ...).
//
// More information can be found at [ORY Kratos Account Recovery Documentation](../self-service/flows/password-reset-account-recovery).
//
//     Schemes: http, https
//
//     Responses:
//       302: emptyResponse
//       500: genericError
func (h *Handler) init(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	req, err := NewRequest(h.c.SelfServiceFlowRecoveryRequestLifespan(), h.d.GenerateCSRFToken(r), r, h.d.RecoveryStrategies())
	if err != nil {
		h.d.SelfServiceErrorManager().Forward(r.Context(), w, r, err)
		return
	}

	if err := h.d.RecoveryRequestPersister().CreateRecoveryRequest(r.Context(), req); err != nil {
		h.d.SelfServiceErrorManager().Forward(r.Context(), w, r, err)
		return
	}

	http.Redirect(w, r,
		urlx.CopyWithQuery(h.c.SelfServiceFlowRecoveryUI(), url.Values{"request": {req.ID.String()}}).String(),
		http.StatusFound,
	)
}

// nolint:deadcode,unused
// swagger:parameters getSelfServiceBrowserRecoveryRequest
type getSelfServiceBrowserRecoveryRequestParameters struct {
	// Request is the Login Request ID
	//
	// The value for this parameter comes from `request` URL Query parameter sent to your
	// application (e.g. `/recover?request=abcde`).
	//
	// required: true
	// in: query
	Request string `json:"request"`
}

// swagger:route GET /self-service/browser/flows/requests/recovery common public admin getSelfServiceBrowserRecoveryRequest
//
// Get the request context of browser-based recovery flows
//
// When accessing this endpoint through ORY Kratos' Public API, ensure that cookies are set as they are required
// for checking the auth session. To prevent scanning attacks, the public endpoint does not return 404 status codes
// but instead 403 or 500.
//
// More information can be found at [ORY Kratos Account Recovery Documentation](../self-service/flows/password-reset-account-recovery).
//
//     Produces:
//     - application/json
//
//     Schemes: http, https
//
//     Responses:
//       200: recoveryRequest
//       403: genericError
//       404: genericError
//       410: genericError
//       500: genericError
func (h *Handler) publicFetch(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	if err := h.fetchRequest(w, r, true); err != nil {
		h.d.Writer().WriteError(w, r, err)
		return
	}
}

func (h *Handler) adminFetch(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	if err := h.fetchRequest(w, r, false); err != nil {
		h.d.Writer().WriteError(w, r, err)
		return
	}
}

func (h *Handler) wrapErrorForbidden(err error, shouldWrap bool) error {
	if shouldWrap {
		return x.ErrInvalidCSRFToken.WithTrace(err).WithDebugf("%s", err)
	}

	return err
}

func (h *Handler) fetchRequest(w http.ResponseWriter, r *http.Request, checkCSRF bool) error {
	rid := x.ParseUUID(r.URL.Query().Get("request"))
	req, err := h.d.RecoveryRequestPersister().GetRecoveryRequest(r.Context(), rid)
	if err != nil {
		return h.wrapErrorForbidden(err, checkCSRF)
	}

	if checkCSRF && !nosurf.VerifyToken(h.d.GenerateCSRFToken(r), req.CSRFToken) {
		return errors.WithStack(x.ErrInvalidCSRFToken)
	}

	if req.ExpiresAt.Before(time.Now().UTC()) {
		return errors.WithStack(x.ErrGone.
			WithReason("The recovery request has expired. Redirect the user to the login endpoint to initialize a new session.").
			WithDetail("redirect_to", urlx.AppendPaths(h.c.SelfPublicURL(), PublicRecoveryInitPath).String()))
	}

	h.d.Writer().Write(w, r, req)
	return nil
}
