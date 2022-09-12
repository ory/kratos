package courier

import (
	"net/http"

	"github.com/julienschmidt/httprouter"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/x"
	"github.com/ory/x/urlx"
)

const AdminRouteCourier = "/courier"
const AdminRouteMessages = AdminRouteCourier + "/messages"

type (
	handlerDependencies interface {
		x.WriterProvider
		x.LoggingProvider
		x.CSRFProvider
		PersistenceProvider
		config.Provider
	}
	Handler struct {
		r handlerDependencies
	}
	HandlerProvider interface {
		CourierHandler() *Handler
	}
)

func NewHandler(r handlerDependencies) *Handler {
	return &Handler{r: r}
}

func (h *Handler) RegisterPublicRoutes(public *x.RouterPublic) {
	h.r.CSRFHandler().IgnoreGlobs(x.AdminPrefix+AdminRouteMessages, AdminRouteMessages)
	public.GET(x.AdminPrefix+AdminRouteMessages, x.RedirectToAdminRoute(h.r))
}

func (h *Handler) RegisterAdminRoutes(admin *x.RouterAdmin) {
	admin.GET(AdminRouteMessages, h.adminListCourierMessages)
}

// A list of messages.
// swagger:model courierMessageList
// nolint:deadcode,unused
type courierMessageList []Message

// nolint:deadcode,unused
// swagger:parameters adminListCourierMessages
type MessagesFilter struct {
	x.PaginationParams
	// Status filters out messages based on status.
	// If no value is provided, it doesn't take effect on filter.
	//
	// required: false
	// in: query
	Status *MessageStatus `json:"status"`
	// Recipient filters out messages based on recipient.
	// If no value is provided, it doesn't take effect on filter.
	//
	// required: false
	// in: query
	Recipient string `json:"recipient"`
}

// swagger:route GET /admin/courier/messages v0alpha2 adminListCourierMessages
//
// # List Messages
//
// Lists all messages by given status and recipient.
//
//	Produces:
//	- application/json
//
//	Schemes: http, https
//
//	Responses:
//	  200: courierMessageList
//	  400: jsonError
//	  500: jsonError
func (h *Handler) adminListCourierMessages(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	filter, err := parseMessagesFilter(r)
	if err != nil {
		h.r.Writer().WriteErrorCode(w, r, http.StatusBadRequest, err)
		return
	}

	l, tc, err := h.r.CourierPersister().ListMessages(r.Context(), filter)
	if err != nil {
		h.r.Writer().WriteError(w, r, err)
		return
	}

	if !h.r.Config().IsInsecureDevMode(r.Context()) {
		for i := range l {
			l[i].Body = "<redacted-unless-dev-mode>"
		}
	}

	x.PaginationHeader(w, urlx.AppendPaths(h.r.Config().SelfAdminURL(r.Context()), AdminRouteMessages), int64(tc), filter.Page, filter.PerPage)
	h.r.Writer().Write(w, r, l)
}

func parseMessagesFilter(r *http.Request) (MessagesFilter, error) {
	var status *MessageStatus

	if r.URL.Query().Has("status") {
		ms, err := ToMessageStatus(r.URL.Query().Get("status"))

		if err != nil {
			return MessagesFilter{}, err
		}

		status = &ms
	}

	page, itemsPerPage := x.ParsePagination(r)
	return MessagesFilter{
		PaginationParams: x.PaginationParams{
			Page:    page,
			PerPage: itemsPerPage,
		},
		Status:    status,
		Recipient: r.URL.Query().Get("recipient"),
	}, nil
}
