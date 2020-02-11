package verify

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/justinas/nosurf"

	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/x"
)

func TestHandler(t *testing.T) {
	_, reg := internal.NewRegistryDefault(t)

	public, admin := func() (*httptest.Server, *httptest.Server) {
		public := x.NewRouterPublic()
		admin := x.NewRouterAdmin()
		reg.RegistrationHandler().RegisterPublicRoutes(public)
		reg.RegistrationHandler().RegisterAdminRoutes(admin)
		reg.RegistrationStrategies().RegisterPublicRoutes(public)
		return httptest.NewServer(nosurf.New(public)), httptest.NewServer(admin)
	}()
	defer public.Close()
	defer admin.Close()

	redirTS := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer redirTS.Close()
}
