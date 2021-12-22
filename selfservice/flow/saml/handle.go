package saml

import (
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/selfservice/errorx"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/x"
	"github.com/ory/x/decoderx"
)

const (
	RouteSamlMetadata = "/self-service/saml/metadata"
	RouteSamlAcs      = "self-service/saml/acs"
)

type (
	handlerDependencies interface {
		x.WriterProvider
		x.CSRFProvider
		session.ManagementProvider
		session.PersistenceProvider
		errorx.ManagementProvider
		config.Provider
	}
	HandlerProvider interface {
		LogoutHandler() *Handler
	}
	Handler struct {
		d  handlerDependencies
		dx *decoderx.HTTP
	}
)

func NewHandler(d handlerDependencies) *Handler {

	return &Handler{
		d:  d,
		dx: decoderx.NewHTTP(),
	}

}

// swagger:model selfServiceLogoutUrl
type selfServiceLogoutUrl struct {
	// SamlMetadataURL is a get endpoint to get the metadata
	//
	// format: uri
	// required: true
	SamlMetadataURL string `json:"saml_metadata_url"`

	// SamlAcsURL is a post endpoint to handle SAML Response
	//
	// format: uri
	// required: true
	SamlAcsURL string `json:"saml_acs_url"`
}

func (h *Handler) RegisterPublicRoutes(router *x.RouterPublic) {
	router.GET(RouteSamlMetadata, h.submitMetadata)
	router.POST(RouteSamlAcs, h.handleResponse)
}

func (h *Handler) submitMetadata(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

func (h *Handler) handleResponse(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	response := r.URL.Query()
	fmt.Println(response)
}
