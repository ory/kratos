// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package courier

import (
	"fmt"
	"net/http"

	"github.com/gofrs/uuid"

	"github.com/ory/herodot"
	"github.com/ory/x/pagination/keysetpagination"
	"github.com/ory/x/pagination/migrationpagination"

	"github.com/julienschmidt/httprouter"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/x"
)

const AdminRouteCourier = "/courier"
const AdminRouteListMessages = AdminRouteCourier + "/messages"
const AdminRouteGetMessage = AdminRouteCourier + "/messages/:msgID"

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
	h.r.CSRFHandler().IgnoreGlobs(x.AdminPrefix+AdminRouteListMessages, AdminRouteListMessages)
	public.GET(x.AdminPrefix+AdminRouteListMessages, x.RedirectToAdminRoute(h.r))
	public.GET(x.AdminPrefix+AdminRouteGetMessage, x.RedirectToAdminRoute(h.r))
}

func (h *Handler) RegisterAdminRoutes(admin *x.RouterAdmin) {
	admin.GET(AdminRouteListMessages, h.listCourierMessages)
	admin.GET(AdminRouteGetMessage, h.getCourierMessage)
}

// Paginated Courier Message List Response
//
// swagger:response listCourierMessages
//
//nolint:deadcode,unused
//lint:ignore U1000 Used to generate Swagger and OpenAPI definitions
type listCourierMessagesResponse struct {
	migrationpagination.ResponseHeaderAnnotation

	// List of identities
	//
	// in:body
	Body []Message
}

// Paginated List Courier Message Parameters
//
// swagger:parameters listCourierMessages
type ListCourierMessagesParameters struct {
	keysetpagination.RequestParameters

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
	filter, paginator, err := parseMessagesFilter(r)
	if err != nil {
		h.r.Writer().WriteErrorCode(w, r, http.StatusBadRequest, err)
		return
	}

	l, tc, nextPage, err := h.r.CourierPersister().ListMessages(r.Context(), filter, paginator)
	if err != nil {
		h.r.Writer().WriteError(w, r, err)
		return
	}

	if !h.r.Config().IsInsecureDevMode(r.Context()) {
		for i := range l {
			l[i].Body = "<redacted-unless-dev-mode>"
		}
	}

	w.Header().Set("X-Total-Count", fmt.Sprint(tc))
	keysetpagination.Header(w, r.URL, nextPage)
	h.r.Writer().Write(w, r, l)
}

func parseMessagesFilter(r *http.Request) (ListCourierMessagesParameters, []keysetpagination.Option, error) {
	var status *MessageStatus

	if r.URL.Query().Has("status") {
		ms, err := ToMessageStatus(r.URL.Query().Get("status"))

		if err != nil {
			return ListCourierMessagesParameters{}, nil, err
		}

		status = &ms
	}

	opts, err := keysetpagination.Parse(r.URL.Query(), keysetpagination.NewMapPageToken)
	if err != nil {
		return ListCourierMessagesParameters{}, nil, err
	}

	return ListCourierMessagesParameters{
		Status:    status,
		Recipient: r.URL.Query().Get("recipient"),
	}, opts, nil
}

// Get Courier Message Parameters
//
// swagger:parameters getCourierMessage
//
//nolint:deadcode,unused
//lint:ignore U1000 Used to generate Swagger and OpenAPI definitions
type getCourierMessage struct {
	// MessageID is the ID of the message.
	//
	// required: true
	// in: path
	MessageID string `json:"id"`
}

// swagger:route GET /admin/courier/messages/{id} courier getCourierMessage
//
// # Get a Message
//
// Gets a specific messages by the given ID.
//
//	Produces:
//	- application/json
//
//	Security:
//		oryAccessToken:
//
//	Schemes: http, https
//
//	Responses:
//		200: message
//		400: errorGeneric
//		default: errorGeneric
func (h *Handler) getCourierMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	msgID, err := uuid.FromString(ps.ByName("msgID"))

	if err != nil {
		h.r.Writer().WriteError(w, r, herodot.ErrBadRequest.WithError(err.Error()).WithDebugf("could not parse parameter {id} as UUID, got %s", ps.ByName("id")))
		return
	}

	message, err := h.r.CourierPersister().FetchMessage(r.Context(), msgID)
	if err != nil {
		h.r.Writer().WriteError(w, r, err)
		return
	}

	if !h.r.Config().IsInsecureDevMode(r.Context()) {
		message.Body = "<redacted-unless-dev-mode>"
	}

	h.r.Writer().Write(w, r, message)
}
