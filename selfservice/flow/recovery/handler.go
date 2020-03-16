package recovery

import (
	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/errorx"
	"github.com/ory/kratos/x"
)

const (
	PublicVerificationInitPath     = "/self-service/browser/flows/verification/:via"
	PublicVerificationCompletePath = "/self-service/browser/flows/verification/:via/complete"
	PublicVerificationRequestPath  = "/self-service/browser/flows/requests/verification"
	PublicVerificationConfirmPath  = "/self-service/browser/flows/verification/:via/confirm/:code"
)

type (
	HandlerProvider interface {
		VerificationHandler() *Handler
	}
	handlerDependencies interface {
		errorx.ManagementProvider
		identity.ManagementProvider
		identity.PrivilegedPoolProvider
		SenderProvider
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
	public.GET(PublicVerificationInitPath, h.init)
	public.GET(PublicVerificationRequestPath, h.publicFetch)
	public.POST(PublicVerificationCompletePath, h.complete)
	public.GET(PublicVerificationConfirmPath, h.verify)
}

func (h *Handler) RegisterAdminRoutes(admin *x.RouterAdmin) {
	admin.GET(PublicVerificationRequestPath, h.adminFetch)
}