// Copyright © 2022 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package courier

import (
	"net/http"

	"github.com/ory/x/pagination/migrationpagination"

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
	admin.GET(AdminRouteMessages, h.listCourierMessages)
}

// Paginated Courier Message List Response
//
// swagger:response listCourierMessages
// nolint:deadcode,unused
type listCourierMessagesResponse struct {
	migrationpagination.ResponseHeaderAnnotation

	// List of identities
	//
	// in:body
	Body []Message
}

// Paginated List Courier Message Parameters
//
// nolint:deadcode,unused
// swagger:parameters listCourierMessages
type ListCourierMessagesParameters struct {
	migrationpagination.RequestParameters

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

// swagger:route GET /admin/courier/messages courier listCourierMessages
//
// # List Messages
//
// Lists all messages by given status and recipient.
//
//	Produces:
//	- application/json
//
//	Security:
//	  oryAccessToken:
//
//	Schemes: http, https
//
//	Responses:
//	  200: listCourierMessages
//	  400: errorGeneric
//	  default: errorGeneric
func (h *Handler) listCourierMessages(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
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

func parseMessagesFilter(r *http.Request) (ListCourierMessagesParameters, error) {
	var status *MessageStatus

	if r.URL.Query().Has("status") {
		ms, err := ToMessageStatus(r.URL.Query().Get("status"))

		if err != nil {
			return ListCourierMessagesParameters{}, err
		}

		status = &ms
	}

	page, itemsPerPage := x.ParsePagination(r)
	return ListCourierMessagesParameters{
		RequestParameters: migrationpagination.RequestParameters{
			Page:    page,
			PerPage: itemsPerPage,
		},
		Status:    status,
		Recipient: r.URL.Query().Get("recipient"),
	}, nil
}
