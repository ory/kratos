package errorx

import (
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"

	"github.com/ory/herodot"
	"github.com/ory/kratos/x"
	"github.com/ory/nosurf"
)

const RouteGet = "/self-service/errors"

var stub500, _ = json.Marshal([]interface{}{herodot.ErrInternalServerError.WithReasonf("This is a stub error.")})

type (
	handlerDependencies interface {
		x.WriterProvider
		PersistenceProvider
	}
	HandlerProvider interface {
		SelfServiceErrorHandler() *Handler
	}
	Handler struct {
		r    handlerDependencies
		csrf x.CSRFToken
	}
)

func NewHandler(
	r handlerDependencies,
) *Handler {
	return &Handler{r: r, csrf: nosurf.Token}
}

func (h *Handler) WithTokenGenerator(f func(r *http.Request) string) {
	h.csrf = f
}

func (h *Handler) RegisterPublicRoutes(public *x.RouterPublic) {
	public.GET(RouteGet, h.publicFetchError)
}

func (h *Handler) RegisterAdminRoutes(public *x.RouterAdmin) {
	public.GET(RouteGet, h.adminFetchError)
}

// User-facing error response
//
// swagger:response errorContainer
// nolint:deadcode,unused
type getSelfServiceErrorResponse struct {
	// in: body
	Body ErrorContainer `json:"error"`
}

// nolint:deadcode,unused
// swagger:parameters getSelfServiceError
type getSelfServiceErrorParameters struct {
	// Error is the container's ID
	//
	// in: query
	// required: true
	Error string `json:"error"`
}

// swagger:route GET /self-service/errors public admin getSelfServiceError
//
// Get User-Facing Self-Service Errors
//
// This endpoint returns the error associated with a user-facing self service errors.
//
// This endpoint supports stub values to help you implement the error UI:
//
// - `?error=stub:500` - returns a stub 500 (Internal Server Error) error.
//
// More information can be found at [ORY Kratos User User Facing Error Documentation](https://www.ory.sh/docs/kratos/self-service/flows/user-facing-errors).
//
//     Produces:
//     - application/json
//
//     Schemes: http, https
//
//     Responses:
//       200: errorContainer
//       403: genericError
//       404: genericError
//       500: genericError
func (h *Handler) publicFetchError(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	if err := h.fetchError(w, r); err != nil {
		h.r.Writer().WriteError(w, r, x.ErrInvalidCSRFToken.WithTrace(err).WithDebugf("%s", err))
		return
	}
}

func (h *Handler) adminFetchError(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	if err := h.fetchError(w, r); err != nil {
		h.r.Writer().WriteError(w, r, err)
		return
	}
}

func (h *Handler) fetchError(w http.ResponseWriter, r *http.Request) error {
	id := r.URL.Query().Get("error")
	switch id {
	case "stub:500":
		h.r.Writer().Write(w, r, &ErrorContainer{ID: x.NewUUID(), Errors: stub500})
		return nil
	}

	es, err := h.r.SelfServiceErrorPersister().Read(r.Context(), x.ParseUUID(id))
	if err != nil {
		return err
	}

	h.r.Writer().Write(w, r, es)
	return nil
}
