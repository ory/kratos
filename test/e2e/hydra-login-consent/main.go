package main

import (
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/ory/hydra-client-go/client"
	"github.com/ory/hydra-client-go/client/admin"
	"github.com/ory/hydra-client-go/models"
	"github.com/ory/x/osx"
	"github.com/ory/x/pointerx"
	"github.com/ory/x/urlx"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	router := httprouter.New()

	adminURL := urlx.ParseOrPanic(osx.GetenvDefault("HYDRA_ADMIN_URL", "http://127.0.0.1:4445"))
	hc := client.NewHTTPClientWithConfig(nil, &client.TransportConfig{Schemes: []string{adminURL.Scheme}, Host: adminURL.Host, BasePath: adminURL.Path})

	router.GET("/", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		w.Write([]byte(`ok`))
	})
	router.GET("/login", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		res, err := hc.Admin.GetLoginRequest(admin.NewGetLoginRequestParams().
			WithLoginChallenge(r.URL.Query().Get("login_challenge")))
		check(err)
		if res.Payload.Skip {
			res, err := hc.Admin.AcceptLoginRequest(admin.NewAcceptLoginRequestParams().
				WithLoginChallenge(r.URL.Query().Get("login_challenge")).
				WithBody(&models.AcceptLoginRequest{Remember: true, RememberFor: 3600,
					Subject: pointerx.String(res.Payload.Subject),
				}))
			check(err)
			http.Redirect(w, r, res.Payload.RedirectTo, http.StatusFound)
			return
		}

		challenge := r.URL.Query().Get("login_challenge")
		_, _ = fmt.Fprintf(w, `<html>
<body>
	<form action="/login?login_challenge=%s" method="post">
		<input type="text" name="username" id="username" />
		<input type="text" name="website" id="website" />
		<button type="submit" name="action" value="accept" id="accept">login</button>
		<button type="submit" name="action" value="reject" id="reject">reject</button>
	</form>
</body>
</html>`, challenge)
	})

	router.POST("/login", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		check(r.ParseForm())
		if r.Form.Get("action") == "accept" {
			res, err := hc.Admin.AcceptLoginRequest(admin.NewAcceptLoginRequestParams().
				WithLoginChallenge(r.URL.Query().Get("login_challenge")).
				WithBody(&models.AcceptLoginRequest{Remember: true, RememberFor: 3600,
					Subject: pointerx.String(r.Form.Get("subject")),
				}))
			check(err)
			http.Redirect(w, r, res.Payload.RedirectTo, http.StatusFound)
			return
		}
		res, err := hc.Admin.RejectLoginRequest(admin.NewRejectLoginRequestParams().
			WithLoginChallenge(r.URL.Query().Get("login_challenge")).
			WithBody(&models.RejectRequest{Error: "login rejected request"}))
		check(err)
		http.Redirect(w, r, res.Payload.RedirectTo, http.StatusFound)
	})

	router.GET("/consent", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		res, err := hc.Admin.GetConsentRequest(admin.NewGetConsentRequestParams().
			WithConsentChallenge(r.URL.Query().Get("consent_challenge")))
		check(err)
		if res.Payload.Skip {
			res, err := hc.Admin.AcceptConsentRequest(admin.NewAcceptConsentRequestParams().
				WithConsentChallenge(r.URL.Query().Get("consent_challenge")).
				WithBody(&models.AcceptConsentRequest{Remember: true, RememberFor: 3600,
					GrantScope: res.Payload.RequestedScope,
				}))
			check(err)
			http.Redirect(w, r, res.Payload.RedirectTo, http.StatusFound)
			return
		}

		checkoxes := ""
		for _, s := range res.Payload.RequestedScope {
			checkoxes += fmt.Sprintf(`<li><input type="checkbox" name="scope" value="%s" id="%s"/></li>`, s, s)
		}

		challenge := r.URL.Query().Get("consent_challenge")
		_, _ = fmt.Fprintf(w, `<html>
<body>
	<form action="/consent?consent_challenge=%s" method="post">
		<ul>
		%s
		</ul>
		<button type="submit" name="action" value="accept" id="accept">login</button>
		<button type="submit" name="action" value="reject" id="reject">reject</button>
	</form>
</body>
</html>`, checkoxes, challenge)
	})

	router.POST("/consent", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		_ = r.ParseForm()
		if r.Form.Get("action") == "accept" {
			res, err := hc.Admin.AcceptConsentRequest(admin.NewAcceptConsentRequestParams().
				WithConsentChallenge(r.URL.Query().Get("consent_challenge")).
				WithBody(&models.AcceptConsentRequest{Remember: true, RememberFor: 3600,
					GrantScope: r.Form["scope"]}))
			check(err)
			http.Redirect(w, r, res.Payload.RedirectTo, http.StatusFound)
		}
		res, err := hc.Admin.RejectConsentRequest(admin.NewRejectConsentRequestParams().
			WithConsentChallenge(r.URL.Query().Get("consent_challenge")).
			WithBody(&models.RejectRequest{Error: "consent rejected request"}))
		check(err)
		http.Redirect(w, r, res.Payload.RedirectTo, http.StatusFound)
	})

	server := &http.Server{Addr: ":" + osx.GetenvDefault("PORT", "4446"), Handler: router}
	check(server.ListenAndServe())
}
